@echo off
setlocal

where go >nul 2>nul
if errorlevel 1 (
    echo Go가 설치되어 있지 않습니다.
    pause
    exit /b 1
)

echo [1/3] 기존 Windows 리소스 파일 정리...
del /q rsrc.syso 2>nul
del /q rsrc_windows_*.syso 2>nul

echo [2/3] Windows EXE 아이콘 리소스 생성...
go run github.com/tc-hib/go-winres@latest simply ^
  --icon app.ico ^
  --manifest gui ^
  --product-name "Quick Anydesk Connect" ^
  --file-description "Tray helper for fast AnyDesk remote connections" ^
  --original-filename "QuickAnydeskConnect.exe"

if errorlevel 1 (
    echo.
    echo Windows 리소스 생성 실패
    pause
    exit /b 1
)

echo [3/3] QuickAnydeskConnect.exe 빌드...
go build -trimpath -ldflags="-s -w -H windowsgui" -o QuickAnydeskConnect.exe .

if errorlevel 1 (
    echo.
    echo 빌드 실패
    pause
    exit /b 1
)

echo.
echo 빌드 완료: QuickAnydeskConnect.exe
echo.
echo 참고:
echo - app.ico는 EXE 내부 아이콘과 트레이 아이콘으로 내장됩니다.
echo - 빌드 후 배포할 때 app.ico는 필요하지 않습니다.
pause
