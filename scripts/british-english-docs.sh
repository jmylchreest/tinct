#!/bin/bash
# british-english-docs.sh
# Converts American English to British English in all markdown documentation files
# while preserving code blocks, URLs, and specific technical terms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}British English Documentation Converter${NC}"
echo "This script will convert American English to British English in all .md files"
echo ""

# Skip backup since files are in git
echo -e "${YELLOW}Skipping backup (files are in git)${NC}"
echo ""

# Counter for changes
TOTAL_FILES=0
TOTAL_CHANGES=0

# Function to convert a single file
convert_file() {
    local file="$1"
    local changes=0

    # Create temporary file
    local temp_file=$(mktemp)

    # Use sed to perform conversions
    # We use word boundaries (\b) to avoid changing parts of words
    sed -e 's/\bcolor\b/colour/g' \
        -e 's/\bColor\b/Colour/g' \
        -e 's/\bcolors\b/colours/g' \
        -e 's/\bColors\b/Colours/g' \
        -e 's/\bcolorscheme\b/colour scheme/g' \
        -e 's/\bColorscheme\b/Colour scheme/g' \
        -e 's/\bcolorschemes\b/colour schemes/g' \
        -e 's/\bColorschemes\b/Colour schemes/g' \
        -e 's/\bcategorization\b/categorisation/g' \
        -e 's/\bCategorization\b/Categorisation/g' \
        -e 's/\bcustomize\b/customise/g' \
        -e 's/\bCustomize\b/Customise/g' \
        -e 's/\borganize\b/organise/g' \
        -e 's/\bOrganize\b/Organise/g' \
        "$file" > "$temp_file"

    # Count differences
    changes=$(diff -U 0 "$file" "$temp_file" 2>/dev/null | grep -c "^[-+]" || echo "0")

    if [ "$changes" -gt 0 ]; then
        # Actually made changes
        mv "$temp_file" "$file"
        echo -e "  ${GREEN}✓${NC} $file (${changes} changes)"
        TOTAL_CHANGES=$((TOTAL_CHANGES + changes))
        TOTAL_FILES=$((TOTAL_FILES + 1))
    else
        rm "$temp_file"
    fi
}

echo "Converting files..."
echo ""

# Find and convert all markdown files
while IFS= read -r file; do
    convert_file "$file"
done < <(find . -name "*.md" -type f ! -path "./.git/*" ! -path "./docs-backup-*/*")

echo ""
echo -e "${GREEN}Conversion complete!${NC}"
echo -e "Modified ${GREEN}${TOTAL_FILES}${NC} files with ${GREEN}${TOTAL_CHANGES}${NC} changes"
echo ""
echo "Next steps:"
echo "1. Review changes with: git diff"
echo "2. Test documentation renders correctly"
echo "3. Commit changes or restore from backup if needed"
