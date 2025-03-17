package repo

import (
	"context"
	"reposync/internal/config"

	"testing"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func MockGitRemote(name string, urls []string) *git.Remote {
	return git.NewRemote(
		memory.NewStorage(),
		&gitconfig.RemoteConfig{
			Name: name,
			URLs: urls,
		})
}

type MockGitRepository struct {
	mock.Mock
}

func (m *MockGitRepository) Remote(name string) (*git.Remote, error) {
	args := m.Called(name)
	return args.Get(0).(*git.Remote), args.Error(1)
}

func (m *MockGitRepository) CreateRemote(cfg *gitconfig.RemoteConfig) (*git.Remote, error) {
	args := m.Called(cfg)
	return args.Get(0).(*git.Remote), args.Error(1)
}

func (m *MockGitRepository) DeleteRemote(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockGitRepository) FetchContext(ctx context.Context, opts *git.FetchOptions) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func TestCreateRemoteIfNeeded(t *testing.T) {
	ctx := context.Background()
	t.Run("creates remote if not found", func(t *testing.T) {
		mockRepo := new(MockGitRepository)
		sync := CreateMirrorSynchronizer(mockRepo)
		mockRepo.On("Remote", "origin").Return((*git.Remote)(nil), git.ErrRemoteNotFound)
		mockRepo.On("CreateRemote", mock.Anything).Return(&git.Remote{}, nil)

		err := sync.CreateRemoteIfNeeded(ctx, "https://example.com/repo.git", "origin")
		assert.NoError(t, err)
	})

	t.Run("does nothing if remote exists with same URL", func(t *testing.T) {
		mockRepo := new(MockGitRepository)
		sync := CreateMirrorSynchronizer(mockRepo)
		mockRemote := MockGitRemote("origin", []string{"https://example.com/repo.git"})
		mockRepo.On("Remote", "origin").Return(mockRemote, nil)

		err := sync.CreateRemoteIfNeeded(ctx, "https://example.com/repo.git", "origin")
		assert.NoError(t, err)
		mockRepo.AssertNotCalled(t, "CreateRemote", mock.Anything)
	})

	t.Run("recreates remote if URL differs", func(t *testing.T) {
		mockRepo := new(MockGitRepository)
		sync := CreateMirrorSynchronizer(mockRepo)
		mockRemote := MockGitRemote("origin", []string{"https://foo-bar"})
		mockRepo.On("Remote", "origin").Return(mockRemote, nil)
		mockRepo.On("DeleteRemote", "origin").Return(nil)
		mockRepo.On("CreateRemote", mock.Anything).Return(&git.Remote{}, nil)

		err := sync.CreateRemoteIfNeeded(ctx, "https://example.com/repo.git", "origin")
		assert.NoError(t, err)
		mockRepo.AssertCalled(t, "CreateRemote", mock.Anything)
	})
}

func TestSyncRepo(t *testing.T) {
	ctx := context.Background()
	t.Run("syncs repository successfully", func(t *testing.T) {
		mockRepo := new(MockGitRepository)
		sync := CreateMirrorSynchronizer(mockRepo)
		remoteConfig := config.RemoteConfig{
			URL:         "https://example.com",
			LocalPrefix: "foo",
			RefSpec:     []gitconfig.RefSpec{"+refs/heads/*:refs/heads/*", "+refs/tags/*:refs/tags/*"},
			Repos:       []string{"Alice", "Bob"},
		}

		mockRepo.On("Remote", "origin").Return((*git.Remote)(nil), git.ErrRemoteNotFound)
		mockRepo.On("CreateRemote", mock.Anything).Return(&git.Remote{}, nil)
		mockRepo.On("FetchContext", ctx, mock.Anything).Return(nil)
		sync.SyncRepo(ctx, remoteConfig, "repo-name")
		mockRepo.AssertCalled(t, "FetchContext", mock.Anything, mock.Anything)
	})

}
