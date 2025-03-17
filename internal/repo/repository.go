package repo

import (
	"context"
	"errors"
	"fmt"
	"reposync/internal/config"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/sirupsen/logrus"
)

type GitRepository interface {
	Remote(name string) (*git.Remote, error)
	CreateRemote(cfg *gitconfig.RemoteConfig) (*git.Remote, error)
	DeleteRemote(name string) error
	FetchContext(ctx context.Context, opts *git.FetchOptions) error
}
type MirrorSynchronizer struct {
	repo GitRepository
}

func CreateMirrorSynchronizer(repo GitRepository) *MirrorSynchronizer {
	return &MirrorSynchronizer{
		repo: repo,
	}
}

func (s *MirrorSynchronizer) CreateRemoteIfNeeded(ctx context.Context, url string, name string) error {
	remote, err := s.repo.Remote(name)

	if err != nil {
		if errors.Is(err, git.ErrRemoteNotFound) {
			logrus.WithContext(ctx).Info("remote not found, proceeding with creation")
		} else {
			logrus.WithContext(ctx).Errorf("failed to get remote: %v", err)
			return err
		}
	}

	if remote != nil {
		if url == remote.Config().URLs[0] {
			return nil
		}

		logrus.WithContext(ctx).Infof("remote %s exists but with a different URL (%s). Recreating it.", name, url)
		_ = s.deleteRemote(name)
	}

	logrus.WithContext(ctx).Infof("remote %s not found. Creating new remote", name)
	_, err = s.repo.CreateRemote(&gitconfig.RemoteConfig{
		Name: name,
		URLs: []string{url},
		Fetch: []gitconfig.RefSpec{
			"+refs/heads/*:refs/heads/*",
			"+refs/tags/*:refs/tags/*",
		},
		Mirror: true,
	})
	if err != nil {
		return err
	}

	logrus.WithContext(ctx).Infof("remote %s created successfully", name)
	return nil
}

func (s *MirrorSynchronizer) SyncRepo(ctx context.Context, remote config.RemoteConfig, repoName string) {
	projectRemoteURL := remote.URL + "/" + repoName
	if err := s.CreateRemoteIfNeeded(ctx, projectRemoteURL, "origin"); err != nil {
		logrus.WithContext(ctx).Errorf("error creating remote: %v", err)
		return
	}

	logrus.WithContext(ctx).Infof("fetching from remote: %s", projectRemoteURL)
	err := s.repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   remote.RefSpec,
		Progress:   logrus.StandardLogger().WithContext(ctx).Writer(),
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		logrus.WithContext(ctx).Error(err)
	}
	if err != nil && errors.Is(err, git.NoErrAlreadyUpToDate) {
		logrus.WithContext(ctx).Info(err)
	}

	logrus.WithContext(ctx).Info("repository successfully synced")
}

func (s *MirrorSynchronizer) deleteRemote(name string) error {
	err := s.repo.DeleteRemote(name)
	if err != nil && err != git.ErrRemoteNotFound {
		return fmt.Errorf("failed to delete remote %v: %v", name, err)
	}
	return nil
}
