#!/bin/bash

# Cleanup script for Counterspell development
# Removes test databases and WAL files created during development and testing

echo "🧹 Cleaning up Counterspell development files..."

# Remove test database files
echo "Removing test databases..."
rm -f test*.db
rm -f counterspell*.db
rm -f examples/*/counterspell*.db

# Remove WAL and SHM files (DuckDB transaction logs)
echo "Removing DuckDB WAL files..."
rm -f *.db.wal *.db.shm
rm -f examples/*/*.db.wal examples/*/*.db.shm

# Remove build artifacts
echo "Removing build artifacts..."
rm -f counterspell-server
rm -f bin/*

echo "✅ Cleanup complete!"
echo ""
echo "Note: .wal files are DuckDB's Write-Ahead Log files for transaction safety."
echo "They're automatically created during normal database operations."
echo "You can safely delete them when not running the application." 