package repo

import (
	"context"
	"path/filepath"
	"reposync/internal/config"
	"reposync/internal/logging"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
)

type Job struct {
	remote   config.RemoteConfig
	repoName string
	gitRoot  string
}

func worker(idx int, jobs <-chan Job, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		ctx := context.Background()
		ctx = context.WithValue(ctx, logging.ProjectNameKey, job.repoName)
		ctx = context.WithValue(ctx, logging.WorkerIdKey, idx)

		fullRepoPath := filepath.Join(job.gitRoot, job.remote.LocalPrefix, job.repoName+".git")
		repository, err := git.PlainOpen(fullRepoPath)
		if err != nil {
			logrus.WithContext(ctx).Errorf("cannot open repository: %s", fullRepoPath)
			continue
		}

		repoToSync := CreateMirrorSynchronizer(repository)
		repoToSync.SyncRepo(ctx, job.remote, job.repoName)
	}
}

func SyncReposInParallel(maxWorkers int, gitRoot string, remotes *map[string]config.RemoteConfig) {
	jobs := make(chan Job)
	var wg sync.WaitGroup

	for i := range maxWorkers {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	go func() {
		for _, remote := range *remotes {
			for _, repoName := range remote.Repos {
				jobs <- Job{remote: remote, repoName: repoName, gitRoot: gitRoot}
			}
		}
		close(jobs)
	}()

	wg.Wait()
}
