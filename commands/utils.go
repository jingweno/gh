package commands

import (
	"fmt"
	"github.com/jingweno/gh/github"
	"github.com/jingweno/octokat"
	"os"
	"regexp"
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

func parsePullRequestId(url string) string {
	pullURLRegex := regexp.MustCompile("https://github\\.com/.+/.+/pull/(\\d+)")
	if pullURLRegex.MatchString(url) {
		return pullURLRegex.FindStringSubmatch(url)[1]
	}

	return ""
}

func fetchPullRequest(id string) (*octokat.PullRequest, error) {
	gh := github.New()
	pullRequest, err := gh.PullRequest(id)
	if err != nil {
		return nil, err
	}

	if pullRequest.Head.Repo.ID == 0 {
		user := pullRequest.User.Login
		return nil, fmt.Errorf("%s's fork is not available anymore", user)
	}

	return pullRequest, nil
}

func convertToGitURL(pullRequest *octokat.PullRequest) (string, error) {
	pullRequestURL := pullRequest.HTMLURL
	user := pullRequest.User.Login
	isSSH := pullRequest.Head.Repo.Private

	project, err := github.ParseProjectFromURL(pullRequestURL)
	if err != nil {
		return "", err
	}

	return project.GitURL("", user, isSSH), nil
}
