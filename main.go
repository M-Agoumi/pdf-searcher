package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

func findPDFs(folder string) ([]string, error) {
	var pdfs []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".pdf") {
			pdfs = append(pdfs, path)
		}
		return nil
	})
	return pdfs, err
}

func searchPDF(path string, keywords []string, requireAll bool) (bool, error) {
	cmd := exec.Command("pdftotext", path, "-")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	text := strings.ToLower(out.String())

	if requireAll {
		for _, keyword := range keywords {
			if !strings.Contains(text, strings.ToLower(keyword)) {
				return false, nil // one missing → no match
			}
		}
		return true, nil // all found
	} else {
		for _, keyword := range keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				return true, nil // any match is enough
			}
		}
		return false, nil
	}
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
	// Command-line flags
	folder := flag.String("folder", "./pdfs", "Folder to search for PDFs")
	threads := flag.Int("threads", 8, "Number of concurrent threads (goroutines)")
	allMatch := flag.Bool("all", false, "Require all keywords to appear in a file (AND match)")
	saveFolder := flag.String("save", "", "Folder to save found PDFs")
	flag.Parse()

	keywords := flag.Args()
	if len(keywords) == 0 {
		fmt.Println("❌ Usage: go run main.go --folder <path> --threads <num> <searchTerm1> [searchTerm2] ...")
		os.Exit(1)
	}

	start := time.Now()
	files, err := findPDFs(*folder)
	if err != nil {
		log.Fatalf("Failed to scan folder: %v", err)
	}
	total := len(files)
	if total == 0 {
		fmt.Println("No PDF files found.")
		return
	}

	fmt.Printf("🔍 Searching %d PDF(s) in \"%s\" for keyword(s): %s\n", total, *folder, strings.Join(keywords, ", "))
	fmt.Printf("🚀 Using %d threads\n\n", *threads)

	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription("Progress"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "█", SaucerPadding: " ", BarStart: "|", BarEnd: "|"}),
	)

	var wg sync.WaitGroup
	sem := make(chan struct{}, *threads)
	var mu sync.Mutex
	found := 0

	if *saveFolder != "" {
		if err := os.MkdirAll(*saveFolder, os.ModePerm); err != nil {
			log.Fatalf("Failed to create folder %s: %v", *saveFolder, err)
		}
	}

	for _, file := range files {
		wg.Add(1)
		sem <- struct{}{}

		go func(f string) {
			defer wg.Done()
			defer func() { <-sem }()
			match, err := searchPDF(f, keywords, *allMatch)
			bar.Add(1)
			if err != nil {
				log.Printf("⚠️ Failed to read %s: %v", filepath.Base(f), err)
				return
			}
			if match {
				mu.Lock()
				found++
				fmt.Printf("✅ Found in: %s\n", filepath.Base(f))
				if *saveFolder != "" {
					srcPath := f
					dstPath := filepath.Join(*saveFolder, filepath.Base(f))
					if err := copyFile(srcPath, dstPath); err != nil {
						log.Printf("❌ Failed to copy %s: %v", filepath.Base(f), err)
					}
				}
				mu.Unlock()
			}
		}(file)
	}

	wg.Wait()
	elapsed := time.Since(start).Seconds()
	fmt.Printf("\n✅ Done! Found %d matching file(s).\n", found)
	fmt.Printf("🕒 Total time: %.2f seconds\n", elapsed)
}
