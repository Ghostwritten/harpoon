package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReadImageList(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name    string
		content string
		want    []string
		wantErr bool
	}{
		{
			name:    "empty file",
			content: "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "only blank lines",
			content: "\n\n  \n\t\n",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "only comments",
			content: "# a\n# b\n",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "comments and blank lines only",
			content: "# comment\n\n  \n# another\n",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "single image line",
			content: "nginx:latest\n",
			want:    []string{"nginx:latest"},
			wantErr: false,
		},
		{
			name:    "multiple images",
			content: "nginx:latest\nredis:alpine\nbusybox:1.36\n",
			want:    []string{"nginx:latest", "redis:alpine", "busybox:1.36"},
			wantErr: false,
		},
		{
			name:    "leading and trailing spaces trimmed",
			content: "  nginx:latest  \n\t redis:alpine \t\n",
			want:    []string{"nginx:latest", "redis:alpine"},
			wantErr: false,
		},
		{
			name:    "comments skipped",
			content: "# list\nnginx:latest\n# inline not supported\nredis:alpine\n",
			want:    []string{"nginx:latest", "redis:alpine"},
			wantErr: false,
		},
		{
			name:    "line starting with # skipped",
			content: "nginx:latest\n# comment\nredis:alpine\n",
			want:    []string{"nginx:latest", "redis:alpine"},
			wantErr: false,
		},
		{
			name:    "duplicates removed",
			content: "nginx:latest\nredis:alpine\nnginx:latest\nbusybox:1.36\nredis:alpine\n",
			want:    []string{"nginx:latest", "redis:alpine", "busybox:1.36"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, "images.txt")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			got, err := readImageList(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("readImageList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readImageList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadImageList_NoSuchFile(t *testing.T) {
	_, err := readImageList(filepath.Join(t.TempDir(), "nonexistent.txt"))
	if err == nil {
		t.Error("readImageList(nonexistent) expected error")
	}
}

func TestReadImageList_Stdin(t *testing.T) {
	// Temporarily replace stdin with a reader
	content := "nginx:latest\nredis:alpine\n"
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		w.Write([]byte(content))
		w.Close()
	}()

	got, err := readImageList("-")
	if err != nil {
		t.Fatalf("readImageList(\"-\") error = %v", err)
	}
	want := []string{"nginx:latest", "redis:alpine"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("readImageList(\"-\") = %v, want %v", got, want)
	}
}

func TestReadImageList_Stdin_Empty(t *testing.T) {
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		w.Write([]byte("# comment only\n\n"))
		w.Close()
	}()

	_, err = readImageList("-")
	if err == nil {
		t.Error("readImageList(\"-\") with empty content expected error")
	}
	if !strings.Contains(err.Error(), "no images found") {
		t.Errorf("expected 'no images found' in error, got: %v", err)
	}
}


func TestValidateImageList(t *testing.T) {
	tests := []struct {
		name      string
		images    []string
		wantLine  int
		wantErr   bool
		errContains string
	}{
		{
			name:     "all valid",
			images:   []string{"nginx:latest", "redis:alpine"},
			wantLine: 0,
			wantErr:  false,
		},
		{
			name:        "empty string invalid",
			images:      []string{"nginx:latest", "", "redis:alpine"},
			wantLine:    2,
			wantErr:     true,
			errContains: "image string cannot be empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lineNum, invalid, err := validateImageList(tt.images)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateImageList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if lineNum != tt.wantLine {
					t.Errorf("validateImageList() lineNum = %d, want %d", lineNum, tt.wantLine)
				}
				if tt.errContains != "" && err != nil && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateImageList() error = %v, want to contain %q", err, tt.errContains)
				}
				_ = invalid // used in error message
			}
		})
	}
}


func TestDedupeImages(t *testing.T) {
	tests := []struct {
		name   string
		images []string
		want   []string
	}{
		{"empty", []string{}, []string{}},
		{"no dupes", []string{"a:1", "b:2"}, []string{"a:1", "b:2"}},
		{"dupes", []string{"a:1", "b:2", "a:1", "c:3", "b:2"}, []string{"a:1", "b:2", "c:3"}},
		{"all same", []string{"a:1", "a:1", "a:1"}, []string{"a:1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedupeImages(tt.images)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dedupeImages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeExtraArgs(t *testing.T) {
	tests := []struct {
		name        string
		configArgs  []string
		passthrough []string
		want        []string
	}{
		{"nil both", nil, nil, []string{}},
		{"empty both", []string{}, []string{}, []string{}},
		{"config only", []string{"--a", "--b"}, nil, []string{"--a", "--b"}},
		{"passthrough only", nil, []string{"--x", "--y"}, []string{"--x", "--y"}},
		{"both, config first", []string{"--a"}, []string{"--x"}, []string{"--a", "--x"}},
		{"both multiple", []string{"--a", "--b"}, []string{"--x", "--y"}, []string{"--a", "--b", "--x", "--y"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeExtraArgs(tt.configArgs, tt.passthrough)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
