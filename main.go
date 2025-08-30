package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var (
	revealMDProcess *exec.Cmd
	revealMDMutex   sync.Mutex
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (up to 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create slides directory
	slidesDir := "./slides"
	err = os.MkdirAll(slidesDir, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save uploaded file as slide.md
	dstPath := filepath.Join(slidesDir, "slide.md")
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy uploaded content to file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Restart reveal-md with new slide
	go restartRevealMD()

	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
}

func restartRevealMD() {
	revealMDMutex.Lock()
	defer revealMDMutex.Unlock()

	// Kill existing reveal-md process if running
	if revealMDProcess != nil && revealMDProcess.Process != nil {
		log.Println("Stopping existing reveal-md process...")
		revealMDProcess.Process.Kill()
		revealMDProcess.Wait()
	}

	// Small delay to ensure port is released
	time.Sleep(500 * time.Millisecond)

	// Start reveal-md with the new slide
	log.Println("Starting reveal-md with updated slides...")
	revealMDProcess = exec.Command("reveal-md", "./slides/slide.md", "--host", "0.0.0.0", "--port", "1948")
	revealMDProcess.Stdout = os.Stdout
	revealMDProcess.Stderr = os.Stderr
	
	if err := revealMDProcess.Start(); err != nil {
		log.Printf("Failed to start reveal-md: %v", err)
	} else {
		log.Println("reveal-md started successfully")
	}
}

func slideProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Proxy requests to reveal-md server
	target, _ := url.Parse("http://localhost:1948")
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Handle errors when reveal-md is not ready
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		http.Error(w, "Slide server is starting, please refresh in a moment", http.StatusServiceUnavailable)
	}
	
	proxy.ServeHTTP(w, r)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func main() {
	// Create default slide if none exists
	slidesDir := "./slides"
	os.MkdirAll(slidesDir, os.ModePerm)
	defaultSlide := filepath.Join(slidesDir, "slide.md")
	if _, err := os.Stat(defaultSlide); os.IsNotExist(err) {
		defaultContent := "# Welcome to Easy Slide\n\n---\n\n## How to Use\n\n1. Upload your markdown file using the upload button\n2. View your presentation here\n3. Use arrow keys to navigate\n\n---\n\n## Markdown Syntax\n\n- Use `---` to separate slides\n- Use `#` for headings\n- Use `-` for bullet points\n- Support for code blocks, images, and more!"
		os.WriteFile(defaultSlide, []byte(defaultContent), 0644)
	}

	// Start reveal-md with default slide
	go restartRevealMD()

	// Set up HTTP routes
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", slideProxyHandler)

	log.Println("Slide generation service starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}