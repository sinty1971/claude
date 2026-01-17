set shell := ["powershell", "-NoLogo", "-Command"]

# buf コマンドのパス設定
buf := `"$env:USERPROFILE/.local/bin/buf.exe"`

# API Server & Frontend Commands

# Start the API server (HTTP/2 over h2c)
api:
    cd ./web-api ; go run cmd/server/main.go

# Test file service
test-file *ARGS:
    cd ./web-api ; go run cmd/test/file/main.go {{ARGS}}

# Test company service
test-company *ARGS:
    cd ./web-api ; go run cmd/test/company/main.go {{ARGS}}

# Test koji service
test-koji *ARGS:
    cd ./web-api ; go run cmd/test/koji/main.go {{ARGS}}

# Start the API server with TLS enabled
api-tls:
    cd ./web-api ; go run cmd/server/main.go -enable-tls

# Start the web development server (React Router v7)
web:
    cd ./web ; npm run dev

# Start the Svelte development server (Bun)
svelte:
    cd ./my-svelte-app ; bun run dev

# Generate SSL certificate for HTTP/2
generate-cert:
    cd ./web-api ; ./generate-cert.sh

# Update all Go packages to latest versions
# ただしメジャーなバージョンは更新しない
api-update:
    cd ./web-api ; go get -u ./...
    cd ./web-api ; go mod tidy

# Generate gRPC stubs for all services
generate-grpc:
    @echo "Generating gRPC stubs..."
    @ {{buf}} generate

# Generate Connect-Web stubs for the web frontend
web-generate-grpc: generate-grpc
    @echo "Connect-Web stubs generated at web/src/gen/ and my-svelte-app/src/gen/"

# Install web dependencies  
web-deps:
    cd ./web && npm install

# Update all npm packages to latest versions
web-update:
    cd ./web && npm update
    cd ./web && npm audit fix

# Generate TypeScript types from OpenAPI spec
generate-types: web-generate-grpc

# Generate React Router v7 route structure diagram
generate-routes:
    @echo "Next.js へ移行したため自動ルート図生成は未対応です"
    @echo "app ディレクトリ構成を直接参照してください"

# Build web for production (React Router v7)
web-build:
    cd ./web && npm run build

# Preview web production build
web-preview:
    cd ./web && npm run preview

# Run web linting
web-lint:
    cd ./web && npm run lint

# Start both API server and web (requires tmux or run in separate terminals)
dev:
    @echo "Starting API server and web..."
    @echo "Run 'just api' in one terminal and 'just web' in another"

# Generate both API and web gRPC stubs
generate-all: generate-grpc generate-types

# Install API server dependencies
api-deps:
    cd ./web-api && go mod tidy
    cd ./web-api && go mod download

# Update all dependencies (Go and npm)
update-all: api-update web-update

# Clean and reinstall all dependencies
clean-install: api-deps web-deps

# Stop API server (Go application)
stop-api:
    #!/bin/bash
    set +e
    echo "Stopping API server..."
    pkill -f "go run cmd/server/main.go" 2>/dev/null
    pkill -f "cmd/server/main.go" 2>/dev/null
    pkill -f "web-api" 2>/dev/null
    for port in 9090 9443; do
        lsof -ti:$port | xargs -r kill -15 2>/dev/null
    done
    sleep 1
    for port in 9090 9443; do
        lsof -ti:$port | xargs -r kill -9 2>/dev/null
    done
    echo "API server stopped"

# Stop web development server (React Router v7 / Vite)
stop-web:
    #!/bin/bash
    set +e
    echo "Stopping web development server..."
    pkill -f "npm run dev" 2>/dev/null
    pkill -f "react-router dev" 2>/dev/null
    pkill -f "vite" 2>/dev/null
    pkill -f "node.*vite" 2>/dev/null
    for port in 5173 5174 5175 5176 5177; do
        lsof -ti:$port | xargs -r kill -15 2>/dev/null
    done
    sleep 1
    for port in 5173 5174 5175 5176 5177; do
        lsof -ti:$port | xargs -r kill -9 2>/dev/null
    done
    echo "Web development server stopped"

# Stop both API and web servers
stop-all: stop-api stop-web
    @echo "All servers stopped"

# Restart API server
restart-api: stop-api
    @echo "Restarting API server..."
    @sleep 1
    cd ./web-api && go run cmd/server/main.go &
    @echo "API server restarted"

# Restart web development server
restart-web: stop-web
    @echo "Restarting web development server..."
    @sleep 1
    cd ./web && npm run dev &
    @echo "Web development server restarted"

# Restart both API and web servers
restart-all: stop-all
    @echo "Restarting all servers..."
    @sleep 2
    cd ./web-api && go run cmd/server/main.go &
    @sleep 1
    cd ./web && npm run dev &
    @echo "All servers restarted"

# Kill process running on port 9090 (legacy command)
kill-port:
    @echo "Stopping process on port 9090..."
    @-lsof -ti:9090 | xargs -r kill -15 2>/dev/null || true
    @sleep 1
    @-lsof -ti:9090 | xargs -r kill -9 2>/dev/null || true
    @echo "Port 9090 cleanup completed"

# Update claude-code
claude-code:
    npm i -g @anthropic-ai/claude-code

# Show API architecture diagram in browser
architecture:
    @echo "Opening architecture diagram..."
    @xdg-open "https://mermaid.live/edit#$(cat doc/server-architecture.md | grep -A 100 '```mermaid' | grep -B 100 '```' | grep -v '```' | base64 -w 0)" 2>/dev/null || open "https://mermaid.live/" 2>/dev/null || echo "Please visit https://mermaid.live/ and paste the mermaid code from doc/server-architecture.md"

# Show API documentation in browser
docs:
    @echo "Opening API documentation..."
    @if command -v xdg-open > /dev/null; then xdg-open docs/proto/apis.md; elif command -v open > /dev/null; then open docs/proto/apis.md; else cat docs/proto/apis.md; fi

# Show available commands
help:
    @just --list
