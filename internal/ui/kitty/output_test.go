package kitty

import (
	"io"
	"os"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func captureTmuxOutput(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	w := &TmuxOutput{File: f}
	if w.Fd() != f.Fd() {
		t.Fatal("output wrapper lost terminal descriptor")
	}
	n, err := w.Write([]byte(content))
	if err != nil || n != len(content) {
		t.Fatalf("Write = %d, %v; want %d, nil", n, err, len(content))
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestTmuxOutput(t *testing.T) {
	plain := "\x1b[2Jordinary text"
	if got := captureTmuxOutput(t, plain); got != plain {
		t.Fatalf("plain output changed: %q", got)
	}
	frame := "\r" + UnicodePlaceholderRow(42, 1, 3) + " new lyric\r\n"
	want := ansi.SetModeSynchronizedOutput + frame + ansi.ResetModeSynchronizedOutput
	if got := captureTmuxOutput(t, frame); got != want {
		t.Fatalf("frame not synchronized as one write: %q", got)
	}
}

func TestTmuxOutputWriteStringUsesSynchronization(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	frame := UnicodePlaceholderRow(42, 1, 3)
	n, err := io.WriteString(&TmuxOutput{File: f}, frame)
	if err != nil || n != len(frame) {
		t.Fatalf("WriteString = %d, %v", n, err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), ansi.SetModeSynchronizedOutput+frame+ansi.ResetModeSynchronizedOutput; got != want {
		t.Fatalf("WriteString bypassed synchronization: %q", got)
	}
}
