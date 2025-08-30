package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const revealTemplate = `<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/reveal.js@4.3.1/dist/reveal.css">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/reveal.js@4.3.1/dist/theme/white.css">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/highlight.js@11.7.0/styles/github.min.css">
    <style>
        .reveal h1, .reveal h2, .reveal h3 { color: #2c3e50; }
        .reveal .slides section { text-align: left; }
        .reveal h1, .reveal h2, .reveal h3 { text-align: center; }
        .reveal pre { width: 100%; }
        .reveal code { background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="reveal">
        <div class="slides">
            {{.Content}}
        </div>
    </div>
    <script src="https://cdn.jsdelivr.net/npm/reveal.js@4.3.1/dist/reveal.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/reveal.js@4.3.1/plugin/markdown/markdown.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/reveal.js@4.3.1/plugin/highlight/highlight.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/reveal.js@4.3.1/plugin/notes/notes.js"></script>
    <script>
        Reveal.initialize({
            hash: true,
            plugins: [ RevealMarkdown, RevealHighlight, RevealNotes ],
            markdown: {
                smartypants: true
            }
        });
    </script>
</body>
</html>`

func convertMarkdownToRevealHTML(markdown string) string {
	// Split markdown by slide separator (---)
	slides := strings.Split(markdown, "\n---\n")
	
	var htmlSlides []string
	for _, slide := range slides {
		slide = strings.TrimSpace(slide)
		if slide != "" {
			// Check if slide has vertical slides (separated by ---)
			if strings.Contains(slide, "\n--\n") {
				verticalSlides := strings.Split(slide, "\n--\n")
				var verticalHTML []string
				for _, vSlide := range verticalSlides {
					verticalHTML = append(verticalHTML, fmt.Sprintf(`<section data-markdown><textarea data-template>%s</textarea></section>`, vSlide))
				}
				htmlSlides = append(htmlSlides, fmt.Sprintf("<section>%s</section>", strings.Join(verticalHTML, "\n")))
			} else {
				htmlSlides = append(htmlSlides, fmt.Sprintf(`<section data-markdown><textarea data-template>%s</textarea></section>`, slide))
			}
		}
	}
	
	return strings.Join(htmlSlides, "\n")
}

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

	// Read file content
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert markdown to HTML
	markdownContent := buf.String()
	slidesHTML := convertMarkdownToRevealHTML(markdownContent)

	// Generate full HTML
	tmpl, err := template.New("reveal").Parse(revealTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var htmlBuf bytes.Buffer
	err = tmpl.Execute(&htmlBuf, struct{ Content template.HTML }{
		Content: template.HTML(slidesHTML),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save HTML file
	staticDir := "./static"
	os.MkdirAll(staticDir, os.ModePerm)
	
	// Save the markdown for reference
	mdPath := filepath.Join(staticDir, "slide.md")
	os.WriteFile(mdPath, []byte(markdownContent), 0644)
	
	// Save the HTML presentation
	htmlPath := filepath.Join(staticDir, "index.html")
	err = os.WriteFile(htmlPath, htmlBuf.Bytes(), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func main() {
	// Create static directory and default presentation
	staticDir := "./static"
	os.MkdirAll(staticDir, os.ModePerm)
	
	// Check if default presentation exists
	indexPath := filepath.Join(staticDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		defaultMarkdown := `# Welcome to Easy Slide

Use the upload button to upload your markdown file

---

## How to Use

1. Upload your markdown file
2. View your presentation here
3. Use arrow keys to navigate

---

## Markdown Syntax

- Use three dashes (---) to separate slides
- Use two dashes (--) for vertical slides
- Standard markdown formatting works
- Code blocks are syntax highlighted`

		slidesHTML := convertMarkdownToRevealHTML(defaultMarkdown)
		tmpl, _ := template.New("reveal").Parse(revealTemplate)
		var htmlBuf bytes.Buffer
		tmpl.Execute(&htmlBuf, struct{ Content template.HTML }{
			Content: template.HTML(slidesHTML),
		})
		os.WriteFile(indexPath, htmlBuf.Bytes(), 0644)
		os.WriteFile(filepath.Join(staticDir, "slide.md"), []byte(defaultMarkdown), 0644)
	}

	// Set up HTTP routes
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/health", healthHandler)
	
	// Serve static files (the presentation)
	http.Handle("/", http.FileServer(http.Dir(staticDir)))

	log.Println("Slide generation service starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}