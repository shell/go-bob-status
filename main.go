package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bndr/gojenkins"
	"github.com/google/go-github/github"
	flags "github.com/jessevdk/go-flags"
	"golang.org/x/oauth2"
)

const (
	gitHubOwner = "revdotcom"
	gitHubRepo  = "revdotcom"
)

type opts struct {
	GithubToken     string `short:"t" long:"token" description:"Github Api Token" required:"true"`
	JenkinsUsername string `short:"u" long:"user" description:"Jenkins username" required:"true"`
	JenkinsPassword string `short:"p" long:"password" description:"Jenkins password" required:"true"`
}

func getGitHubClient(gitHubToken string) (*github.Client, *context.Context) {
	ctx := context.Background()
	tokenService := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitHubToken},
	)
	tokenClient := oauth2.NewClient(ctx, tokenService)
	ghClient := github.NewClient(tokenClient)
	return ghClient, &ctx
}

func getJenkinsClient(jenkinsUsername, jenkinsPassword string) *gojenkins.Jenkins {
	client := gojenkins.CreateJenkins(nil, "http://ci.rev.com", jenkinsUsername, jenkinsPassword)

	_, err := client.Init()
	if err != nil {
		panic("Something went wrong")
	}

	return client
}

func main() {
	args := opts{}
	parser := flags.NewParser(&args, flags.Default)

	_, erra := parser.Parse()
	if erra != nil {
		os.Exit(1)
	}

	ghClient, ctx := getGitHubClient(args.GithubToken)
	client := getJenkinsClient(args.JenkinsUsername, args.JenkinsPassword)

	jobs, errj := client.GetAllJobs()

	if errj != nil {
		panic("can't fetch jobs")
	}

	for _, job := range jobs {
		jobName := job.GetName()

		if strings.HasPrefix(jobName, "Rev.com-build-feature_") {
			build, _ := job.GetLastBuild()
			sha := build.GetRevision()

			var status string
			if build.IsRunning() {
				status = "pending"
			} else {
				if build.IsGood() {
					status = "success"
				} else {
					status = "failure"
				}
			}
			buildURL := build.GetUrl() + "console" // Point to the console log directly

			repoStatus := github.RepoStatus{
				State:     &status,
				TargetURL: &buildURL,
			}
			ghClient.Repositories.CreateStatus(*ctx, gitHubOwner, gitHubRepo, sha, &repoStatus)
			fmt.Printf("sha: %v with status: %v\n", sha, status)
		}
	}
}

