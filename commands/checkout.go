package commands

import (
	"fmt"
	"github.com/jingweno/gh/github"
	"github.com/jingweno/gh/utils"
	"regexp"
)

var cmdCheckout = &Command{
	Run:          checkout,
	GitExtension: true,
	Usage:        "checkout PULLREQ-URL [BRANCH]",
	Short:        "Switch the active branch to another branch",
	Long: `Checks out the head of the pull request as a local branch, to allow for
reviewing, rebasing and otherwise cleaning up the commits in the pull
request before merging. The name of the local branch can explicitly be
set with BRANCH.
`,
}

/**
  $ gh checkout https://github.com/jingweno/gh/pull/73
  > git remote add -f -t feature git://github:com/foo/gh.git
  > git checkout --track -B foo-feature foo/feature

  $ gh checkout https://github.com/jingweno/gh/pull/73 custom-branch-name
**/
func checkout(command *Command, args *Args) {
	if !args.IsParamsEmpty() {
		err := transformCheckoutArgs(args)
		utils.Check(err)
	}
}

func transformCheckoutArgs(args *Args) error {
	words := args.Words()
	if len(words) == 0 {
		return nil
	}

	checkoutURL := words[0]
	url, err := github.ParseURL(checkoutURL)
	if err != nil {
		return nil
	}
	var newBranchName string
	if len(words) > 1 {
		newBranchName = words[1]
	}

	pullURLRegex := regexp.MustCompile("^pull/(\\d+)")
	projectPath := url.ProjectPath()
	if !pullURLRegex.MatchString(projectPath) {
		return nil
	}

	id := pullURLRegex.FindStringSubmatch(projectPath)[1]
	gh := github.NewClient(url.Project.Host)
	pullRequest, err := gh.PullRequest(url.Project, id)
	if err != nil {
		return err
	}

	if idx := args.IndexOfParam(newBranchName); idx >= 0 {
		args.RemoveParam(idx)
	}

	user, branch := parseUserBranchFromPR(pullRequest)
	if pullRequest.Head.Repo.ID == 0 {
		return fmt.Errorf("Error: %s's fork is not available anymore", user)
	}

	if newBranchName == "" {
		newBranchName = fmt.Sprintf("%s-%s", user, branch)
	}

	repo := github.LocalRepo()
	_, err = repo.RemoteByName(user)
	if err == nil {
		args.Before("git", "remote", "set-branches", "--add", user, branch)
		remoteURL := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", branch, user, branch)
		args.Before("git", "fetch", user, remoteURL)
	} else {
		u := url.Project.GitURL("", user, pullRequest.Head.Repo.Private)
		args.Before("git", "remote", "add", "-f", "-t", branch, user, u)
	}

	idx := args.IndexOfParam(checkoutURL)
	args.RemoveParam(idx)
	args.InsertParam(idx, "--track", "-B", newBranchName, fmt.Sprintf("%s/%s", user, branch))

	return nil
}
