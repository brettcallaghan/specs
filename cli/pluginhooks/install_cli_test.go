// Package pluginhooks tests the specs plugin's SessionStart hook,
// plugin/hooks/install-cli, by running it against a fake release host.
package pluginhooks

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

const pinned = "0.2.0"

// The plugin directory, relative to this package.
var pluginDir = filepath.Join("..", "..", "plugin")

// The tools the hook may use, linked into a directory that is the whole of
// its PATH, so nothing on the machine running the tests (a real `specs`)
// can leak into a case.
var tools = []string{
	"bash", "sh", "curl", "tar", "gzip", "sha256sum", "shasum", "awk", "sed",
	"mkdir", "rm", "mv", "chmod", "ln", "cat", "dirname", "uname", "printf",
	"unzip",
}

// fakeHost stands in for github.com's release downloads.
type fakeHost struct {
	srv *httptest.Server

	mu       sync.Mutex
	files    map[string][]byte // path under /brettcallaghan/specs/releases/download/ -> bytes
	requests []string
}

func newFakeHost(t *testing.T) *fakeHost {
	t.Helper()
	h := &fakeHost{files: map[string][]byte{}}
	h.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.requests = append(h.requests, r.URL.Path)
		body, ok := h.files[strings.TrimPrefix(r.URL.Path, "/brettcallaghan/specs/releases/download/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(h.srv.Close)
	return h
}

// publish adds release cli/v<version> for this platform: an archive holding
// a `specs` that reports version, and a SHA256SUMS covering it.
func (h *fakeHost) publish(t *testing.T, version string) string {
	t.Helper()
	name := fmt.Sprintf("specs_v%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
	data := archive(t, version)
	sum := sha256.Sum256(data)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.files["cli/v"+version+"/"+name] = data
	h.files["cli/v"+version+"/SHA256SUMS"] = []byte(hex.EncodeToString(sum[:]) + "  " + name + "\n")
	return name
}

// publishWindows adds release cli/v<version> for windows/amd64: a zip holding
// specs.exe, and a SHA256SUMS covering it.
func (h *fakeHost) publishWindows(t *testing.T, version string) {
	t.Helper()
	name := "specs_v" + version + "_windows_amd64.zip"
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.CreateHeader(&zip.FileHeader{Name: "specs.exe", Method: zip.Deflate})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(fakeSpecs(version))); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(buf.Bytes())
	h.mu.Lock()
	defer h.mu.Unlock()
	h.files["cli/v"+version+"/"+name] = buf.Bytes()
	h.files["cli/v"+version+"/SHA256SUMS"] = []byte(hex.EncodeToString(sum[:]) + "  " + name + "\n")
}

func (h *fakeHost) corrupt(version, name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := "cli/v" + version + "/" + name
	h.files[key] = append([]byte("tampered"), h.files[key]...)
}

func (h *fakeHost) seen() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.requests...)
}

// fakeSpecs is a `specs` that reports version the way a release build does.
func fakeSpecs(version string) string {
	return "#!/bin/sh\nprintf 'version: v%s\\nrevision: null\\n' '" + version + "'\n"
}

