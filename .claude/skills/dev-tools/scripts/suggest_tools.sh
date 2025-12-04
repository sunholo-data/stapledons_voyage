#!/bin/bash
# Analyze codebase for potential new CLI tools
# Usage: suggest_tools.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

cd "$PROJECT_ROOT"

echo "Analyzing codebase for CLI tool opportunities..."
echo ""

# Current CLI commands
echo "## Current CLI Commands"
echo ""
./bin/voyage help 2>/dev/null || go run ./cmd/cli help 2>/dev/null || echo "(CLI not built)"
echo ""

# Check for common patterns that could become CLI tools
echo "## Potential New Tools"
echo ""

# Check for save/load functionality
if grep -rq "SaveGame\|LoadGame" engine/ cmd/ 2>/dev/null; then
    echo "- voyage save: Save file inspection/management detected"
    echo "  Files: $(grep -rl "SaveGame\|LoadGame" engine/ cmd/ 2>/dev/null | head -3 | tr '\n' ' ')"
    echo ""
fi

# Check for starmap functionality
if [ -d "assets/starmap" ] || grep -rq "starmap\|Starmap" engine/ 2>/dev/null; then
    echo "- voyage starmap: Starmap data tools"
    echo "  Current starmap assets: $(ls assets/starmap 2>/dev/null | wc -l | tr -d ' ') files"
    echo ""
fi

# Check for config functionality
if [ -f "config.json" ] || grep -rq "LoadConfig\|config\.json" . 2>/dev/null; then
    echo "- voyage config: Configuration management"
    echo ""
fi

# Check for replay/recording functionality
if grep -rq "Record\|Replay\|recording" engine/ 2>/dev/null; then
    echo "- voyage replay: Input replay functionality detected"
    echo ""
fi

# Check for export functionality
if grep -rq "Export\|export" engine/ 2>/dev/null; then
    echo "- voyage export: Data export functionality"
    echo ""
fi

# Check for AILANG files
if ls sim/*.ail 1>/dev/null 2>&1; then
    echo "- voyage ail: AILANG code analysis"
    echo "  AILANG files: $(ls sim/*.ail | wc -l | tr -d ' ')"
    echo ""
fi

# Check for screenshot functionality
if grep -rq "screenshot\|Screenshot" engine/ 2>/dev/null; then
    echo "- voyage screenshot: Screenshot capture tools"
    echo ""
fi

echo "## Package Analysis"
echo ""
echo "Engine packages that might benefit from CLI tools:"
for dir in engine/*/; do
    pkg=$(basename "$dir")
    case $pkg in
        handlers|render|assets|display|bench|scenario|screenshot|starmap|save)
            file_count=$(find "$dir" -name "*.go" | wc -l | tr -d ' ')
            echo "  - $pkg/ ($file_count files)"
            ;;
    esac
done
echo ""

echo "## Suggestions"
echo ""
echo "Based on analysis, consider adding CLI tools for:"
echo "1. Data inspection (saves, config, starmap)"
echo "2. Testing utilities (scenarios, screenshots)"
echo "3. Development aids (AILANG analysis, asset generation)"
echo ""
echo "To add a new tool, edit cmd/cli/main.go and update the dev-tools skill."
