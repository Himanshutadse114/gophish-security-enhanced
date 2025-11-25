@echo off
REM ============================================================
REM Gophish - One-Step Setup Script for Windows
REM ============================================================
REM This script will:
REM   1. Check if Docker is installed and running
REM   2. Build the Gophish Docker image
REM   3. Start the container
REM   4. Display login credentials and access URL
REM ============================================================

setlocal enabledelayedexpansion

set IMAGE_NAME=gophish-unified
set CONTAINER_NAME=gophish
set ERROR=0

echo.
echo ============================================================
echo   Gophish - One-Step Setup
echo ============================================================
echo.

REM Check if Docker is installed
echo [1/5] Checking Docker installation...
docker --version >nul 2>&1
if errorlevel 1 (
    REM Try with full path (common Docker Desktop location)
    "%ProgramFiles%\Docker\Docker\resources\bin\docker.exe" --version >nul 2>&1
    if errorlevel 1 (
        REM Try alternative path
        "%ProgramFiles(x86)%\Docker\Docker\resources\bin\docker.exe" --version >nul 2>&1
        if errorlevel 1 (
            echo.
            echo ============================================================
            echo   [ERROR] Docker is not installed!
            echo ============================================================
            echo.
            echo Docker Desktop is required to run Gophish.
            echo Please install Docker Desktop first, then run this script again.
            echo.
            echo Download from: https://www.docker.com/products/docker-desktop
            echo.
            echo ============================================================
            echo.
            pause
            exit /b 1
        )
    )
)
REM Docker found - show version
docker --version 2>nul
if errorlevel 1 (
    "%ProgramFiles%\Docker\Docker\resources\bin\docker.exe" --version 2>nul
    if errorlevel 1 (
        "%ProgramFiles(x86)%\Docker\Docker\resources\bin\docker.exe" --version 2>nul
    )
)
echo    [OK] Docker is installed

REM Check if Docker daemon is running
echo [2/5] Checking if Docker is running...
docker ps >nul 2>&1
if errorlevel 1 (
    REM Try with full path
    "%ProgramFiles%\Docker\Docker\resources\bin\docker.exe" ps >nul 2>&1
    if errorlevel 1 (
        "%ProgramFiles(x86)%\Docker\Docker\resources\bin\docker.exe" ps >nul 2>&1
        if errorlevel 1 (
            echo.
            echo ============================================================
            echo   [ERROR] Docker Desktop is not running!
            echo ============================================================
            echo.
            echo Please start Docker Desktop from Start Menu
            echo Wait for it to fully start - look for whale icon in system tray
            echo Then run this script again
            echo.
            echo ============================================================
            echo.
            pause
            exit /b 1
        )
    )
)
echo    [OK] Docker is running

REM Stop and remove existing container if it exists
echo [3/5] Cleaning up existing containers...
docker ps -a | findstr /C:"%CONTAINER_NAME%" >nul 2>&1
if not errorlevel 1 (
    echo    Stopping existing container...
    docker stop %CONTAINER_NAME% >nul 2>&1
    docker rm %CONTAINER_NAME% >nul 2>&1
    echo    [OK] Old container removed
) else (
    echo    [OK] No existing container found
)

REM Check if image exists, if not build it
echo [4/5] Checking Docker image...
docker images | findstr /C:"%IMAGE_NAME%" >nul 2>&1
if errorlevel 1 (
    echo    Image not found. Building Docker image...
    echo    This may take 5-10 minutes on first run...
    echo.
    docker build -f Dockerfile.unified -t %IMAGE_NAME%:latest .
    if errorlevel 1 (
        echo.
        echo [ERROR] Docker build failed!
        echo.
        echo Please check the error messages above.
        echo Make sure you have enough disk space and memory.
        echo.
        pause
        exit /b 1
    )
    echo    [OK] Image built successfully
) else (
    echo    [OK] Image already exists
)

REM Run the container
echo [5/5] Starting Gophish container...
docker run -d --name %CONTAINER_NAME% -p 8443:8443 -p 8080:8080 -v gophish-data:/opt/gophish %IMAGE_NAME%:latest
if errorlevel 1 (
    echo.
    echo [ERROR] Failed to start container!
    echo.
    echo Possible issues:
    echo   - Port 8443 or 8080 is already in use
    echo   - Insufficient system resources
    echo.
    echo To check what's using the ports:
    echo   netstat -an ^| findstr "8443 8080"
    echo.
    pause
    exit /b 1
)
echo    [OK] Container started

echo.
echo Fixing database permissions...
docker exec %CONTAINER_NAME% chown -R app:app /opt/gophish 2>nul
docker exec %CONTAINER_NAME% chmod -R 775 /opt/gophish 2>nul
docker exec %CONTAINER_NAME% bash -c "if [ -f /opt/gophish/gophish.db ]; then chown app:app /opt/gophish/gophish.db && chmod 664 /opt/gophish/gophish.db; fi" 2>nul

echo.
echo ============================================================
echo   Setup Complete!
echo ============================================================
echo.
echo Waiting for services to initialize (45 seconds)...
echo This is normal - Gophish needs time to start up.
echo.
timeout /t 45 /nobreak >nul

echo.
echo ============================================================
echo   Login Credentials
echo ============================================================
echo.
REM Try to extract credentials from logs
docker logs %CONTAINER_NAME% 2>&1 | findstr /I "login" | findstr /I "username" | findstr /I "password" >nul 2>&1
if errorlevel 1 (
    echo    Credentials not found yet. Waiting a bit more...
    timeout /t 15 /nobreak >nul
)

REM Display login credentials - look for the line with "Please login"
for /f "tokens=*" %%i in ('docker logs %CONTAINER_NAME% 2^>^&1 ^| findstr /I "Please login"') do (
    echo    %%i
    set CREDS_FOUND=1
)

if not defined CREDS_FOUND (
    echo    [INFO] Searching for credentials in logs...
    echo.
    echo    Run this command to see credentials:
    echo    docker logs %CONTAINER_NAME% ^| findstr /I "login"
    echo.
    echo    Or use: get-password.bat
    echo.
)

echo.
echo ============================================================
echo   Access Information
echo ============================================================
echo.
echo    Admin Dashboard:  https://localhost:8443
echo    Landing Pages:    http://localhost:8080
echo.
echo    NOTE: You will see a security warning because we use
echo    self-signed certificates. Click "Advanced" then
echo    "Proceed to localhost" - this is safe for local use.
echo.

REM Check if services are running
echo ============================================================
echo   Service Status
echo ============================================================
echo.
docker ps | findstr /C:"%CONTAINER_NAME%" >nul 2>&1
if errorlevel 1 (
    echo    [WARNING] Container is not running!
    echo    Check logs with: docker logs %CONTAINER_NAME%
) else (
    echo    [OK] Container is running
    echo.
    echo    To view logs: docker logs %CONTAINER_NAME%
    echo    To stop:      docker stop %CONTAINER_NAME%
    echo    To start:    docker start %CONTAINER_NAME%
    echo    To remove:   docker stop %CONTAINER_NAME% ^&^& docker rm %CONTAINER_NAME%
)

echo.
echo ============================================================
echo   Useful Commands
echo ============================================================
echo.
echo Get password:     get-password.bat
echo View logs:        view-logs.bat
echo Check status:     check-status.bat
echo Reset database:   reset-database.bat
echo Fix permissions: fix-permissions-complete.bat
echo.
echo ============================================================
echo.
pause

