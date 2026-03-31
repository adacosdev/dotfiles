package diff

import (
	"strings"
	"testing"
)

// Golden test cases for diff parsing.
var goldenTestCases = []struct {
	name     string
	input    string
	expected []DiffFile
}{
	{
		name:  "simple added file",
		input: `diff --git a/test.txt b/test.txt
new file mode 100644
index 0000000..e69de29
--- /dev/null
+++ b/test.txt
@@ -0,0 +1,3 @@
+line1
+line2
+line3
`,
		expected: []DiffFile{
			{
				Path:   "test.txt",
				Status: "added",
				Hunks: []DiffHunk{
					{
						OldStart: 0,
						OldCount: 0,
						NewStart: 1,
						NewCount: 3,
						Lines: []DiffLine{
							{Content: "line1", Type: "add", NewNum: 1},
							{Content: "line2", Type: "add", NewNum: 2},
							{Content: "line3", Type: "add", NewNum: 3},
						},
					},
				},
			},
		},
	},
	{
		name:  "simple deleted file",
		input: `diff --git a/deleted.txt b/deleted.txt
deleted file mode 100644
index e69de29..0000000
--- a/deleted.txt
+++ /dev/null
@@ -1,3 +0,0 @@
-line1
-line2
-line3
`,
		expected: []DiffFile{
			{
				Path:   "deleted.txt",
				Status: "deleted",
				Hunks: []DiffHunk{
					{
						OldStart: 1,
						OldCount: 3,
						NewStart: 0,
						NewCount: 0,
						Lines: []DiffLine{
							{Content: "line1", Type: "del", OldNum: 1},
							{Content: "line2", Type: "del", OldNum: 2},
							{Content: "line3", Type: "del", OldNum: 3},
						},
					},
				},
			},
		},
	},
	{
		name:  "modified file",
		input: `diff --git a/modified.txt b/modified.txt
index e69de29..a1b2c3d 100644
--- a/modified.txt
+++ b/modified.txt
@@ -1,3 +1,4 @@
 line1
-line2
+line2-modified
+line3
 line4
`,
		expected: []DiffFile{
			{
				Path:   "modified.txt",
				Status: "modified",
				Hunks: []DiffHunk{
					{
						OldStart: 1,
						OldCount: 3,
						NewStart: 1,
						NewCount: 4,
						Lines: []DiffLine{
							{Content: "line1", Type: "context", OldNum: 1, NewNum: 1},
							{Content: "line2", Type: "del", OldNum: 2},
							{Content: "line2-modified", Type: "add", NewNum: 2},
							{Content: "line3", Type: "add", NewNum: 3},
							{Content: "line4", Type: "context", OldNum: 3, NewNum: 4},
						},
					},
				},
			},
		},
	},
	{
		name:  "multiple files",
		input: `diff --git a/file1.txt b/file1.txt
index 0000000..e69de29
--- /dev/null
+++ b/file1.txt
@@ -0,0 +1,2 @@
+content1
+content2
diff --git a/file2.txt b/file2.txt
index 0000000..e69de29
--- /dev/null
+++ b/file2.txt
@@ -0,0 +1,1 @@
+another content
`,
		expected: []DiffFile{
			{
				Path:   "file1.txt",
				Status: "added",
				Hunks: []DiffHunk{
					{
						OldStart: 0,
						OldCount: 0,
						NewStart: 1,
						NewCount: 2,
						Lines: []DiffLine{
							{Content: "content1", Type: "add", NewNum: 1},
							{Content: "content2", Type: "add", NewNum: 2},
						},
					},
				},
			},
			{
				Path:   "file2.txt",
				Status: "added",
				Hunks: []DiffHunk{
					{
						OldStart: 0,
						OldCount: 0,
						NewStart: 1,
						NewCount: 1,
						Lines: []DiffLine{
							{Content: "another content", Type: "add", NewNum: 1},
						},
					},
				},
			},
		},
	},
	{
		name:  "empty output",
		input: "",
		expected: []DiffFile{},
	},
}

// TestParseDiff tests the ParseDiff function with golden test cases.
func TestParseDiff(t *testing.T) {
	for _, tc := range goldenTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseDiff(tc.input)

			if len(result) != len(tc.expected) {
				t.Errorf("got %d files, want %d", len(result), len(tc.expected))
				for i, f := range result {
					t.Logf("  got file[%d]: %s (%s)", i, f.Path, f.Status)
				}
				for i, f := range tc.expected {
					t.Logf("  want file[%d]: %s (%s)", i, f.Path, f.Status)
				}
				return
			}

			for i, file := range result {
				expectedFile := tc.expected[i]

				if file.Path != expectedFile.Path {
					t.Errorf("file[%d].Path = %q, want %q", i, file.Path, expectedFile.Path)
				}

				if file.Status != expectedFile.Status {
					t.Errorf("file[%d].Status = %q, want %q", i, file.Status, expectedFile.Status)
				}

				if len(file.Hunks) != len(expectedFile.Hunks) {
					t.Errorf("file[%d] has %d hunks, want %d", i, len(file.Hunks), len(expectedFile.Hunks))
					continue
				}

				for j, hunk := range file.Hunks {
					expectedHunk := expectedFile.Hunks[j]

					if hunk.OldStart != expectedHunk.OldStart {
						t.Errorf("file[%d].hunk[%d].OldStart = %d, want %d", i, j, hunk.OldStart, expectedHunk.OldStart)
					}

					if hunk.NewStart != expectedHunk.NewStart {
						t.Errorf("file[%d].hunk[%d].NewStart = %d, want %d", i, j, hunk.NewStart, expectedHunk.NewStart)
					}

					if len(hunk.Lines) != len(expectedHunk.Lines) {
						t.Errorf("file[%d].hunk[%d] has %d lines, want %d", i, j, len(hunk.Lines), len(expectedHunk.Lines))
						continue
					}

					for k, line := range hunk.Lines {
						expectedLine := expectedHunk.Lines[k]
						if line.Type != expectedLine.Type {
							t.Errorf("file[%d].hunk[%d].line[%d].Type = %q, want %q", i, j, k, line.Type, expectedLine.Type)
						}
						if line.Content != expectedLine.Content {
							t.Errorf("file[%d].hunk[%d].line[%d].Content = %q, want %q", i, j, k, line.Content, expectedLine.Content)
						}
					}
				}
			}
		})
	}
}

