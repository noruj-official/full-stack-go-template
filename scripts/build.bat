@echo off
rem Navigate to project root
cd /d "%~dp0.."

rem Build CSS
echo Building CSS...
call npm run build
if %errorlevel% neq 0 (
    echo [Error] CSS build failed
    exit /b %errorlevel%
)

rem Generate templ files
echo Generating templ files...
go run github.com/a-h/templ/cmd/templ generate
if %errorlevel% neq 0 (
    echo [Error] templ generation failed
    exit /b %errorlevel%
)

rem Build the application
echo Building application...
go build -o tmp\main.exe .\cmd\server
if %errorlevel% neq 0 (
    echo [Error] build failed
    exit /b %errorlevel%
)

echo Build completed successfully!
