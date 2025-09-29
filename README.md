# ThreadBound - iMessages Book Generator

A Go-based toolchain to convert iMessages database exports into a formatted book using Pandoc.

## Features

- **SQLite Database Processing**: Extracts messages, contacts, and attachments from iMessages database
- **Conversation Layout**: Formats messages as a conversation with sender identification and timestamps
- **Attachment Support**: Includes images and files in the book (with format conversion)
- **Professional PDF Output**: Uses Pandoc with custom LaTeX templates for high-quality books
- **Custom Page Size**: Optimized for 8.5" × 5.5" book format
- **CLI Interface**: Easy-to-use command line tools

## Requirements

- **Go 1.22+**: For building and running the tool
- **Pandoc**: For converting Markdown to PDF
- **XeLaTeX**: PDF engine for high-quality output
- **System Fonts**: Helvetica and Courier (or similar)

### Installing Dependencies

**macOS:**
```bash
# Install Pandoc
brew install pandoc

# Install LaTeX (MacTeX)
brew install --cask mactex
```

**Linux:**
```bash
# Install Pandoc
sudo apt-get install pandoc

# Install LaTeX
sudo apt-get install texlive-xetex texlive-fonts-recommended
```

## Getting Started

### 1. Build the tool
```bash
cd src
go build -o threadbound ./cmd/threadbound
```

### 2. Set up configuration (optional)
Copy the sample config file and customize it:
```bash
cp src/threadbound.yaml.sample src/threadbound.yaml
# Edit threadbound.yaml with your preferences
```

The config file allows you to set default values for all options:
```yaml
# ThreadBound Configuration File
title: "Our Group Chat"
author: "The Squad"
database_path: "chat.db"
attachments_path: "Attachments"
output_path: "book.md"
template_dir: "templates"
include_images: true
include_previews: true
page_width: "5.5in"
page_height: "8.5in"
```

### 3. Generate a book

**Using the build script (recommended):**
```bash
# Without config file
./build-book.sh "My Group Chat" "Your Name"

# With config file (title and author can still be overridden)
./build-book.sh "My Group Chat" "Your Name" "output_dir" "src/threadbound.yaml"
```

**Using commands directly:**
```bash
# With config file
./src/threadbound generate --config src/threadbound.yaml
./src/threadbound build-pdf --config src/threadbound.yaml

# Without config file (using CLI flags)
./src/threadbound generate --title "My Chat" --author "Me"
./src/threadbound build-pdf --input book.md
```

## Manual Usage

### Step 1: Generate Markdown

```bash
./src/threadbound generate \
    --db chat.db \
    --title "Our Group Chat" \
    --author "Your Name" \
    --output book.md \
    --include-images=true
```

### Step 2: Build PDF

```bash
./src/threadbound build-pdf \
    --input book.md \
    --template-dir templates \
    --page-width 5.5in \
    --page-height 8.5in
```

### Available CLI Options

Run `./src/threadbound [command] --help` to see all available options for each command.

**Global flags:**
- `--config`: Path to YAML config file

**Generate command flags:**
- `--db`: Path to iMessages database (default: "chat.db")
- `--attachments`: Path to attachments directory (default: "Attachments")
- `--output`: Output markdown file (default: "book.md")
- `--title`: Book title (default: "Our Messages")
- `--author`: Book author
- `--include-images`: Include images in output (default: true)
- `--page-width`: Page width (default: "5.5in")
- `--page-height`: Page height (default: "8.5in")

**Build-PDF command flags:**
- `--input`: Input markdown file (default: "book.md")
- `--template-dir`: Template directory (default: "templates")
- `--page-width`: Page width (default: "5.5in")
- `--page-height`: Page height (default: "8.5in")

## Project Structure

```
threadbound/
├── src/
│   ├── cmd/threadbound/main.go     # CLI entry point
│   ├── internal/
│   │   ├── database/sqlite.go      # Database operations
│   │   ├── models/message.go       # Data structures
│   │   ├── markdown/generator.go   # Markdown generation
│   │   ├── attachments/processor.go # File processing
│   │   └── book/                   # Book building logic
│   ├── threadbound                 # Compiled binary
│   └── threadbound.yaml.sample     # Sample config file
├── templates/
│   └── book.tex                    # LaTeX template
├── build-book.sh                   # Complete pipeline script
└── README.md
```

## Configuration Options

### Generate Command

- `--db`: Path to iMessages database (default: `chat.db`)
- `--attachments`: Path to attachments directory (default: `Attachments`)
- `--title`: Book title
- `--author`: Book author
- `--output`: Output markdown file (default: `book.md`)
- `--include-images`: Include images in output (default: `true`)
- `--include-previews`: Generate link previews (default: `false`)

### Build PDF Command

- `--input`: Input markdown file (default: `book.md`)
- `--template-dir`: Template directory (default: `templates`)
- `--page-width`: Page width (default: `5.5in`)
- `--page-height`: Page height (default: `8.5in`)

## Customization

### LaTeX Template

Edit `templates/book.tex` to customize:
- Page layout and margins
- Font choices
- Message bubble styling
- Colors and spacing

### Message Processing

Modify `internal/markdown/generator.go` to change:
- Date formatting
- Sender name display
- Message bubble layout
- Attachment handling

## File Format Support

### Supported Image Formats
- JPEG, PNG, GIF, BMP, TIFF
- HEIC (with limitations in PDF output)

### Supported Attachment Types
- Images (embedded in book)
- Other files (listed as attachments)

## Troubleshooting

### PDF Generation Issues

1. **Font not found**: Update `templates/book.tex` with available system fonts
2. **HEIC images**: Consider converting to JPEG for better PDF compatibility
3. **Large files**: Use `--include-images=false` for text-only version

### Database Issues

1. **Permission denied**: Ensure read access to database file
2. **Empty results**: Check database path and table structure
3. **Attachments not found**: Verify attachments directory path

## Output

The generated book includes:
- **Title page** with customizable title and author
- **Copyright page** with legal information
- **Table of contents** with chapter navigation
- **Messages** formatted as conversations with:
  - Date headers for each day
  - Sender identification
  - Timestamps
  - Embedded images
  - Attachment references

## Examples

### Basic Usage
```bash
./build-book.sh "Family Chat 2025"
```

### Advanced Usage
```bash
./imessages-book generate \
    --db /path/to/chat.db \
    --title "Adventure Planning" \
    --author "The Crew" \
    --include-images=true \
    --output adventure-chat.md

./imessages-book build-pdf \
    --input adventure-chat.md \
    --page-width 6in \
    --page-height 9in
```

## License

This project is for personal use with your own iMessage data. Respect privacy and obtain consent before processing others' messages.