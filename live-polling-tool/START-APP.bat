@echo off
setlocal
cd /d "%~dp0"
title Live Polling Tool - Startup

echo ==========================================
echo       LIVE POLLING TOOL - STARTUP
echo ==========================================
echo.
where docker >nul 2>nul
if errorlevel 1 (
  echo ERROR: Docker CLI not found.
  echo Install Docker Desktop, open it, and wait until it is running:
  echo https://www.docker.com/products/docker-desktop/
  echo Then run this file again.
  pause
  exit /b 1
)
where node >nul 2>nul
if errorlevel 1 (
  echo ERROR: Node.js is not installed or not on PATH.
  echo Install the LTS version from https://nodejs.org/ then reopen this file.
  pause
  exit /b 1
)

echo Checking Docker engine...
docker info >nul 2>nul
if errorlevel 1 (
  echo ERROR: Docker Desktop is installed but is not running.
  echo Start Docker Desktop, wait for it to finish loading, then run this file again.
  pause
  exit /b 1
)

echo Starting MongoDB, Redis, and Go/Gin backend...
docker compose up -d --build mongodb redis backend
if errorlevel 1 (
  echo ERROR: Docker Compose could not start services. Review the error above.
  pause
  exit /b 1
)

echo Installing frontend packages if needed...
if not exist "frontend\node_modules" (
  pushd frontend
  call npm install
  if errorlevel 1 (
    popd
    echo ERROR: npm install failed.
    pause
    exit /b 1
  )
  popd
)

echo Opening frontend development server in a new window...
start "Live Polling Tool - Frontend" cmd /k "cd /d "%~dp0frontend" && npm run dev"

echo.
echo Startup commands completed.
echo Frontend: http://localhost:5173
echo Backend health: http://localhost:8080/health
echo.
echo Keep the frontend window and Docker Desktop open while using the app.
echo To stop backend and databases later, run: docker compose down
pause
