#!/bin/bash

set -e

echo "🚀 Setting up Zoekt for your repository..."

# Install Go if not found
if ! command -v go &> /dev/null; then
    echo "📦 Installing Go..."
    brew install go
fi

# Add Go binaries to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Install Zoekt tools
echo "📥 Installing Zoekt tools..."
go install github.com/sourcegraph/zoekt/cmd/zoekt-git-index@latest
go install github.com/sourcegraph/zoekt/cmd/zoekt@latest
go install github.com/sourcegraph/zoekt/cmd/zoekt-webserver@latest

# Create index directory
echo "📁 Creating index directory..."
mkdir -p ~/.zoekt

# Index the repository
echo "🔍 Indexing repository: /Users/web1havv/vaibhav site"
"$(go env GOPATH)/bin/zoekt-git-index" -index ~/.zoekt "/Users/web1havv/vaibhav site"

echo ""
echo "✅ Done! Your repository is now indexed and ready to search!"
echo ""
echo "Try these commands:"
echo "  $(go env GOPATH)/bin/zoekt 'your search term'"
echo "  $(go env GOPATH)/bin/zoekt-webserver -index ~/.zoekt/"
echo ""
echo "Or start the web server and visit http://localhost:6070"





