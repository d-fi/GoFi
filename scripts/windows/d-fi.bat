@echo off
setlocal

set "DFI_EXE=%~dp0d-fi.exe"

if not exist "%DFI_EXE%" (
  echo d-fi.exe was not found beside this launcher.
  echo Put d-fi.bat in the same folder as d-fi.exe.
  pause
  exit /b 1
)

:menu
cls
echo d-fi
echo.
echo 1^) Start CLI
echo 2^) Start Web UI
echo 3^) Set Deezer ARL
echo 4^) Exit
echo.
set /p "choice=Select option: "

if "%choice%"=="1" goto cli
if "%choice%"=="2" goto web
if "%choice%"=="3" goto arl
if "%choice%"=="4" goto done

echo.
echo Invalid option.
pause
goto menu

:cli
cls
"%DFI_EXE%"
echo.
pause
goto menu

:web
cls
echo Starting web UI at http://127.0.0.1:8080
echo Press Ctrl+C to stop the server.
echo.
"%DFI_EXE%" web
echo.
pause
goto menu

:arl
cls
echo Paste your Deezer ARL cookie.
echo.
set /p "arl=ARL: "
if "%arl%"=="" (
  echo.
  echo ARL was empty.
  pause
  goto menu
)
"%DFI_EXE%" --set-arl "%arl%"
echo.
pause
goto menu

:done
endlocal
