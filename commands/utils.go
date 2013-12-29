package commands

import (
	"github.com/jingweno/gh/github"
	"github.com/jingweno/gh/utils"
	"github.com/jingweno/go-octokit/octokit"
	"os"
	"path/filepath"
	"strings"
)

func isDir(file string) bool {
	f, err := os.Open(file)
	if err != nil {
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false
	}

	return fi.IsDir()
}

func parseUserBranchFromPR(pullRequest *octokit.PullRequest) (user string, branch string) {
	userBranch := strings.SplitN(pullRequest.Head.Label, ":", 2)
	user = userBranch[0]
	if len(userBranch) > 1 {
		branch = userBranch[1]
	} else {
		branch = pullRequest.Head.Ref
	}

	return
}

func hasGitRemote(name string) bool {
	remotes, err := github.Remotes()
	utils.Check(err)
	for _, remote := range remotes {
		if remote.Name == name {
			return true
		}
	}

	return false
}

func isEmptyDir(path string) bool {
	fullPath := filepath.Join(path, "*")
	match, _ := filepath.Glob(fullPath)
	return match == nil
}
