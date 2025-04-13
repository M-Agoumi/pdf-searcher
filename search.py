import os
import time
import PyPDF2
from concurrent.futures import ThreadPoolExecutor, as_completed

def process_pdf(file_path, search_term):
    try:
        with open(file_path, 'rb') as f:
            reader = PyPDF2.PdfReader(f)
            for i, page in enumerate(reader.pages):
                text = page.extract_text()
                if text and search_term.lower() in text.lower():
                    return f"✅ Found in: {os.path.basename(file_path)} (Page {i+1})"
    except Exception as e:
        return f"⚠️ Error reading {os.path.basename(file_path)}: {e}"
    return None

def search_pdfs(folder_path, search_term, max_workers=8):
    start_time = time.perf_counter()

    pdf_paths = []
    for root, _, files in os.walk(folder_path):
        for file in files:
            if file.lower().endswith('.pdf'):
                pdf_paths.append(os.path.join(root, file))

    print(f"🔍 Searching {len(pdf_paths)} PDF(s) for the word: \"{search_term}\"...\n")

    found_count = 0
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = {executor.submit(process_pdf, path, search_term): path for path in pdf_paths}
        for future in as_completed(futures):
            result = future.result()
            if result:
                found_count += 1
                print(result)

    end_time = time.perf_counter()
    elapsed = end_time - start_time

    print(f"\n✅ Done! Found {found_count} matching file(s).")
    print(f"🕒 Total time: {elapsed:.2f} seconds")

# 🔧 Usage
search_folder   = "./pdfs"       # replace with your folder path
word_to_find    = "XXXXX"    # replace with your search term

search_pdfs(search_folder, word_to_find)
