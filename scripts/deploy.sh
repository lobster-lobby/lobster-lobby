#!/bin/bash
set -euo pipefail

REPO_DIR="$HOME/clawd/vault/dev/repos/lobster-lobby"
DEPLOY_DIR="$HOME/lobster-lobby"

echo "=== Lobster Lobby Deploy ==="

# Pull latest
cd "$REPO_DIR"
git pull origin main

# Build frontend
echo "Building frontend..."
cd "$REPO_DIR/frontend"
npm run build
rm -rf "$DEPLOY_DIR/frontend/dist"
mkdir -p "$DEPLOY_DIR/frontend/dist"
cp -r dist/* "$DEPLOY_DIR/frontend/dist/"

# Build backend
echo "Building backend..."
cd "$REPO_DIR/backend"
go build -o "$DEPLOY_DIR/backend" ./cmd/server

# Restart service
echo "Restarting service..."
systemctl --user restart lobster-lobby-api.service
sleep 2

# Health check
STATUS=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8090/api/policies)
if [ "$STATUS" = "200" ]; then
    echo "✅ Deploy successful! API health check passed."
else
    echo "❌ Deploy failed! API returned $STATUS"
    exit 1
fi
