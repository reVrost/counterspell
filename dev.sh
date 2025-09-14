#!/bin/bash

# Get the project root directory
PROJECT_ROOT=$(pwd)

# Kill any existing marketpal server processes
pkill -f "local/marketpal serve" || true
# Kill any existing npm dev server processes
pkill -f "npm run dev" || true

# Ensure PROJECT_ROOT is an absolute path (optional but good practice)
PROJECT_ROOT=$(
  cd "$PROJECT_ROOT"
  pwd
)

echo "Launching Kitty panes from: $PROJECT_ROOT"

kitty @launch --location=vsplit --cwd "$PROJECT_ROOT" \
  zsh -ic 'direnv allow; air -c .air.toml; echo "Command finished, starting interactive zsh..."; exec zsh'

# Pane 2: Runs npm run dev in ui directory
# Use 'zsh -ic' and source ~/.zshrc for consistency
kitty @launch --location=vsplit --cwd "$PROJECT_ROOT/ui" \
  zsh -ic 'npm run dev; echo "Command finished, starting interactive zsh..."; exec zsh'

echo "Kitty launch commands sent."
