# Penguin Backend & Frontend Commands

# Start the backend server
backend:
    cd ./backend && go run cmd/main.go

# Start the frontend development server
frontend:
    cd ./frontend && npm run dev

# Install backend dependencies
backend-deps:
    cd ./backend && go mod tidy

# Install frontend dependencies  
frontend-deps:
    cd ./frontend && npm install

# Build frontend for production
frontend-build:
    cd ./frontend && npm run build

# Run frontend linting
frontend-lint:
    cd ./frontend && npm run lint

# Start both backend and frontend (requires tmux or run in separate terminals)
dev:
    @echo "Starting backend and frontend..."
    @echo "Run 'just backend' in one terminal and 'just frontend' in another"

# Clean and reinstall all dependencies
clean-install: backend-deps frontend-deps

# Kill process running on port 8080
kill-port:
    @echo "Stopping process on port 8080..."
    @-pkill -f ":8080" 2>/dev/null || true
    @-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
    @echo "Port 8080 cleanup completed"

# Update claude-code
claude-code:
    npm i -g @anthropic-ai/claude-code

# Show backend architecture diagram in browser
architecture:
    @echo "Opening architecture diagram..."
    @xdg-open "https://mermaid.live/edit#$(cat doc/backend-architecture.md | grep -A 100 '```mermaid' | grep -B 100 '```' | grep -v '```' | base64 -w 0)" 2>/dev/null || open "https://mermaid.live/" 2>/dev/null || echo "Please visit https://mermaid.live/ and paste the mermaid code from doc/backend-architecture.md"

# Show available commands
help:
    @just --list