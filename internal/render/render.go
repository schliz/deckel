package render

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Renderer handles template parsing and rendering for full pages, HTMX partials,
// and out-of-band swaps.
type Renderer struct {
	templates   map[string]*template.Template
	templateDir string
	isDev       bool
	funcMap     template.FuncMap
	mu          sync.RWMutex
}

// New creates a Renderer that parses and caches all templates at startup.
// Each page template is parsed together with the layout and all partials/components.
// In dev mode, templates are re-parsed on every request.
func New(templateDir string, isDev bool, funcMap template.FuncMap) (*Renderer, error) {
	r := &Renderer{
		templateDir: templateDir,
		isDev:       isDev,
		funcMap:     funcMap,
	}

	if err := r.parseTemplates(); err != nil {
		return nil, err
	}

	return r, nil
}

// parseTemplates walks the template directory and builds a cached template set.
// Each page gets its own clone of the layout + partials + components.
func (r *Renderer) parseTemplates() error {
	templates := make(map[string]*template.Template)

	layoutDir := filepath.Join(r.templateDir, "layouts")
	pagesDir := filepath.Join(r.templateDir, "pages")
	partialsDir := filepath.Join(r.templateDir, "partials")
	componentsDir := filepath.Join(r.templateDir, "components")

	// Collect shared templates (layouts, partials, components).
	var shared []string
	for _, dir := range []string{layoutDir, partialsDir, componentsDir} {
		files, err := globHTML(dir)
		if err != nil {
			return fmt.Errorf("reading templates from %s: %w", dir, err)
		}
		shared = append(shared, files...)
	}

	// Parse each page template combined with shared templates.
	pages, err := globHTML(pagesDir)
	if err != nil {
		return fmt.Errorf("reading page templates: %w", err)
	}

	for _, page := range pages {
		name := templateName(pagesDir, page)

		files := make([]string, 0, len(shared)+1)
		files = append(files, shared...)
		files = append(files, page)

		t, err := template.New(filepath.Base(page)).Funcs(r.funcMap).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", name, err)
		}

		templates[name] = t
	}

	r.mu.Lock()
	r.templates = templates
	r.mu.Unlock()

	return nil
}

// Page renders a full page for normal requests, or just the page content block
// for HTMX requests (detected via HX-Request header).
func (r *Renderer) Page(w http.ResponseWriter, req *http.Request, name string, data any) {
	t, err := r.getTemplate(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// For HTMX requests, render only the "content" block.
	if isHTMX(req) {
		if err := t.ExecuteTemplate(w, "content", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// For full page requests, execute the base layout template.
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Fragment renders a named template fragment.
func (r *Renderer) Fragment(w http.ResponseWriter, req *http.Request, name string, data any) {
	t, err := r.getTemplate(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AppendOOB appends an out-of-band swapped fragment to the response.
// The fragment should contain hx-swap-oob attribute for HTMX processing.
func (r *Renderer) AppendOOB(w http.ResponseWriter, name string, data any) {
	// Find the template in any of the cached page templates.
	r.mu.RLock()
	var t *template.Template
	for _, tmpl := range r.templates {
		if lookup := tmpl.Lookup(name); lookup != nil {
			t = lookup
			break
		}
	}
	r.mu.RUnlock()

	if t == nil {
		http.Error(w, fmt.Sprintf("template %q not found", name), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getTemplate returns a cached template by name, re-parsing in dev mode.
func (r *Renderer) getTemplate(name string) (*template.Template, error) {
	if r.isDev {
		if err := r.parseTemplates(); err != nil {
			return nil, fmt.Errorf("re-parsing templates: %w", err)
		}
	}

	r.mu.RLock()
	t, ok := r.templates[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("template %q not found", name)
	}

	return t, nil
}

// isHTMX checks if the request was made by HTMX.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// globHTML returns all .html files in the given directory (non-recursive).
func globHTML(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".html") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files, nil
}

// templateName extracts a short name from a page template path.
// e.g., "/templates/pages/dashboard.html" → "dashboard"
func templateName(baseDir, path string) string {
	rel, _ := filepath.Rel(baseDir, path)
	return strings.TrimSuffix(rel, ".html")
}
