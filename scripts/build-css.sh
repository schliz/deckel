#!/usr/bin/env bash
# Generate Tailwind + DaisyUI CSS from input.css
# Requires: npm install (run once to install tailwindcss + daisyui)
set -euo pipefail
cd "$(dirname "$0")/.."
npx tailwindcss -i static/css/input.css -o static/css/styles.css --minify
