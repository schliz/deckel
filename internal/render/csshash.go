package render

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CSSHashedPath computes a SHA-256 content hash of the given CSS file and
// creates a hard link with the hash embedded in the filename. Returns the
// URL path (e.g., "/static/css/styles.a3f8b2c1.css") for use in templates.
// If the source file does not exist, it returns the unhashed path as fallback.
func CSSHashedPath(staticDir, cssFile string) (string, error) {
	srcPath := filepath.Join(staticDir, "css", cssFile)

	data, err := os.ReadFile(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Fallback: return unhashed path (e.g., dev mode without built CSS).
			return "/static/css/" + cssFile, nil
		}
		return "", fmt.Errorf("reading CSS file %s: %w", srcPath, err)
	}

	hash := sha256.Sum256(data)
	hashStr := fmt.Sprintf("%x", hash[:8]) // 16 hex chars, enough for uniqueness

	ext := filepath.Ext(cssFile)
	base := strings.TrimSuffix(cssFile, ext)
	hashedName := fmt.Sprintf("%s.%s%s", base, hashStr, ext)
	hashedPath := filepath.Join(staticDir, "css", hashedName)

	// Remove old hashed files matching the pattern styles.*.css.
	matches, _ := filepath.Glob(filepath.Join(staticDir, "css", base+".*.css"))
	for _, m := range matches {
		if m != srcPath && m != hashedPath {
			os.Remove(m)
		}
	}

	// Create hard link (or copy) for the hashed filename.
	if _, err := os.Stat(hashedPath); os.IsNotExist(err) {
		if err := os.Link(srcPath, hashedPath); err != nil {
			// Fallback: copy the file if hard link fails.
			if err := os.WriteFile(hashedPath, data, 0644); err != nil {
				return "", fmt.Errorf("creating hashed CSS file: %w", err)
			}
		}
	}

	return "/static/css/" + hashedName, nil
}
