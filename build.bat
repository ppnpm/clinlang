@echo off
echo Compiling ClinLang for WebAssembly...

set GOOS=js
set GOARCH=wasm
go build -o web/clinlang.wasm ./cmd/wasm

if %errorlevel% neq 0 (
    echo.
    echo ❌ Build failed!
    pause
    exit /b %errorlevel%
)

echo.
echo ✅ Successfully built web/clinlang.wasm!
echo You can now open index.html in your browser.
pause
