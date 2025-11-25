#!/bin/bash

# Script to track progress of filling remaining documentation pages
# This helps us know which pages have been completed

DOCS_DIR="docs"

echo "==================================="
echo "Documentation Filling Progress"
echo "==================================="
echo ""

# Count total markdown files
TOTAL=$(find "$DOCS_DIR" -name "*.md" -not -path "*/node_modules/*" | wc -l | tr -d ' ')

# Count files with "Coming Soon"
REMAINING=$(find "$DOCS_DIR" -name "*.md" -exec grep -l "Coming Soon" {} \; 2>/dev/null | wc -l | tr -d ' ')

# Calculate completed
COMPLETED=$((TOTAL - REMAINING))
PERCENT=$((COMPLETED * 100 / TOTAL))

echo "Total Pages: $TOTAL"
echo "Completed: $COMPLETED ($PERCENT%)"
echo "Remaining: $REMAINING"
echo ""

# List remaining pages
if [ "$REMAINING" -gt 0 ]; then
    echo "Pages still needing content:"
    find "$DOCS_DIR" -name "*.md" -exec grep -l "Coming Soon" {} \; 2>/dev/null | sed 's|^docs/||' | sort
fi

echo ""
echo "==================================="