func archive(t *testing.T, version string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := []byte(fakeSpecs(version))
	if err := tw.WriteHeader(&tar.Header{Name: "specs", Mode: 0o755, Size: int64(len(body))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// env is one session's surroundings: a copy of the plugin pinning version,
// its data directory, the env file, and a PATH of the allowed tools only.
type env struct {
	t       *testing.T
	host    *fakeHost
	script  string
	data    string
	envFile string
	path    string // the PATH directory
	linkDir string
	extra   []string
}

func newEnv(t *testing.T, host *fakeHost, version string) *env {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the hook's tests run under a POSIX shell")
	}
	root := t.TempDir()
	plugin := filepath.Join(root, "plugin")
	mustMkdir(t, filepath.Join(plugin, "hooks"))
	mustMkdir(t, filepath.Join(plugin, ".claude-plugin"))
	src, err := os.ReadFile(filepath.Join(pluginDir, "hooks", "install-cli"))
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(plugin, "hooks", "install-cli")
	mustWrite(t, script, string(src), 0o755)
	mustWrite(t, filepath.Join(plugin, ".claude-plugin", "plugin.json"),
		`{"name": "specs", "version": "`+version+`", "author": {"name": "x", "version": "9.9.9"}}`, 0o644)

	e := &env{
		t: t, host: host, script: script,
		data:    filepath.Join(root, "data"),
		envFile: filepath.Join(root, "env"),
		path:    filepath.Join(root, "path"),
		linkDir: filepath.Join(root, "link"),
	}
	mustMkdir(t, e.path)
	mustMkdir(t, e.linkDir)
	mustWrite(t, e.envFile, "", 0o644)
	for _, tool := range tools {
		if p, err := exec.LookPath(tool); err == nil {
			if err := os.Symlink(p, filepath.Join(e.path, tool)); err != nil {
				t.Fatal(err)
			}
		}
	}
	return e
}

// run runs the hook once and returns its standard output.
func (e *env) run() string {
	e.t.Helper()
	cmd := exec.Command("bash", e.script)
	cmd.Env = append([]string{
		"PATH=" + e.path,
		"HOME=" + e.t.TempDir(),
		"CLAUDE_PLUGIN_DATA=" + e.data,
		"CLAUDE_ENV_FILE=" + e.envFile,
		"SPECS_CLI_DOWNLOAD_URL=" + e.host.srv.URL,
		"SPECS_CLI_LINK_DIR=" + e.linkDir,
	}, e.extra...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		e.t.Fatalf("the hook exited %v; it must always exit 0\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	return stdout.String()
}

func (e *env) installed() string {
	return filepath.Join(e.data, "bin", pinned, "specs")
}

// sessionSpecs resolves `specs` the way the session's Bash tool would:
// sourcing the env file, then searching PATH.
func (e *env) sessionSpecs() string {
	e.t.Helper()
	out, err := exec.Command("bash", "-c", `PATH="$1"; . "$2"; specs --version`, "_", e.path, e.envFile).CombinedOutput()
	if err != nil {
		e.t.Fatalf("specs --version in the session: %v: %s", err, out)
	}
	return string(out)
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}

func TestTheHookIsRegisteredForSessionStart(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(pluginDir, "hooks", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	const want = `bash "${CLAUDE_PLUGIN_ROOT}/hooks/install-cli"`
	for _, group := range doc.Hooks["SessionStart"] {
		for _, h := range group.Hooks {
			if h.Type == "command" && h.Command == want {
				return
			}
		}
	}
	t.Fatalf("hooks.json registers no SessionStart command %s", want)
}

func TestAFirstSessionInstallsThePinnedVersionAndPutsItFirstOnThePath(t *testing.T) {
	host := newFakeHost(t)
	host.publish(t, pinned)
	e := newEnv(t, host, pinned)

	if out := e.run(); out != "" {
		t.Fatalf("a clean install printed %q; it must print nothing", out)
	}
	if _, err := os.Stat(e.installed()); err != nil {
		t.Fatalf("not installed: %v", err)
	}
	if got := e.sessionSpecs(); !strings.Contains(got, "version: v"+pinned) {
		t.Fatalf("the session runs %q, not v%s", got, pinned)
	}
	if n := len(host.seen()); n != 2 {
		t.Fatalf("made %d requests (%v); want SHA256SUMS and the archive", n, host.seen())
	}
	if entries, _ := os.ReadDir(e.linkDir); len(entries) != 0 {
		t.Fatalf("linked into %s outside a cloud session", e.linkDir)
	}
}

func TestALaterSessionDownloadsNothing(t *testing.T) {
	host := newFakeHost(t)
	host.publish(t, pinned)
	e := newEnv(t, host, pinned)
	e.run()
	before := len(host.seen())

	mustWrite(t, e.envFile, "", 0o644)
	if out := e.run(); out != "" {
		t.Fatalf("printed %q", out)
	}
	if after := len(host.seen()); after != before {
		t.Fatalf("a provisioned version was downloaded again: %v", host.seen()[before:])
	}
	if got := e.sessionSpecs(); !strings.Contains(got, "v"+pinned) {
		t.Fatalf("the session runs %q", got)
	}
}

func TestTheRightVersionAlreadyOnThePathIsUsedAsItIs(t *testing.T) {
	host := newFakeHost(t)
	e := newEnv(t, host, pinned)
	mustWrite(t, filepath.Join(e.path, "specs"), fakeSpecs(pinned), 0o755)

	if out := e.run(); out != "" {
		t.Fatalf("printed %q", out)
	}
	if n := len(host.seen()); n != 0 {
		t.Fatalf("downloaded %v though the PATH already has v%s", host.seen(), pinned)
	}
	if body, _ := os.ReadFile(e.envFile); len(body) != 0 {
		t.Fatalf("changed the session's PATH: %q", body)
	}
}

func TestAnotherVersionOnThePathIsShadowedForTheSessionAndNamed(t *testing.T) {
	host := newFakeHost(t)
	host.publish(t, pinned)
	e := newEnv(t, host, pinned)
	theirs := filepath.Join(e.path, "specs")
	mustWrite(t, theirs, fakeSpecs("0.1.0"), 0o755)

	out := e.run()
	if !strings.HasPrefix(out, "specs plugin: ") || !strings.Contains(out, theirs+" (v0.1.0)") ||
		!strings.Contains(out, "SPECS_CLI_PROVISION=off") || strings.Count(out, "\n") != 1 {
		t.Fatalf("the notice is %q; want one line naming %s (v0.1.0) and the opt-out", out, theirs)
	}
	if got := e.sessionSpecs(); !strings.Contains(got, "v"+pinned) {
		t.Fatalf("the session runs %q, not the pinned v%s", got, pinned)
	}
	if body, _ := os.ReadFile(theirs); string(body) != fakeSpecs("0.1.0") {
		t.Fatal("the user's own specs was changed")
	}
}

func TestOptingOutLeavesEverythingAlone(t *testing.T) {
	host := newFakeHost(t)
	host.publish(t, pinned)
	e := newEnv(t, host, pinned)
	e.extra = []string{"SPECS_CLI_PROVISION=off"}

	if out := e.run(); out != "" {
		t.Fatalf("printed %q", out)
	}
	if len(host.seen()) != 0 {
		t.Fatalf("downloaded %v", host.seen())
	}
	if body, _ := os.ReadFile(e.envFile); len(body) != 0 {
		t.Fatalf("changed the session's PATH: %q", body)
	}
}

func TestACloudSessionLinksThePinnedVersionOverAnyOther(t *testing.T) {
	host := newFakeHost(t)
	host.publish(t, pinned)
	e := newEnv(t, host, pinned)
	e.extra = []string{"CLAUDE_CODE_REMOTE=true"}
	mustWrite(t, filepath.Join(e.path, "specs"), fakeSpecs("0.1.0"), 0o755)
	link := filepath.Join(e.linkDir, "specs")
	mustWrite(t, link, fakeSpecs("0.1.0"), 0o755)

	if out := e.run(); out != "" {
		t.Fatalf("a cloud session printed %q; it must print nothing", out)
	}
	target, err := os.Readlink(link)
	if err != nil || target != e.installed() {
		t.Fatalf("%s points at %q (%v), not %s", link, target, err, e.installed())
	}
}

func TestFailuresAreOneLineOfContextAndInstallNothing(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(*fakeHost)
		version string
		want    string
	}{
		{"no release", func(*fakeHost) {}, pinned, "HTTP 404"},
		{"checksum mismatch", func(h *fakeHost) {
			name := h.publish(t, pinned)
			h.corrupt(pinned, name)
		}, pinned, "checksum mismatch"},
		{"no version", func(*fakeHost) {}, "", "no single version"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			host := newFakeHost(t)
			c.setup(host)
			e := newEnv(t, host, c.version)

			out := e.run()
			if !strings.HasPrefix(out, "specs plugin: ") || !strings.Contains(out, c.want) ||
				strings.Count(out, "\n") != 1 {
				t.Fatalf("printed %q; want one `specs plugin:` line containing %q", out, c.want)
			}
			if _, err := os.Stat(e.installed()); err == nil {
				t.Fatal("installed a binary despite the failure")
			}
			if body, _ := os.ReadFile(e.envFile); len(body) != 0 {
				t.Fatalf("changed the session's PATH: %q", body)
			}
			if entries, _ := os.ReadDir(filepath.Join(e.data, "bin")); len(entries) > 0 {
				for _, en := range entries {
					if strings.HasPrefix(en.Name(), ".tmp-") {
						t.Fatalf("left %s behind", en.Name())
					}
				}
			}
		})
	}
}

func TestAWindowsZipIsUnpackedWhereTarCannotReadOne(t *testing.T) {
	host := newFakeHost(t)
	host.publishWindows(t, pinned)
	e := newEnv(t, host, pinned)
	// A Git Bash uname, and a GNU-like tar that cannot read a zip.
	for name, body := range map[string]string{
		"uname": "#!/bin/sh\ncase \"$1\" in -s) echo MINGW64_NT-10.0-19045 ;; -m) echo x86_64 ;; esac\n",
		"tar":   "#!/bin/sh\necho 'tar: This does not look like a tar archive' >&2\nexit 2\n",
	} {
		p := filepath.Join(e.path, name)
		_ = os.Remove(p)
		mustWrite(t, p, body, 0o755)
	}

	if out := e.run(); out != "" {
		t.Fatalf("printed %q", out)
	}
	exe := filepath.Join(e.data, "bin", pinned, "specs.exe")
	if body, err := os.ReadFile(exe); err != nil || string(body) != fakeSpecs(pinned) {
		t.Fatalf("specs.exe not installed from the zip: %v", err)
	}
}

func TestAFailureTellsTheAgentNotToRunAnotherVersionLeftOnThePath(t *testing.T) {
	host := newFakeHost(t)
	e := newEnv(t, host, pinned)
	theirs := filepath.Join(e.path, "specs")
	mustWrite(t, theirs, fakeSpecs("0.1.0"), 0o755)

	out := e.run()
	if !strings.HasPrefix(out, "specs plugin: ") || !strings.Contains(out, "HTTP 404") ||
		!strings.Contains(out, "Do not run "+theirs+" (v0.1.0)") || strings.Count(out, "\n") != 1 {
		t.Fatalf("printed %q; want one line naming the failure and warning off %s", out, theirs)
	}
}
