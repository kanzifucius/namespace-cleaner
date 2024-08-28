package github

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/owenrumney/go-github-pr-commenter/commenter"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"time"
)

type GitHub struct {
	owner string
	repo  string
	token string

	log *zap.Logger
}

func NewGitHub(owner string, repo string, token string, log *zap.Logger) *GitHub {
	return &GitHub{
		owner: owner,
		repo:  repo,
		token: token,
		log:   log,
	}

}

// get pull request details
func (gh *GitHub) GetPrDetails(prNumber int) (*github.PullRequest, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	pr, _, err := client.PullRequests.Get(ctx, gh.owner, gh.repo, prNumber)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// is pull request merged
func (gh *GitHub) IsPrMerged(prNumber int) (bool, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	pr, _, err := client.PullRequests.Get(ctx, gh.owner, gh.repo, prNumber)
	if err != nil {
		return false, err
	}

	if pr.Merged == nil {
		return false, nil
	}

	return *pr.Merged, nil
}

// is pull request closed
func (gh *GitHub) IsPrClosed(prNumber int) (bool, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	pr, _, err := client.PullRequests.Get(ctx, gh.owner, gh.repo, prNumber)
	if err != nil {
		return false, err
	}
	if pr.ClosedAt == nil {
		return false, nil
	}
	return *pr.ClosedAt != time.Time{}, nil
}

func (gh *GitHub) CommentOnPr(comment string, prNumber int) {

	c, err := commenter.NewCommenter(gh.token, gh.owner, gh.repo, prNumber)
	if err != nil {
		fmt.Println(err.Error())
	}

	// process whatever static analysis results you've gathered
	gh.log.Info(fmt.Sprintf("commenting on pr %d message %s", prNumber, comment))
	err = c.WriteGeneralComment(comment)
	if err != nil {
		gh.log.Error("failed to create comment")
		if errors.Is(err, commenter.CommentNotValidError{}) {
			gh.log.Error(fmt.Sprintf("result not relevant for commit. %s", err.Error()))
		} else {
			gh.log.Error(fmt.Sprintf("an error occurred writing the comment: %s", err.Error()))
		}
	}

}