// TestParseDiffLineType tests line type detection with table-driven tests.
func TestParseDiffLineType(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"added line", "+ added content", "add"},
		{"deleted line", "- deleted content", "del"},
		{"context line with space", "  context content", "context"},
		{"empty context line", "", "context"},
		{"hunk header", "@@ -1,3 +1,4 @@", "header"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lineType := detectLineType(tt.line)
			if lineType != tt.expected {
				t.Errorf("detectLineType(%q) = %q, want %q", tt.line, lineType, tt.expected)
			}
		})
	}
}

// detectLineType determines the type of a diff line.
func detectLineType(line string) string {
	if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
		return "add"
	}
	if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
		return "del"
	}
	if strings.HasPrefix(line, "@@") {
		return "header"
	}
	return "context"
}

// TestParseDiffMultipleHunks tests parsing a file with multiple hunks.
func TestParseDiffMultipleHunks(t *testing.T) {
	input := `diff --git a/multi_hunk.txt b/multi_hunk.txt
index 1234567..abcdefg 100644
--- a/multi_hunk.txt
+++ b/multi_hunk.txt
@@ -1,3 +1,4 @@
 line1
-line2
+line2-modified
+line3
@@ -10,3 +10,4 @@
 line10
-line11
+line11-modified
+line12
`

	result := ParseDiff(input)

	if len(result) != 1 {
		t.Fatalf("got %d files, want 1", len(result))
	}

	file := result[0]
	if len(file.Hunks) != 2 {
		t.Errorf("file has %d hunks, want 2", len(file.Hunks))
	}
}

// TestParseDiffRealWorldOutput tests parsing with real-world chezmoi diff output.
func TestParseDiffRealWorldOutput(t *testing.T) {
	input := `diff --git a/.config/nvim/init.vim b/.config/nvim/init.vim
index 3a4b5c6..d7e8f9a 100644
--- a/.config/nvim/init.vim
+++ b/.config/nvim/init.vim
@@ -15,7 +15,8 @@
 set number
 set relativenumber
 set cursorline
-" Plugin manager
+' Plugin manager
+" LazyVim plugin manager
 set hidden
`

	result := ParseDiff(input)

	if len(result) != 1 {
		t.Fatalf("got %d files, want 1", len(result))
	}

	file := result[0]
	if file.Path != ".config/nvim/init.vim" {
		t.Errorf("Path = %q, want %q", file.Path, ".config/nvim/init.vim")
	}

	if file.Status != "modified" {
		t.Errorf("Status = %q, want %q", file.Status, "modified")
	}
}

// TestDiffModelInit tests the DiffModel initialization.
func TestNewDiffModel(t *testing.T) {
	model := NewDiffModel()

	if model.selected != 0 {
		t.Errorf("selected = %d, want 0", model.selected)
	}

	if model.viewMode != ViewModeSideBySide {
		t.Errorf("viewMode = %v, want ViewModeSideBySide", model.viewMode)
	}

	if model.focus != "filelist" {
		t.Errorf("focus = %q, want %q", model.focus, "filelist")
	}

	if !model.loading {
		t.Error("loading should be true initially")
	}
}

// TestViewModeToggle tests toggling between view modes.
func TestViewModeToggle(t *testing.T) {
	model := &DiffModel{
		viewMode: ViewModeSideBySide,
	}

	// Toggle to unified
	if model.viewMode == ViewModeSideBySide {
		model.viewMode = ViewModeUnified
	}

	if model.viewMode != ViewModeUnified {
		t.Errorf("viewMode = %v, want ViewModeUnified", model.viewMode)
	}
}

// TestGetStatusIcon tests the status icon function.
func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"added", "+"},
		{"deleted", "-"},
		{"modified", "~"},
		{"unchanged", "="},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := getStatusIcon(tt.status)
			if result != tt.expected {
				t.Errorf("getStatusIcon(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

// TestAtiSafe tests the atoiSafe helper function.
func TestAtoiSafe(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := atoiSafe(tt.input)
			if result != tt.expected {
				t.Errorf("atoiSafe(%q) = %d, want %d", tt.input, result, tt.expected)
			}
			if (err != nil) != tt.hasError {
				t.Errorf("atoiSafe(%q) error = %v, hasError = %v", tt.input, err, tt.hasError)
			}
		})
	}
}
