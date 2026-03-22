package main

import (
	"fmt"
	"mime"
	"net/http"
	"os"
)

func main() {
	port := ":3000"
	
	// CRITICAL FIX: Add special mime type for Wasm globally.
	// Windows often doesn't know what a .wasm file is, causing the browser
	// to reject it for security reasons if we don't explicitly declare it universally.
	mime.AddExtensionType(".wasm", "application/wasm")

	fmt.Println("──────────────────────────────────────────")
	fmt.Printf("🏥 ClinLang Web App is running!\n")
	fmt.Printf("🔗 Open your browser to: http://localhost%s\n", port)
	fmt.Println("──────────────────────────────────────────")
	fmt.Println("Press Ctrl+C to stop the server.")

	err := http.ListenAndServe(port, http.FileServer(http.Dir(".")))
	if err != nil {
		fmt.Println("Fatal error:", err)
		os.Exit(1)
	}
}
