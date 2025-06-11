# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Folder Management System (フォルダー管理システム) that provides a web interface for browsing and managing folders. It consists of:
- **Backend**: Go (1.21) with Fiber v2 framework
- **Frontend**: React (19.1.0) with TypeScript and Vite

## Development Commands

### Frontend Development
```bash
cd frontend
npm install          # Install dependencies
npm run dev          # Start dev server (http://localhost:5173)
npm run build        # Build for production
npm run lint         # Run ESLint
npm run preview      # Preview production build
```

### Backend Development
```bash
cd backend
go mod tidy          # Install/update dependencies
go run cmd/main.go   # Start server (http://localhost:8080)
```

## Architecture

### Backend Structure
- `cmd/main.go`: Entry point, sets up Fiber server with CORS
- `internal/handlers/`: HTTP request handlers (folder_handler.go)
- `internal/services/`: Business logic (folder_service.go)
- `internal/models/`: Data models (folder.go, instant.go)

The backend serves a REST API at `http://localhost:8080/api` with the main endpoint:
- `GET /api/folders?path=<optional-path>` - Returns folder contents

### Frontend Structure
- `src/App.tsx`: Main app component with routing
- `src/components/`: UI components (FolderGrid, FolderModal)
- `src/services/api.ts`: Backend API client
- `src/types/`: TypeScript type definitions

### Key Implementation Details

1. **Default Path**: The system defaults to browsing `~/penguin/2-工事` directory
2. **CORS**: Backend allows all origins with `AllowOrigins: "*"`
3. **File Type Detection**: Frontend displays different icons based on file extensions:
   - Folders: 📁
   - PDFs: 📄
   - Images (jpg, jpeg, png, gif): 🖼️
   - Videos (mp4, avi, mov): 🎬
   - Audio (mp3, wav): 🎵
   - Others: 📎

4. **API Response Format**: The backend returns an array of folder items with properties like name, path, size, isDirectory, etc.