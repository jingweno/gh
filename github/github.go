package github

import (
	"fmt"
	"github.com/octokit/go-octokit/octokit"
)

const (
	GitHubHost  string = "github.com"
	OAuthAppURL string = "http://owenou.com/gh"
)

type GitHub struct {
	Project *Project
	Config  *Config
}

func (gh *GitHub) PullRequest(id string) (pr *octokit.PullRequest, err error) {
	url, err := octokit.PullRequestsURL.Expand(octokit.M{"owner": gh.Project.Owner, "repo": gh.Project.Name, "number": id})
	if err != nil {
		return
	}

	client := gh.octokit()
	pr, result := client.PullRequests(url).One()
	if result.HasError() {
		err = result.Err
	}

	return
}

func (gh *GitHub) CreatePullRequest(base, head, title, body string) (pr *octokit.PullRequest, err error) {
	url, err := octokit.PullRequestsURL.Expand(octokit.M{"owner": gh.Project.Owner, "repo": gh.Project.Name})
	if err != nil {
		return
	}

	client := gh.octokit()
	params := octokit.PullRequestParams{Base: base, Head: head, Title: title, Body: body}
	pr, result := client.PullRequests(url).Create(params)
	if result.HasError() {
		err = result.Err
	}

	return
}

func (gh *GitHub) CreatePullRequestForIssue(base, head, issue string) (pr *octokit.PullRequest, err error) {
	url, err := octokit.PullRequestsURL.Expand(octokit.M{"owner": gh.Project.Owner, "repo": gh.Project.Name})
	if err != nil {
		return
	}

	client := gh.octokit()
	params := octokit.PullRequestForIssueParams{Base: base, Head: head, Issue: issue}
	pr, result := client.PullRequests(url).Create(params)
	if result.HasError() {
		err = result.Err
	}

	return
}

func (gh *GitHub) Repository(project Project) (repo *octokit.Repository, err error) {
	url, err := octokit.RepositoryURL.Expand(octokit.M{"owner": project.Owner, "repo": project.Name})
	if err != nil {
		return
	}

	client := gh.octokit()
	repo, result := client.Repositories(url).One()
	if result.HasError() {
		err = result.Err
	}

	return
}

// TODO: detach GitHub from Project
func (gh *GitHub) IsRepositoryExist(project Project) bool {
	repo, err := gh.Repository(project)

	return err == nil && repo != nil
}

func (gh *GitHub) CreateRepository(project Project, description, homepage string, isPrivate bool) (repo *octokit.Repository, err error) {
	var repoURL octokit.Hyperlink
	if project.Owner != gh.Config.FetchUser() {
		repoURL = octokit.OrgRepositoriesURL
	} else {
		repoURL = octokit.UserRepositoriesURL
	}

	url, err := repoURL.Expand(octokit.M{"org": project.Owner})
	if err != nil {
		return
	}

	client := gh.octokit()
	params := octokit.Repository{Name: project.Name, Description: description, Homepage: homepage, Private: isPrivate}
	repo, result := client.Repositories(url).Create(params)
	if result.HasError() {
		err = result.Err
	}

	return
}

func (gh *GitHub) Releases() (releases []octokit.Release, err error) {
	url, err := octokit.ReleasesURL.Expand(octokit.M{"owner": gh.Project.Owner, "repo": gh.Project.Name})
	if err != nil {
		return
	}

	client := gh.octokit()
	releases, result := client.Releases(url).All()
	if result.HasError() {
		err = result.Err
		return
	}

	return
}

func (gh *GitHub) CIStatus(sha string) (status *octokit.Status, err error) {
	url, err := octokit.StatusesURL.Expand(octokit.M{"owner": gh.Project.Owner, "repo": gh.Project.Name, "ref": sha})
	if err != nil {
		return
	}

	client := gh.octokit()
	statuses, result := client.Statuses(url).All()
	if result.HasError() {
		err = result.Err
		return
	}

	if len(statuses) > 0 {
		status = &statuses[0]
	}

	return
}

func (gh *GitHub) ForkRepository(name, owner string, noRemote bool) (repo *octokit.Repository, err error) {
	config := gh.Config
	project := Project{Name: name, Owner: config.User}
	r, err := gh.Repository(project)
	if err == nil && r != nil {
		err = fmt.Errorf("Error creating fork: %s exists on %s", r.FullName, GitHubHost)
		return
	}

	url, err := octokit.ForksURL.Expand(octokit.M{"owner": owner, "repo": name})
	if err != nil {
		return
	}

	client := gh.octokit()
	repo, result := client.Repositories(url).Create(nil)
	if result.HasError() {
		err = result.Err
	}

	return
}

func (gh *GitHub) Issues() (issues []octokit.Issue, err error) {
	url, err := octokit.RepoIssuesURL.Expand(octokit.M{"owner": gh.Project.Owner, "repo": gh.Project.Name})
	if err != nil {
		return
	}

	client := gh.octokit()
	issues, result := client.Issues(url).All()
	if result.HasError() {
		err = result.Err
		return
	}

	return
}

func (gh *GitHub) ExpandRemoteUrl(owner, name string, isSSH bool) (url string) {
	project := gh.Project
	if owner == "origin" {
		config := gh.Config
		owner = config.FetchUser()
	}

	return project.GitURL(name, owner, isSSH)
}

func findOrCreateToken(user, password, twoFactorCode string) (token string, err error) {
	url, err := octokit.AuthorizationsURL.Expand(nil)
	if err != nil {
		return
	}

	basicAuth := octokit.BasicAuth{Login: user, Password: password, OneTimePassword: twoFactorCode}
	client := octokit.NewClient(basicAuth)
	authsService := client.Authorizations(url)

	auths, result := authsService.All()
	if result.HasError() {
		err = result.Err
		return
	}

	for _, auth := range auths {
		if auth.NoteURL == OAuthAppURL {
			token = auth.Token
			break
		}
	}

	if token == "" {
		authParam := octokit.AuthorizationParams{}
		authParam.Scopes = append(authParam.Scopes, "repo")
		authParam.Note = "gh"
		authParam.NoteURL = OAuthAppURL

		auth, result := authsService.Create(authParam)
		if result.HasError() {
			err = result.Err
			return
		}

		token = auth.Token
	}

	return
}

func (gh *GitHub) octokit() *octokit.Client {
	config := gh.Config
	config.FetchCredentials()
	tokenAuth := octokit.TokenAuth{AccessToken: config.Token}

	return octokit.NewClient(tokenAuth)
}

func New() *GitHub {
	project := CurrentProject()
	c := CurrentConfig()
	c.FetchUser()

	return &GitHub{project, c}
}

// TODO: detach project from GitHub
func NewWithoutProject() *GitHub {
	c := CurrentConfig()
	c.FetchUser()

	return &GitHub{nil, c}
}
