@echo off
cd /d "%~dp0"

where go >nul 2>nul
if errorlevel 1 (
    echo Go is not installed. Download it from https://go.dev/dl/
    pause
    exit /b 1
)

go build -ldflags="-H=windowsgui -s -w" -o NetCheck.exe
if errorlevel 1 (
    echo.
    echo Build FAILED. Check main.go for errors.
    pause
    exit /b 1
)

echo.
echo Done: NetCheck.exe
pause
