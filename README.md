# 🧾 PDF Keyword Searcher

A fast, concurrent PDF indexing and search tool written in Go. Perfect for scanning large collections of PDFs using keywords — with support for Docker, full-text search (FTS5), and blazing performance.

---

## 🚀 Features
- 🔎 Full-text search powered by SQLite FTS5
- ⚡ Fast, multithreaded PDF indexing using `pdftotext`
- 📦 Docker support for easy setup
- 🛠️ Command-line interface for both indexing and searching

---

## 📁 Project Structure
```
.
├── indexer.go          # Indexes PDF content into SQLite
├── searcher.go         # Searches indexed PDFs by keywords
├── main.go             # Ad-hoc one-time search without database/indexing
├── Dockerfile          # Multi-stage Docker build
├── docker-compose.yml  # Volume-mounted environment
├── enter-container.sh  # Bash script to enter container (Linux/macOS)
├── enter-container.bat # Windows script to enter container
├── db/                 # (Mounted) SQLite database location
├── pdfs/               # (Mounted) Directory with large/full PDF dataset
└── test/               # Small test dataset for quick experimentation
```

---

## 🐳 Docker Setup

### 🔨 Build & Start the Container
```bash
docker compose up --build
```

### 🧭 Enter the Container Shell
- **Linux/macOS**:
  ```bash
  ./enter-container.sh
  ```
- **Windows**:
  ```cmd
  enter-container.bat
  ```

---

## 📌 Usage Inside the Container

### 📥 Index PDFs
```bash
./indexer --folder /app/pdfs --db /app/db/index.db --threads 8
```

### 🔍 Search PDFs (OR match)
```bash
./searcher --db /app/db/index.db keyword1 keyword2
```
✅ This returns PDFs that contain **at least one** of the given keywords.  
It’s the fastest and broadest search mode — good for finding **any relevant match**.

Example:  
Finds PDFs that have either `"invoice"` **or** `"receipt"` somewhere in the content.

---

### 🔒 Search PDFs (AND match)
```bash
./searcher --db /app/db/index.db --all keyword1 keyword2
```
🔐 This returns PDFs that contain **all** the given keywords, but **not necessarily together**.  
Each word can appear anywhere in the document — even on separate pages.

Example:  
Finds PDFs that mention both `"project"` and `"budget"`, even if they’re in different sections.

---

### 🧵 Search PDFs (Exact phrase match)
```bash
./searcher --db /app/db/index.db --exact keyword1 keyword2
```
🧵 This returns PDFs that contain the **exact phrase** as written — same words, same order, side-by-side.

Example:  
Only finds PDFs that have the exact phrase `"project budget"` (not one with `"budget"` in a different paragraph).

---

### 🚫 No Index? Just Search Once
If you don't want to build an index and just want to search your PDF collection once directly:
```bash
./main --folder /app/pdfs agoumi invoice
```

Supports multithreaded scanning and a progress bar. Ideal for one-off searches without setting up a database.

### 🧪 Quick Testing
To avoid scanning your entire dataset each time, you can use the `test/` folder:
```bash
./main --folder /app/test agoumi invoice
```
This is a much smaller set of PDFs to validate functionality before running on the full set.

---

## 📦 Requirements (if building locally)
- Go 1.22+
- `pdftotext` (from `poppler-utils`)
- SQLite compiled with FTS5
- CGO enabled (`CGO_ENABLED=1`)

---

## 🛠 Build Locally (with FTS5)
```bash
CGO_ENABLED=1 go build -tags sqlite_fts5 -o indexer indexer.go
CGO_ENABLED=1 go build -tags sqlite_fts5 -o searcher searcher.go
CGO_ENABLED=1 go build -o main main.go
```

---

## 🙌 Credits
- PDF parsing via [`poppler-utils`](https://poppler.freedesktop.org/)
- FTS5 search powered by SQLite
- Progress bar: [`schollz/progressbar`](https://github.com/schollz/progressbar)
- Go SQLite driver: [`mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3)

---

## 📬 License

### DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE (WTFPL)

> This program is free software. It comes without any warranty, to the extent permitted by applicable law.
> 
> You can do whatever the fuck you want with this software.
> 
> THE AUTHOR IS NOT RESPONSIBLE FOR ANY DAMAGE OR CONSEQUENCES OF USING THIS SOFTWARE.
