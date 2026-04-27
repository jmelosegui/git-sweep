package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v1.0.1", "v1.0.0", true},
		{"v1.1.0", "v1.0.9", true},
		{"v2.0.0", "v1.99.99", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.0.1", false},
		{"v1.0.0", "v1.0.0-rc.1", true}, // release > pre-release
		{"v1.0.0-rc.2", "v1.0.0-rc.1", true},
		{"v1.0.0-rc.1", "v1.0.0", false}, // pre-release < release
		{"v1.0.0", "v0.0.0-dev", false},  // dev sentinel never nags
		{"", "v1.0.0", false},
		{"v1.0.0", "", false},
		{"1.0.1", "1.0.0", true}, // tolerant of missing v
	}
	for _, c := range cases {
		got := IsNewer(c.latest, c.current)
		if got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestCheck_UsesFreshCache(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC)

	// Seed cache with a value 1h old; TTL is 24h, so this should hit.
	cached := Result{LatestTag: "v9.9.9", CheckedAt: now.Add(-1 * time.Hour)}
	data, err := json.Marshal(cached)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, cacheFileName), data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Fail loudly if anything tries to call out to the network.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("unexpected HTTP request when cache is fresh")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	res := Check(context.Background(), Options{
		Repo:     "owner/repo",
		CacheDir: dir,
		Now:      func() time.Time { return now },
	})
	if res.LatestTag != "v9.9.9" {
		t.Fatalf("want cached tag v9.9.9, got %q", res.LatestTag)
	}
}

func TestCheck_RefreshesStaleCache(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC)

	stale := Result{LatestTag: "v0.0.1", CheckedAt: now.Add(-48 * time.Hour)}
	data, err := json.Marshal(stale)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, cacheFileName), data, 0o644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v2.0.0"}`))
	}))
	defer server.Close()

	res := checkAgainst(t, server, dir, now)
	if res.LatestTag != "v2.0.0" {
		t.Fatalf("want refreshed tag v2.0.0, got %q", res.LatestTag)
	}

	// Cache should now hold the fresh value.
	stored, ok := readCache(filepath.Join(dir, cacheFileName))
	if !ok {
		t.Fatal("expected cache to be written")
	}
	if stored.LatestTag != "v2.0.0" {
		t.Fatalf("want stored tag v2.0.0, got %q", stored.LatestTag)
	}
}

func TestCheck_NetworkErrorIsSilent(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	res := checkAgainst(t, server, dir, now)
	if res.LatestTag != "" {
		t.Fatalf("expected empty result on HTTP failure, got %q", res.LatestTag)
	}
	if _, err := os.Stat(filepath.Join(dir, cacheFileName)); !os.IsNotExist(err) {
		t.Fatalf("expected no cache write on failure, got err=%v", err)
	}
}

// checkAgainst runs Check with an HTTPClient that redirects api.github.com
// to the test server, so we can exercise the real fetch path without a
// network dependency.
func checkAgainst(t *testing.T, server *httptest.Server, dir string, now time.Time) Result {
	t.Helper()
	client := server.Client()
	client.Transport = &rewriteTransport{base: client.Transport, target: server.URL}
	return Check(context.Background(), Options{
		Repo:       "owner/repo",
		CacheDir:   dir,
		Now:        func() time.Time { return now },
		HTTPClient: client,
	})
}

type rewriteTransport struct {
	base   http.RoundTripper
	target string
}

func (r *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rewritten := req.Clone(req.Context())
	u := rewritten.URL
	// Replace the entire scheme+host so the request hits our test server.
	parsed, err := http.NewRequest(req.Method, r.target+u.RequestURI(), nil)
	if err != nil {
		return nil, err
	}
	rewritten.URL = parsed.URL
	rewritten.Host = parsed.URL.Host
	base := r.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(rewritten)
}
