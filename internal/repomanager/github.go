package repomanager

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// GitHubClient wraps the GitHub API client.
type GitHubClient struct {
	client *github.Client
}

// GitHubRelease represents a GitHub release with its assets.
type GitHubRelease struct {
	TagName     string
	PublishedAt time.Time
	URL         string
	Assets      []GitHubAsset
	Prerelease  bool
}

// GitHubAsset represents a release asset.
type GitHubAsset struct {
	Name        string
	DownloadURL string
	Size        int64
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient() *GitHubClient {
	var httpClient *http.Client

	// Use token if available (for higher rate limits)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = oauth2.NewClient(context.Background(), ts)
	}

	return &GitHubClient{
		client: github.NewClient(httpClient),
	}
}

// GetReleases resolves version specifier and returns matching releases.
func (c *GitHubClient) GetReleases(owner, repo, versionSpec string) ([]*GitHubRelease, error) {
	ctx := context.Background()

	switch versionSpec {
	case "latest":
		// Get latest non-prerelease
		release, _, err := c.client.Repositories.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest release: %w", err)
		}
		return []*GitHubRelease{convertRelease(release)}, nil

	case "all":
		// Get all releases (excluding prereleases)
		opts := &github.ListOptions{PerPage: 100}
		var allReleases []*GitHubRelease

		for {
			releases, resp, err := c.client.Repositories.ListReleases(ctx, owner, repo, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list releases: %w", err)
			}

			// Filter out prereleases
			for _, release := range releases {
				if !release.GetPrerelease() {
					allReleases = append(allReleases, convertRelease(release))
				}
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}

		return allReleases, nil

	default:
		// Specific version tag
		release, _, err := c.client.Repositories.GetReleaseByTag(ctx, owner, repo, versionSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to get release %s: %w", versionSpec, err)
		}
		return []*GitHubRelease{convertRelease(release)}, nil
	}
}

// convertRelease converts a GitHub API release to our GitHubRelease type.
func convertRelease(r *github.RepositoryRelease) *GitHubRelease {
	release := &GitHubRelease{
		TagName:     r.GetTagName(),
		PublishedAt: r.GetPublishedAt().Time,
		URL:         r.GetHTMLURL(),
		Prerelease:  r.GetPrerelease(),
		Assets:      make([]GitHubAsset, 0, len(r.Assets)),
	}

	for _, asset := range r.Assets {
		release.Assets = append(release.Assets, GitHubAsset{
			Name:        asset.GetName(),
			DownloadURL: asset.GetBrowserDownloadURL(),
			Size:        int64(asset.GetSize()),
		})
	}

	return release
}

// ParseGitHubRepo parses owner/repo format.
func ParseGitHubRepo(repo string) (owner, repoName string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format, expected owner/repo")
	}
	return parts[0], parts[1], nil
}
