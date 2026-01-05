@echo off
REM Development server startup script for Windows
REM Run from anywhere - navigates to project root automatically
setlocal enabledelayedexpansion

REM Get the directory where this script is located
set "SCRIPT_DIR=%~dp0"

REM Navigate to project root (one level up from scripts folder)
cd /d "%SCRIPT_DIR%.."

echo.
echo  ======================================
echo   Go Starter Development Server
echo  ======================================
echo.
echo  Project Root: %CD%
echo.

REM Check if .env exists
if not exist .env (
    echo [!] No .env file found. Copying from .env.example...
    copy .env.example .env
    echo [OK] Created .env file
)

REM Check if node_modules exists
if not exist node_modules (
    echo [*] Installing npm dependencies...
    call npm install
    echo [OK] npm dependencies installed
)

REM Build Tailwind CSS
echo [*] Building Tailwind CSS...
call npm run build
echo [OK] Tailwind CSS compiled

REM Check if Docker is available and start PostgreSQL
where docker >nul 2>nul
if %errorlevel% equ 0 (
    docker ps | findstr go_starter_db >nul 2>nul
    if %errorlevel% neq 0 (
        echo [*] Starting PostgreSQL container...
        docker-compose up -d
        echo [OK] PostgreSQL started
        timeout /t 2 /nobreak >nul
    )
)

echo.
echo [*] Checking dev server status (port 3000)...
netstat -ano | findstr LISTENING | findstr :3000 >nul 2>nul
if %errorlevel% equ 0 (
    echo [OK] Dev server already running on http://localhost:3000
) else (
    echo [*] Starting Go server on http://localhost:3000
    REM Start server in a new window so this script can continue
    start "Go Starter Server" go run ./cmd/server
    REM Give the server a moment to initialize
    timeout /t 2 /nobreak >nul
)

echo.
set /p START_PREVIEW=[?] Open preview in browser now? (Y/N): 
if /I "%START_PREVIEW%"=="Y" (
    start "" http://localhost:3000/
)
