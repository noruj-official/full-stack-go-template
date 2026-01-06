@echo off
REM Development server with Air HMR and Tailwind Watcher
setlocal enabledelayedexpansion

REM Get the directory where this script is located
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%.."

echo.
echo  ======================================
echo   Full Stack Go Template - Air + Tailwind
echo  ======================================
echo.

REM Check if .env exists
if not exist .env (
    echo [!] No .env file found. Copying from .env.example...
    copy .env.example .env
    echo [OK] Created .env file
)

REM Check for Air installation
where air >nul 2>nul
if %errorlevel% neq 0 (
    echo [!] Air command not found in PATH. Checking GOPATH...
    for /f "tokens=*" %%i in ('go env GOPATH') do set GOPATH=%%i
    set "PATH=!GOPATH!\bin;%PATH%"
)

where air >nul 2>nul
if %errorlevel% neq 0 (
    echo [!] Air is not installed. Installing...
    go install github.com/air-verse/air@latest
    
    REM Refresh PATH after install
    for /f "tokens=*" %%i in ('go env GOPATH') do set GOPATH=%%i
    set "PATH=!GOPATH!\bin;%PATH%"
)

echo [*] Starting Tailwind Watcher (Background)...
start "Tailwind Watcher" /min npm run watch

echo [*] Starting Air Server...
air

echo.
echo [*] Cleaning up...
taskkill /FI "WINDOWTITLE eq Tailwind Watcher" /T /F >nul 2>nul
