//go:build sqlite_fts5

package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

type pdfResult struct {
	filename string
	content  string
}

func extractText(path string) (string, error) {
	cmd := exec.Command("pdftotext", path, "-")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func findPDFs(folder string) ([]string, error) {
	var pdfs []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".pdf") {
			pdfs = append(pdfs, path)
		}
		return nil
	})
	return pdfs, err
}

func confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + " [y/N]: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	return err
}

func main() {
	folder := flag.String("folder", "./pdfs", "Folder containing PDFs to index")
	dbPath := flag.String("db", "index.db", "SQLite DB file path")
	threads := flag.Int("threads", 8, "Number of concurrent threads")
	flag.Parse()

	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		if !confirm(fmt.Sprintf("Database '%s' not found. Do you want to create it?", *dbPath)) {
			fmt.Println("Exiting.")
			return
		}
	}

	// Ensure parent directory of DB path exists
	if err := os.MkdirAll(filepath.Dir(*dbPath), os.ModePerm); err != nil {
		log.Fatalf("Failed to create DB directory: %v", err)
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS pdfs USING fts5(filename, content)`)
	if err != nil {
		log.Fatalf("Failed to create FTS5 table: %v", err)
	}

	files, err := findPDFs(*folder)
	if err != nil {
		log.Fatalf("Failed to find PDFs: %v", err)
	}

	total := len(files)
	if total == 0 {
		fmt.Println("No PDF files found.")
		return
	}

	fmt.Printf("\n📁 Found %d PDF(s) to index in %s\n", total, *folder)
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription("Indexing"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "█", SaucerPadding: " ", BarStart: "|", BarEnd: "|"}),
	)

	results := make(chan pdfResult, *threads)
	done := make(chan bool)

	// Insert goroutine
	go func() {
		for res := range results {
			_, err := db.Exec("INSERT INTO pdfs (filename, content) VALUES (?, ?)", res.filename, res.content)
			if err != nil {
				log.Printf("❌ Failed to insert %s: %v\n", res.filename, err)
			}
			bar.Add(1)
		}
		done <- true
	}()

	// Worker goroutines
	var wg sync.WaitGroup
	sem := make(chan struct{}, *threads)

	for _, file := range files {
		wg.Add(1)
		sem <- struct{}{}

		go func(f string) {
			defer wg.Done()
			defer func() { <-sem }()

			text, err := extractText(f)
			if err != nil {
				log.Printf("❌ Failed to extract %s: %v\n", filepath.Base(f), err)
				return
			}
			results <- pdfResult{filename: filepath.Base(f), content: text}
		}(file)
	}

	wg.Wait()
	close(results)
	<-done

	fmt.Println("\n✅ Indexing complete.")
}