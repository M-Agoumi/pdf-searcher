//go:build sqlite_fts5

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

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
	dbPath := flag.String("db", "index.db", "Path to SQLite database")
	all := flag.Bool("all", false, "Require all keywords to match (AND logic)")
	exact := flag.Bool("exact", false, "Require exact phrase match")
	saveFolder := flag.String("save", "", "Folder to save found PDFs")
	flag.Parse()

	keywords := flag.Args()
	if len(keywords) == 0 {
		fmt.Println("❌ Usage: go run searcher.go [--db path] [--all] <keyword1> <keyword2> ...")
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	var query string
	if *exact && *all {
		// Quote multi-word args, keep AND between all
		for i, kw := range keywords {
			if strings.Contains(kw, " ") {
				keywords[i] = fmt.Sprintf("\"%s\"", kw)
			}
		}
		query = strings.Join(keywords, " AND ")
	} else if *exact {
		query = fmt.Sprintf("\"%s\"", strings.Join(keywords, " "))
	} else if *all {
		query = strings.Join(keywords, " AND ")
	} else {
		query = strings.Join(keywords, " OR ")
	}
	
	rows, err := db.Query("SELECT filename FROM pdfs WHERE content MATCH ?", query)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	defer rows.Close()

	fmt.Printf("\n🔍 Search Results for [%s]:\n", query)
	found := false
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err == nil {
			fmt.Println("✅", filename)
			found = true

			if *saveFolder != "" {
				srcPath := filepath.Join("./pdfs", filename)
				dstPath := filepath.Join(*saveFolder, filename)
				if err := copyFile(srcPath, dstPath); err != nil {
					log.Printf("❌ Failed to copy %s: %v", filename, err)
				}
			}
		}
	}
	if !found {
		fmt.Println("❌ No matching files found.")
	}
}
