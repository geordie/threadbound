#!/bin/bash

# Complete iMessages Book Generation Pipeline
# Usage: ./build-book.sh [title] [author] [output_dir] [config_file]

set -e

TITLE="${1:-Our Group Chat}"
AUTHOR="${2:-}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_DIR="${3:-output_$TIMESTAMP}"
CONFIG_FILE="${4:-}"

echo "📚 iMessages Book Generation Pipeline"
echo "Title: ${TITLE}"
echo "Author: ${AUTHOR}"
echo "Output Directory: ${OUTPUT_DIR}"
if [ -n "${CONFIG_FILE}" ]; then
    echo "Config File: ${CONFIG_FILE}"
fi
echo

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Step 1: Generate Markdown
echo "🔄 Step 1: Generating markdown from database..."
if [ -n "${CONFIG_FILE}" ]; then
    ./src/threadbound generate \
        --config "${CONFIG_FILE}" \
        --title "${TITLE}" \
        --author "${AUTHOR}" \
        --output "${OUTPUT_DIR}/book.md"
else
    ./src/threadbound generate \
        --title "${TITLE}" \
        --author "${AUTHOR}" \
        --output "${OUTPUT_DIR}/book.md"
fi

if [ $? -ne 0 ]; then
    echo "❌ Failed to generate markdown"
    exit 1
fi

echo "✅ Markdown generated: ${OUTPUT_DIR}/book.md"

# Step 2: Generate PDF
echo "🔄 Step 2: Generating PDF..."
if [ -n "${CONFIG_FILE}" ]; then
    ./src/threadbound build-pdf \
        --config "${CONFIG_FILE}" \
        --input "${OUTPUT_DIR}/book.md"
else
    ./src/threadbound build-pdf \
        --input "${OUTPUT_DIR}/book.md" \
        --template-dir templates
fi

if [ $? -ne 0 ]; then
    echo "⚠️  PDF generation had issues, but may have succeeded"
    echo "   Check the output file manually"
fi

# Check if PDF was created
PDF_FILE="${OUTPUT_DIR}/book.pdf"
if [ -f "${PDF_FILE}" ]; then
    echo "✅ PDF generated: ${PDF_FILE}"

    # Get file size
    SIZE=$(ls -lh "${PDF_FILE}" | awk '{print $5}')
    echo "📄 File size: ${SIZE}"

    # Suggest preview command
    echo "👀 To preview: open ${PDF_FILE}"
else
    echo "❌ PDF was not created"
    exit 1
fi

echo
echo "🎉 Book generation complete!"
echo "📁 Output directory: ${OUTPUT_DIR}"
echo "📖 Book file: ${PDF_FILE}"
echo "📝 Markdown source: ${OUTPUT_DIR}/book.md"