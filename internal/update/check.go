// Package update implements a lightweight, best-effort check for newer
// git-sweep releases on GitHub. It caches the latest known tag locally so
// the network is hit at most once per cache TTL, and it never blocks the
// caller for more than a short timeout.
//
// The package is intentionally permissive: any error (no network, missing
// cache directory, malformed JSON) is treated as a no-op so a transient
// failure can never stop the user's actual workflow.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultCacheTTL = 24 * time.Hour
	defaultTimeout  = 1500 * time.Millisecond
	cacheFileName   = "version-check.json"
)

// Result captures what the check learned about the latest release.
// LatestTag is the release tag (e.g., "v1.2.3") and is empty when unknown.
type Result struct {
	LatestTag string    `json:"latest_tag"`
	CheckedAt time.Time `json:"checked_at"`
}

// Options configures a Check call. Zero values pick sensible defaults.
type Options struct {
	// Repo is the owner/name slug, e.g. "jmelosegui/git-sweep". Required.
	Repo string

	// CacheDir overrides the on-disk cache directory. When empty the
	// OS-appropriate default is used (DefaultCacheDir).
	CacheDir string

	// CacheTTL controls how long a successful check is cached. Zero means
	// 24h.
	CacheTTL time.Duration

	// HTTPClient overrides the HTTP client. When nil a client with a
	// short timeout is used.
	HTTPClient *http.Client

	// Now overrides the clock for tests.
	Now func() time.Time
}

// Check returns the latest known release tag, using the on-disk cache when
// it is still fresh and otherwise consulting the GitHub releases API.
//
// Any error -- a missing cache file, a corrupt cache, a network failure,
// a non-2xx HTTP status -- is intentionally swallowed: the returned Result
// will have an empty LatestTag and the caller should treat that as "no
// information available right now."
func Check(ctx context.Context, opts Options) Result {
	if opts.Repo == "" {
		return Result{}
	}
	now := time.Now
	if opts.Now != nil {
		now = opts.Now
	}
	ttl := opts.CacheTTL
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}
	dir := opts.CacheDir
	if dir == "" {
		dir = DefaultCacheDir()
	}

	cachePath := filepath.Join(dir, cacheFileName)

	if cached, ok := readCache(cachePath); ok {
		if now().Sub(cached.CheckedAt) < ttl {
			return cached
		}
	}

	tag, err := fetchLatestTag(ctx, opts.Repo, opts.HTTPClient)
	if err != nil || tag == "" {
		return Result{}
	}

	res := Result{LatestTag: tag, CheckedAt: now()}
	_ = writeCache(cachePath, res)
	return res
}

// DefaultCacheDir returns the OS-appropriate cache directory for git-sweep.
//
// On Linux/macOS this honours XDG_CACHE_HOME, falling back to ~/.cache.
// On Windows it uses %LOCALAPPDATA%. When neither is resolvable an empty
// string is returned and callers should treat the cache as unavailable.
func DefaultCacheDir() string {
	if runtime.GOOS == "windows" {
		if d := os.Getenv("LOCALAPPDATA"); d != "" {
			return filepath.Join(d, "git-sweep")
		}
	}
	if d := os.Getenv("XDG_CACHE_HOME"); d != "" {
		return filepath.Join(d, "git-sweep")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".cache", "git-sweep")
	}
	return ""
}

func readCache(path string) (Result, bool) {
	if path == "" {
		return Result{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{}, false
	}
	var r Result
	if err := json.Unmarshal(data, &r); err != nil {
		return Result{}, false
	}
	if r.LatestTag == "" {
		return Result{}, false
	}
	return r, true
}

func writeCache(path string, r Result) error {
	if path == "" {
		return errors.New("empty cache path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func fetchLatestTag(ctx context.Context, repo string, client *http.Client) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "git-sweep-update-check")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return "", err
	}
	return body.TagName, nil
}

// IsNewer reports whether `latest` is strictly greater than `current`
// according to semver-with-leading-v rules used by our release tags.
// Empty inputs and the development sentinel "v0.0.0-dev" return false so
// the caller never nags during local builds.
func IsNewer(latest, current string) bool {
	if latest == "" || current == "" {
		return false
	}
	if current == "v0.0.0-dev" {
		return false
	}
	return compareSemver(latest, current) > 0
}

// compareSemver returns -1, 0, 1. It accepts an optional leading "v" and
// supports the subset of SemVer 2.0.0 that our release tags use:
// MAJOR.MINOR.PATCH with an optional "-PRERELEASE" suffix made up of
// dot-separated alphanumeric identifiers. Build metadata (+...) is
// ignored. Anything that fails to parse is treated as zero so two
// unparseable inputs compare equal and a parseable one always wins over
// garbage.
func compareSemver(a, b string) int {
	aMain, aPre := splitSemver(a)
	bMain, bPre := splitSemver(b)
	if c := compareNumericTriple(aMain, bMain); c != 0 {
		return c
	}
	// Per SemVer 2.0.0: a version without a pre-release suffix has higher
	// precedence than one with.
	switch {
	case aPre == "" && bPre == "":
		return 0
	case aPre == "":
		return 1
	case bPre == "":
		return -1
	default:
		return comparePrerelease(aPre, bPre)
	}
}

func splitSemver(v string) ([3]int, string) {
	v = strings.TrimPrefix(v, "v")
	// Drop build metadata.
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	main, pre := v, ""
	if i := strings.IndexByte(v, '-'); i >= 0 {
		main, pre = v[:i], v[i+1:]
	}
	parts := strings.SplitN(main, ".", 3)
	var nums [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return [3]int{}, ""
		}
		nums[i] = n
	}
	return nums, pre
}

func compareNumericTriple(a, b [3]int) int {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func comparePrerelease(a, b string) int {
	aIDs := strings.Split(a, ".")
	bIDs := strings.Split(b, ".")
	n := len(aIDs)
	if len(bIDs) < n {
		n = len(bIDs)
	}
	for i := 0; i < n; i++ {
		if c := compareIdentifier(aIDs[i], bIDs[i]); c != 0 {
			return c
		}
	}
	switch {
	case len(aIDs) < len(bIDs):
		return -1
	case len(aIDs) > len(bIDs):
		return 1
	default:
		return 0
	}
}

func compareIdentifier(a, b string) int {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)
	switch {
	case aErr == nil && bErr == nil:
		switch {
		case an < bn:
			return -1
		case an > bn:
			return 1
		default:
			return 0
		}
	case aErr == nil:
		// Numeric identifiers have lower precedence than alphanumeric.
		return -1
	case bErr == nil:
		return 1
	default:
		return strings.Compare(a, b)
	}
}
