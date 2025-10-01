#!/bin/bash

# Complete iMessages Book Generation Pipeline
# Usage: ./build-book.sh [options] [title] [author] [output_dir]
# Options:
#   --config FILE    Use config file
#
# Or legacy positional: ./build-book.sh [title] [author] [output_dir] [config_file]

set -e

# Parse arguments
TITLE=""
AUTHOR=""
OUTPUT_DIR=""
CONFIG_FILE=""

# First pass - look for flags
while [[ $# -gt 0 ]]; do
    case $1 in
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        *)
            # Not a flag, save as positional argument
            if [ -z "$TITLE" ]; then
                TITLE="$1"
            elif [ -z "$AUTHOR" ]; then
                AUTHOR="$1"
            elif [ -z "$OUTPUT_DIR" ]; then
                OUTPUT_DIR="$1"
            elif [ -z "$CONFIG_FILE" ]; then
                # Legacy: 4th positional arg is config file
                CONFIG_FILE="$1"
            fi
            shift
            ;;
    esac
done

# Set defaults
TITLE="${TITLE:-Our Group Chat}"
YEAR=$(date +"%Y")
MONTH=$(date +"%m")
DAY=$(date +"%d")
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_DIR="${OUTPUT_DIR:-output/$YEAR/$MONTH/$DAY/$TIMESTAMP}"

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

# Step 1: Generate TeX
echo "🔄 Step 1: Generating TeX from database..."
if [ -n "${CONFIG_FILE}" ]; then
    # When using config file, only pass title/author if explicitly provided (not defaults)
    CMD="./threadbound generate --config \"${CONFIG_FILE}\" --output \"${OUTPUT_DIR}/book.tex\""

    # Only override config file values if they were explicitly provided as arguments
    if [ "$TITLE" != "Our Group Chat" ] && [ -n "$TITLE" ]; then
        CMD="$CMD --title \"${TITLE}\""
    fi
    if [ -n "${AUTHOR}" ]; then
        CMD="$CMD --author \"${AUTHOR}\""
    fi

    eval $CMD
else
    ./threadbound generate \
        --title "${TITLE}" \
        --author "${AUTHOR}" \
        --output "${OUTPUT_DIR}/book.tex"
fi

if [ $? -ne 0 ]; then
    echo "❌ Failed to generate TeX"
    exit 1
fi

echo "✅ TeX generated: ${OUTPUT_DIR}/book.tex"

# Step 2: Generate PDF
echo "🔄 Step 2: Generating PDF..."
if [ -n "${CONFIG_FILE}" ]; then
    ./threadbound build-pdf \
        --config "${CONFIG_FILE}" \
        --input "${OUTPUT_DIR}/book.tex"
else
    ./threadbound build-pdf \
        --input "${OUTPUT_DIR}/book.tex" \
        --template-dir src/internal/templates/tex
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
echo "📝 TeX source: ${OUTPUT_DIR}/book.tex"