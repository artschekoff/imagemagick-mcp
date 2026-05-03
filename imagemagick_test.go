package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func requireMagick(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("magick"); err != nil {
		t.Skip("magick not in PATH — skipping integration test")
	}
}

func TestRatioClose(t *testing.T) {
	cases := []struct {
		a, b float64
		want bool
	}{
		{1.0, 1.0, true},
		{1.0, 1.005, true},
		{1.0, 1.02, false},
		{16.0 / 9.0, 16.0 / 9.0, true},
		{1.333, 1.334, true},
	}
	for _, c := range cases {
		if got := ratioClose(c.a, c.b, 0.01); got != c.want {
			t.Errorf("ratioClose(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestModeValidation(t *testing.T) {
	requireMagick(t)
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "100x100", "xc:white", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}
	dst := filepath.Join(tmp, "out.png")
	err := cropResizeBlurBg(src, dst, 200, 200, "invalid", 30)
	if err == nil {
		t.Fatal("expected error for invalid mode, got nil")
	}
}

func TestCropResizeBlurBg_SameRatio(t *testing.T) {
	requireMagick(t)
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "200x200", "xc:blue", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}
	dst := filepath.Join(tmp, "out.png")
	if err := cropResizeBlurBg(src, dst, 100, 100, "blur", 30); err != nil {
		t.Fatalf("cropResizeBlurBg: %v", err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	w, h, err := identifySize(dst)
	if err != nil {
		t.Fatalf("identifySize: %v", err)
	}
	if w != 100 || h != 100 {
		t.Errorf("expected 100x100, got %dx%d", w, h)
	}
}

func TestCropResizeBlurBg_BlurMode(t *testing.T) {
	requireMagick(t)
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "200x100", "xc:red", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}
	dst := filepath.Join(tmp, "out.png")
	if err := cropResizeBlurBg(src, dst, 100, 100, "blur", 30); err != nil {
		t.Fatalf("cropResizeBlurBg blur: %v", err)
	}
	w, h, err := identifySize(dst)
	if err != nil {
		t.Fatalf("identifySize: %v", err)
	}
	if w != 100 || h != 100 {
		t.Errorf("expected 100x100, got %dx%d", w, h)
	}
}

func TestCropResizeBlurBg_ContainMode(t *testing.T) {
	requireMagick(t)
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "200x100", "xc:green", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}
	dst := filepath.Join(tmp, "out.png")
	if err := cropResizeBlurBg(src, dst, 100, 100, "contain", 30); err != nil {
		t.Fatalf("cropResizeBlurBg contain: %v", err)
	}
	w, h, err := identifySize(dst)
	if err != nil {
		t.Fatalf("identifySize: %v", err)
	}
	if w != 100 || h != 100 {
		t.Errorf("expected 100x100, got %dx%d", w, h)
	}
}

func TestCropResizeBlurBg_CoverMode(t *testing.T) {
	requireMagick(t)
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "200x100", "xc:yellow", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}
	dst := filepath.Join(tmp, "out.png")
	if err := cropResizeBlurBg(src, dst, 100, 100, "cover", 30); err != nil {
		t.Fatalf("cropResizeBlurBg cover: %v", err)
	}
	w, h, err := identifySize(dst)
	if err != nil {
		t.Fatalf("identifySize: %v", err)
	}
	if w != 100 || h != 100 {
		t.Errorf("expected 100x100, got %dx%d", w, h)
	}
}

func TestCropResizeBlurBg_MissingInput(t *testing.T) {
	err := cropResizeBlurBg("/nonexistent/path.png", "/tmp/out.png", 100, 100, "blur", 30)
	if err == nil {
		t.Fatal("expected error for missing input, got nil")
	}
}

func TestCropResizeBlurBg_CreatesOutputDir(t *testing.T) {
	requireMagick(t)
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.png")
	if err := exec.Command("magick", "-size", "50x50", "xc:white", src).Run(); err != nil {
		t.Fatalf("create test image: %v", err)
	}
	dst := filepath.Join(tmp, "subdir", "nested", "out.png")
	if err := cropResizeBlurBg(src, dst, 50, 50, "blur", 30); err != nil {
		t.Fatalf("cropResizeBlurBg: %v", err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("output not found at nested path: %v", err)
	}
}
