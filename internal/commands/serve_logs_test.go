package commands

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestServeHelperProcess runs the real serve command as a subprocess of
// TestServeLogLinesCarryTheLifPrefix: serve blocks until it is killed, so it
// cannot run inside the test process.
func TestServeHelperProcess(t *testing.T) {
	if os.Getenv("LIF_SERVE_HELPER") != "1" {
		t.Skip("only runs as a subprocess of TestServeLogLinesCarryTheLifPrefix")
	}
	// TestMain builds ./cmd/lif relative to this package's directory, with
	// the caller's HOME and Go environment, so the subprocess moves into the
	// project and onto its own HOME only after that has run.
	if err := os.Chdir(os.Getenv("LIF_SERVE_DIR")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", os.Getenv("LIF_SERVE_HOME"))
	root := NewRootCmd()
	RegisterServeCmd(root)
	root.SetArgs([]string{"--lif-path", os.Getenv("LIF_SERVE_STORE"), "serve", "--port", os.Getenv("LIF_SERVE_PORT")})
	Execute(root)
}

// syncBuffer lets the test read what the subprocess has printed so far.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// AC5: the serve log lines carry the prefix [lif] instead of [beans], and the
// GraphQL playground is titled for the product.
func TestServeLogLinesCarryTheLifPrefix(t *testing.T) {
	project := t.TempDir()
	store := filepath.Join(project, ".lif")
	// A leftover in-store worktrees directory makes serve print its warning.
	if err := os.MkdirAll(filepath.Join(store, ".worktrees", "old"), 0755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", project, "init", "-q").CombinedOutput(); err != nil {
		t.Skipf("git init: %v\n%s", err, out)
	}
	// A GitHub remote plus a stand-in gh CLI on PATH makes serve report the
	// forge it detected.
	if out, err := exec.Command("git", "-C", project, "remote", "add", "origin", "https://github.com/o/r.git").CombinedOutput(); err != nil {
		t.Skipf("git remote add: %v\n%s", err, out)
	}
	stubDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(stubDir, "gh"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	var out syncBuffer
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestServeHelperProcess$")
	cmd.Env = append(os.Environ(),
		"PATH="+stubDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"LIF_SERVE_DIR="+project,
		"LIF_SERVE_HOME="+t.TempDir(),
		"LIF_SERVE_HELPER=1",
		"LIF_SERVE_STORE="+store,
		fmt.Sprintf("LIF_SERVE_PORT=%d", port),
	)
	cmd.Stdout, cmd.Stderr = pw, pw
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pw.Close()
	go func() { _, _ = io.Copy(&out, pr) }()
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	deadline := time.Now().Add(30 * time.Second)
	for !strings.Contains(out.String(), "Allowed origins") {
		if time.Now().After(deadline) {
			t.Fatalf("serve did not report its startup lines in time:\n%s", out.String())
		}
		time.Sleep(50 * time.Millisecond)
	}

	got := out.String()
	for _, want := range []string{
		"[lif] worktrees directory: ",
		"[lif] WARNING: found old worktrees in ",
		"[lif] detected forge: ",
		fmt.Sprintf("[lif] Starting server at http://localhost:%d/", port),
		fmt.Sprintf("[lif] GraphQL Playground: http://localhost:%d/playground", port),
		"[lif] Allowed origins: ",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("serve output lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "[beans]") {
		t.Errorf("serve output still carries the prefix [beans]:\n%s", got)
	}

	// The startup line prints just before the listener accepts, so poll.
	var resp *http.Response
	for {
		resp, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/playground", port))
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET /playground: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	defer resp.Body.Close()
	page, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(page), "<title>Lif GraphQL</title>") {
		t.Errorf("playground page lacks the title %q:\n%.400s", "Lif GraphQL", page)
	}
}
