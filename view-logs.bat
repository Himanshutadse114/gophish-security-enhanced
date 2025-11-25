@echo off
REM ============================================================
REM Gophish - View Logs
REM ============================================================
REM This script shows the Gophish container logs
REM ============================================================

set CONTAINER_NAME=gophish

echo.
echo ============================================================
echo   Gophish - Container Logs
echo ============================================================
echo.

REM Check if container exists
docker ps -a | findstr /C:"%CONTAINER_NAME%" >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Container '%CONTAINER_NAME%' not found!
    echo.
    echo Make sure Gophish is running. Run setup.bat first.
    echo.
    pause
    exit /b 1
)

echo Showing last 50 lines of logs...
echo (Press Ctrl+C to stop, or wait for it to finish)
echo.
echo ============================================================
echo.

REM Show last 50 lines
docker logs --tail 50 %CONTAINER_NAME%

echo.
echo ============================================================
echo.
echo Commands:
echo   - View all logs:        docker logs %CONTAINER_NAME%
echo   - Follow logs (live):    docker logs -f %CONTAINER_NAME%
echo   - Last 100 lines:       docker logs --tail 100 %CONTAINER_NAME%
echo   - Find login:           docker logs %CONTAINER_NAME% ^| findstr /I "login"
echo.
pause

