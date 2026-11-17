package repomanager

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type githubTestTransport func(*http.Request) (*http.Response, error)

func (f githubTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGitHubClientReleaseAuthentication(t *testing.T) {
	for _, token := range []string{"", "test-token"} {
		t.Run(token, func(t *testing.T) {
			t.Setenv("GITHUB_TOKEN", token)
			previous := http.DefaultTransport
			t.Cleanup(func() { http.DefaultTransport = previous })
			http.DefaultTransport = githubTestTransport(func(r *http.Request) (*http.Response, error) {
				expected := ""
				if token != "" {
					expected = "Bearer " + token
				}
				if got := r.Header.Get("Authorization"); got != expected {
					t.Errorf("authorization = %q, want %q", got, expected)
				}
				if r.URL.Path != "/repos/owner/repo/releases/latest" {
					t.Errorf("path = %q", r.URL.Path)
				}
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"tag_name":"v1.2.3","assets":[{"name":"plugin.tar.gz","browser_download_url":"https://example.org/plugin.tar.gz","size":42}]}`))}, nil
			})
			client, err := NewGitHubClient()
			if err != nil {
				t.Fatal(err)
			}
			releases, err := client.GetReleases("owner", "repo", "latest")
			if err != nil {
				t.Fatal(err)
			}
			if len(releases) != 1 || releases[0].TagName != "v1.2.3" || len(releases[0].Assets) != 1 || releases[0].Assets[0].Size != 42 {
				t.Fatalf("unexpected releases: %+v", releases)
			}
		})
	}
}
