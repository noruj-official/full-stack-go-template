# Development server startup script for Windows PowerShell
# Run from anywhere - navigates to project root automatically

# Get the directory where this script is located
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Navigate to project root (one level up from scripts folder)
Set-Location (Join-Path $ScriptDir "..")

Write-Host ""
Write-Host "  ======================================" -ForegroundColor Cyan
Write-Host "   Go Starter Development Server" -ForegroundColor Cyan
Write-Host "  ======================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Project Root: $(Get-Location)" -ForegroundColor Gray
Write-Host ""

# Check if .env exists
if (-not (Test-Path .env)) {
    Write-Host "[!] No .env file found. Copying from .env.example..." -ForegroundColor Yellow
    Copy-Item .env.example .env
    Write-Host "[OK] Created .env file" -ForegroundColor Green
}

# Check if node_modules exists
if (-not (Test-Path node_modules)) {
    Write-Host "[*] Installing npm dependencies..." -ForegroundColor Yellow
    npm install
    Write-Host "[OK] npm dependencies installed" -ForegroundColor Green
}

# Build Tailwind CSS
Write-Host "[*] Building Tailwind CSS..." -ForegroundColor Yellow
npm run build
Write-Host "[OK] Tailwind CSS compiled" -ForegroundColor Green

# Check if Docker is available and start PostgreSQL
$dockerExists = Get-Command docker -ErrorAction SilentlyContinue
if ($dockerExists) {
    $dbRunning = docker ps 2>$null | Select-String "go_starter_db"
    if (-not $dbRunning) {
        Write-Host "[*] Starting PostgreSQL container..." -ForegroundColor Yellow
        docker-compose up -d
        Write-Host "[OK] PostgreSQL started" -ForegroundColor Green
        Start-Sleep -Seconds 2
    }
}

Write-Host ""
Write-Host "[*] Checking dev server status (port 3000)..." -ForegroundColor Green
$serverListening = $false
try {
    $conn = Get-NetTCPConnection -LocalPort 3000 -State Listen -ErrorAction Stop
    if ($conn) { $serverListening = $true }
} catch {
    # Fallback to netstat in environments without Get-NetTCPConnection
    $serverListening = netstat -ano | Select-String ":3000" | Select-String "LISTENING"
}

if ($serverListening) {
    Write-Host "[OK] Dev server already running on http://localhost:3000" -ForegroundColor Green
} else {
    Write-Host "[*] Starting Go server on http://localhost:3000" -ForegroundColor Cyan
    # Start the server without blocking this script
    Start-Process -FilePath "go" -ArgumentList "run","./cmd/server" | Out-Null
    Start-Sleep -Seconds 2
}

$resp = Read-Host "[?] Open preview in browser now? (Y/N)"
if ($resp -match '^(?i)y(es)?$') {
    Start-Process "http://localhost:3000/"
}
