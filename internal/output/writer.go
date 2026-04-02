// Package output handles file I/O operations for scraper results.
package output

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Writer manages the output directory and file persistence.
type Writer struct {
	BaseDir string
}

// NewWriter creates a Writer rooted at the given base directory.
func NewWriter(baseDir string) *Writer {
	return &Writer{BaseDir: baseDir}
}

// OutputPaths holds all generated file paths for a scrape session.
type OutputPaths struct {
	Dir            string // full directory path
	ChildDirName   string // directory basename (timestamp_site)
	CollyHTMLPath  string
	ChromeHTMLPath string
	ScreenshotPath string
	JSONPath       string
}

// Prepare creates the timestamped output directory and returns the paths
// that the caller can use to save results.
func (w *Writer) Prepare(targetURL string) (*OutputPaths, error) {
	siteName, err := ExtractSiteName(targetURL)
	if err != nil {
		return nil, fmt.Errorf("extract site name: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	childDirName := fmt.Sprintf("%s_%s", timestamp, siteName)

	fullDir := filepath.Join(w.BaseDir, childDirName)

	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	return &OutputPaths{
		Dir:            fullDir,
		ChildDirName:   childDirName,
		CollyHTMLPath:  filepath.Join(fullDir, childDirName+"_colly.html"),
		ChromeHTMLPath: filepath.Join(fullDir, childDirName+"_chromedp.html"),
		ScreenshotPath: filepath.Join(fullDir, childDirName+".png"),
		JSONPath:       filepath.Join(fullDir, childDirName+"_results.json"),
	}, nil
}

// SaveFile writes raw bytes to the given path. Returns nil if data is nil.
func (w *Writer) SaveFile(path string, data []byte) error {
	if data == nil {
		return nil
	}
	return os.WriteFile(path, data, 0644)
}

// SaveJSON marshals v to indented JSON and writes it to path.
func (w *Writer) SaveJSON(path string, v any) error {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	return os.WriteFile(path, jsonData, 0644)
}

// Cleanup removes the given directory and all its contents.
func (w *Writer) Cleanup(dir string) error {
	return os.RemoveAll(dir)
}

// ExtractSiteName derives a filesystem-safe name from a URL's hostname.
// Example: "https://www.example.com/path" → "example_com"
func ExtractSiteName(targetURL string) (string, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	site := strings.ReplaceAll(u.Hostname(), "www.", "")
	site = strings.ReplaceAll(site, ".", "_")
	return site, nil
}
