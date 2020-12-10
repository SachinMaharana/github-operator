package git

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v28/github"
	githubv1 "github.com/sachinmaharana/github-operator/api/v1"
	"golang.org/x/oauth2"
)

// Client interface
type Client interface {
	GetRepo(context.Context, string, string) (*github.Repository, *github.Response, error)
	CreateRepo(context.Context, string, *githubv1.Repo) error
}

type client struct {
	ts oauth2.TokenSource
	tc *http.Client
	c  *github.Client
}

// New is ...
func New(ctx context.Context, token string) (cl Client, err error) {
	cli := &client{}
	cli.ts = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	cli.tc = oauth2.NewClient(ctx, cli.ts)

	cli.c = github.NewClient(cli.tc)
	return cli, nil
}

func (c *client) GetRepo(ctx context.Context, org, name string) (repo *github.Repository, resp *github.Response, err error) {
	repo, resp, err = c.c.Repositories.Get(ctx, org, name)
	if err != nil {
		return repo, resp, err
	}

	return repo, resp, nil
}

func (c *client) CreateRepo(ctx context.Context, org string, repo *githubv1.Repo) error {
	authUser, _, err := c.c.Users.Get(ctx, "")
	if err != nil {
		log.Fatal(err)
	}

	if authUser.GetLogin() == org {
		// username is equal to target organization
		org = "" // pass an empty string, only for repository creation
	}

	r := newRepository(repo)
	_, _, err = c.c.Repositories.Create(ctx, org, r)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func newRepository(repo *githubv1.Repo) *github.Repository {
	return &github.Repository{
		Name:        &repo.Name,
		Description: &repo.Spec.Description,
		Homepage:    &repo.Spec.Homepage,
		Private:     &repo.Spec.Options.Private,
		HasIssues:   &repo.Spec.Options.Issues,
		HasProjects: &repo.Spec.Options.Projects,
		IsTemplate:  &repo.Spec.Options.Template,
	}
}
