package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/amplify"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AmplifyResources struct {
	App           *amplify.App
	Branch        *amplify.Branch
	AppID         pulumi.StringOutput
	AppARN        pulumi.StringOutput
	DefaultDomain pulumi.StringOutput
	BranchURL     pulumi.StringOutput
}

func createAmplifyApp(ctx *pulumi.Context, projectName, environment string, alb *ALBResources,
	githubRepo, githubBranch, githubToken string, tags pulumi.StringMap) (*AmplifyResources, error) {

	// Build specification for Vite React app
	buildSpec := `version: 1
frontend:
  phases:
    preBuild:
      commands:
        - cd frontend
        - npm ci
    build:
      commands:
        - npm run build
  artifacts:
    baseDirectory: frontend/dist
    files:
      - '**/*'
  cache:
    paths:
      - frontend/node_modules/**/*`

	// Create Amplify App arguments
	amplifyArgs := &amplify.AppArgs{
		Name:       pulumi.String(fmt.Sprintf("%s-%s-frontend", projectName, environment)),
		Repository: pulumi.String(githubRepo),
		Platform:   pulumi.String("WEB"),
		BuildSpec:  pulumi.String(buildSpec),

		// Environment variables for the frontend build
		EnvironmentVariables: pulumi.StringMap{
			"VITE_API_URL":  alb.DNSName.ApplyT(func(dns string) string { return fmt.Sprintf("http://%s/api", dns) }).(pulumi.StringOutput),
			"VITE_AUTH_URL": alb.DNSName.ApplyT(func(dns string) string { return fmt.Sprintf("http://%s/auth", dns) }).(pulumi.StringOutput),
			"NODE_ENV":      pulumi.String("production"),
		},

		// Custom rules for SPA routing
		CustomRules: amplify.AppCustomRuleArray{
			// Redirect all requests to index.html for client-side routing
			&amplify.AppCustomRuleArgs{
				Source: pulumi.String("</^[^.]+$|\\.(?!(css|gif|ico|jpg|js|png|txt|svg|woff|woff2|ttf|map|json)$)([^.]+$)/>"),
				Target: pulumi.String("/index.html"),
				Status: pulumi.String("200"),
			},
		},

		// Enable auto branch creation for feature branches (optional)
		EnableBranchAutoBuild:    pulumi.Bool(true),
		EnableBranchAutoDeletion: pulumi.Bool(true),

		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-frontend", projectName, environment)),
			"Service":     pulumi.String("frontend"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	}

	// Only set AccessToken if provided (required for private repos, optional for public)
	if githubToken != "" {
		amplifyArgs.AccessToken = pulumi.String(githubToken)
	}

	// Create Amplify App
	amplifyApp, err := amplify.NewApp(ctx, fmt.Sprintf("%s-%s-frontend", projectName, environment), amplifyArgs)
	if err != nil {
		return nil, err
	}

	// Create branch (e.g., main)
	branch, err := amplify.NewBranch(ctx, fmt.Sprintf("%s-%s-frontend-branch", projectName, environment), &amplify.BranchArgs{
		AppId:      amplifyApp.ID(),
		BranchName: pulumi.String(githubBranch),

		// Enable auto build on push
		EnableAutoBuild: pulumi.Bool(true),

		// Production branch settings
		Stage: pulumi.String("PRODUCTION"),

		// Framework detection
		Framework: pulumi.String("React"),

		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-frontend-branch", projectName, environment)),
			"Branch":      pulumi.String(githubBranch),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Generate branch URL
	branchURL := pulumi.All(branch.BranchName, amplifyApp.DefaultDomain).ApplyT(
		func(args []interface{}) string {
			branchName := args[0].(string)
			defaultDomain := args[1].(string)
			return fmt.Sprintf("https://%s.%s", branchName, defaultDomain)
		},
	).(pulumi.StringOutput)

	return &AmplifyResources{
		App:           amplifyApp,
		Branch:        branch,
		AppID:         amplifyApp.ID().ToStringOutput(),
		AppARN:        amplifyApp.Arn,
		DefaultDomain: amplifyApp.DefaultDomain,
		BranchURL:     branchURL,
	}, nil
}
