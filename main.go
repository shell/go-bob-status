package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/go-redis/redis"
	"github.com/google/go-github/github"
	flags "github.com/jessevdk/go-flags"
	"golang.org/x/oauth2"
)

const (
	gitHubOwner = "revdotcom"
	gitHubRepo  = "revdotcom"
	noColor     = "\033[0m"
	colorGreen  = "\033[0;32m"
	colorRed    = "\033[0;31m"
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
	client := gojenkins.CreateJenkins(nil, "https://ci.rev.com", jenkinsUsername, jenkinsPassword)

	if _, err := client.Init(); err != nil {
		panic("Something went wrong")
	}

	return client
}

func getRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if _, err := client.Ping().Result(); err != nil {
		panic("can't connect to Redis")
	}
	return client
}

func getRedisKey(sha string) string {
	return "github:" + gitHubRepo + ":" + sha
}

func main() {
	args := opts{}
	parser := flags.NewParser(&args, flags.Default)

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	ghClient, ctx := getGitHubClient(args.GithubToken)
	jenkinsClient := getJenkinsClient(args.JenkinsUsername, args.JenkinsPassword)
	redisClient := getRedisClient()

	jobs, errj := jenkinsClient.GetAllJobs()

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

			statusColor := colorGreen
			if status != "success" {
				statusColor = colorRed
			}

			rStatus, rErr := redisClient.Get(getRedisKey(sha)).Result()
			if rErr != nil || rStatus != status {
				repoStatus := github.RepoStatus{
					State:     &status,
					TargetURL: &buildURL,
				}
				ghClient.Repositories.CreateStatus(*ctx, gitHubOwner, gitHubRepo, sha, &repoStatus)
				redisClient.Set(getRedisKey(sha), status, time.Hour*24*14)

				fmt.Printf("%vsha: %v with status: %v%v\n", noColor, sha, statusColor, status)
			} else {
				fmt.Printf("%vsha[Cache]: \033[1;30m%v %vwith status: %v%v\n", noColor, sha, noColor, statusColor, status)
			}
		}
	}
}
