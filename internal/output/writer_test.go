package output

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractSiteName(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"simple", "https://example.com", "example_com"},
		{"with www", "https://www.example.com", "example_com"},
		{"subdomain", "https://sub.example.com", "sub_example_com"},
		{"with path", "https://example.com/path/to/page", "example_com"},
		{"turkish domain", "https://www.trendyol.com", "trendyol_com"},
		{"edu domain", "https://avesis.erciyes.edu.tr", "avesis_erciyes_edu_tr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSiteName(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ExtractSiteName(%q) = %q, want %q", tt.url, got, tt.expected)
			}
		})
	}
}

func TestWriter_Prepare(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewWriter(tmpDir)

	paths, err := w.Prepare("https://www.example.com")
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(paths.Dir); os.IsNotExist(err) {
		t.Error("output directory was not created")
	}

	// Verify all paths are populated
	if paths.CollyHTMLPath == "" {
		t.Error("CollyHTMLPath is empty")
	}
	if paths.ChromeHTMLPath == "" {
		t.Error("ChromeHTMLPath is empty")
	}
	if paths.ScreenshotPath == "" {
		t.Error("ScreenshotPath is empty")
	}
	if paths.JSONPath == "" {
		t.Error("JSONPath is empty")
	}
}

func TestWriter_SaveFile(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewWriter(tmpDir)

	path := filepath.Join(tmpDir, "test.html")
	data := []byte("<html><body>hello</body></html>")

	if err := w.SaveFile(path, data); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	read, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(read) != string(data) {
		t.Errorf("content mismatch: got %q, want %q", string(read), string(data))
	}
}

func TestWriter_SaveFile_NilData(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewWriter(tmpDir)

	path := filepath.Join(tmpDir, "nil.html")
	if err := w.SaveFile(path, nil); err != nil {
		t.Fatalf("SaveFile(nil) should not error, got: %v", err)
	}

	// File should NOT have been created
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should not exist for nil data")
	}
}

func TestWriter_SaveJSON(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewWriter(tmpDir)

	path := filepath.Join(tmpDir, "result.json")
	data := map[string]string{"key": "value"}

	if err := w.SaveJSON(path, data); err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}

	read, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(read) == 0 {
		t.Error("JSON file is empty")
	}
}

func TestWriter_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewWriter(tmpDir)

	paths, err := w.Prepare("https://example.com")
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	// Create a file inside
	_ = w.SaveFile(filepath.Join(paths.Dir, "dummy.html"), []byte("test"))

	if err := w.Cleanup(paths.Dir); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	if _, err := os.Stat(paths.Dir); !os.IsNotExist(err) {
		t.Error("directory should have been cleaned up")
	}
}
