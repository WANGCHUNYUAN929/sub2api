//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type updateServiceCacheStub struct {
	data string
}

func (s *updateServiceCacheStub) GetUpdateInfo(context.Context) (string, error) {
	if s.data == "" {
		return "", errors.New("cache miss")
	}
	return s.data, nil
}

func (s *updateServiceCacheStub) SetUpdateInfo(_ context.Context, data string, _ time.Duration) error {
	s.data = data
	return nil
}

type updateServiceGitHubClientStub struct {
	release  *GitHubRelease
	releases map[string]*GitHubRelease
	fetched  []string
}

func (s *updateServiceGitHubClientStub) FetchLatestRelease(_ context.Context, repo string) (*GitHubRelease, error) {
	s.fetched = append(s.fetched, repo)
	if s.releases != nil {
		if release, ok := s.releases[repo]; ok {
			return release, nil
		}
		return nil, errors.New("release not found")
	}
	return s.release, nil
}

func (s *updateServiceGitHubClientStub) DownloadFile(context.Context, string, string, int64) error {
	panic("DownloadFile should not be called when no update is available")
}

func (s *updateServiceGitHubClientStub) FetchChecksumFile(context.Context, string) ([]byte, error) {
	panic("FetchChecksumFile should not be called when no update is available")
}

func TestUpdateServicePerformUpdateNoUpdateReturnsSentinel(t *testing.T) {
	svc := NewUpdateService(
		&updateServiceCacheStub{},
		&updateServiceGitHubClientStub{
			release: &GitHubRelease{
				TagName: "v0.1.132",
				Name:    "v0.1.132",
			},
		},
		"0.1.132",
		"release",
	)

	err := svc.PerformUpdate(context.Background())

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoUpdateAvailable))
	require.ErrorIs(t, err, ErrNoUpdateAvailable)
}

func TestUpdateServiceCheckUpdateUsesCustomUpdateRepoAndOfficialUpstream(t *testing.T) {
	t.Setenv(updateRepoEnvKey, "custom/sub2api")
	t.Setenv(upstreamRepoEnvKey, "Wei-Shaw/sub2api")

	github := &updateServiceGitHubClientStub{
		releases: map[string]*GitHubRelease{
			"custom/sub2api": {
				TagName: "v0.1.133",
				Name:    "custom-v0.1.133",
				HTMLURL: "https://github.com/custom/sub2api/releases/tag/v0.1.133",
			},
			"Wei-Shaw/sub2api": {
				TagName: "v0.1.134",
				Name:    "official-v0.1.134",
				HTMLURL: "https://github.com/Wei-Shaw/sub2api/releases/tag/v0.1.134",
			},
		},
	}
	svc := NewUpdateService(&updateServiceCacheStub{}, github, "0.1.132", "release")

	info, err := svc.CheckUpdate(context.Background(), true)

	require.NoError(t, err)
	require.Equal(t, []string{"custom/sub2api", "Wei-Shaw/sub2api"}, github.fetched)
	require.Equal(t, "custom/sub2api", info.UpdateRepo)
	require.Equal(t, "Wei-Shaw/sub2api", info.UpstreamRepo)
	require.Equal(t, "0.1.133", info.LatestVersion)
	require.True(t, info.HasUpdate)
	require.Equal(t, "0.1.134", info.UpstreamLatestVersion)
	require.True(t, info.HasUpstreamUpdate)
	require.Equal(t, "official-v0.1.134", info.UpstreamReleaseInfo.Name)
}
