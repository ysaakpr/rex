package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type LogGroups struct {
	APILogGroup         *cloudwatch.LogGroup
	WorkerLogGroup      *cloudwatch.LogGroup
	SuperTokensLogGroup *cloudwatch.LogGroup
	MigrationLogGroup   *cloudwatch.LogGroup
}

func createLogGroups(ctx *pulumi.Context, projectName, environment string, tags pulumi.StringMap) (*LogGroups, error) {
	// API Log Group
	apiLogGroup, err := cloudwatch.NewLogGroup(ctx, fmt.Sprintf("%s-%s-api-logs", projectName, environment), &cloudwatch.LogGroupArgs{
		Name:            pulumi.String(fmt.Sprintf("/ecs/%s-%s-api", projectName, environment)),
		RetentionInDays: pulumi.Int(7), // 7 days for dev, increase for production
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-api-logs", projectName, environment)),
			"Service":     pulumi.String("api"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Worker Log Group
	workerLogGroup, err := cloudwatch.NewLogGroup(ctx, fmt.Sprintf("%s-%s-worker-logs", projectName, environment), &cloudwatch.LogGroupArgs{
		Name:            pulumi.String(fmt.Sprintf("/ecs/%s-%s-worker", projectName, environment)),
		RetentionInDays: pulumi.Int(7),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-worker-logs", projectName, environment)),
			"Service":     pulumi.String("worker"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Note: Frontend log group removed - frontend is now deployed via AWS Amplify
	// Amplify has its own logging in the Amplify console

	// SuperTokens Log Group
	supertokensLogGroup, err := cloudwatch.NewLogGroup(ctx, fmt.Sprintf("%s-%s-supertokens-logs", projectName, environment), &cloudwatch.LogGroupArgs{
		Name:            pulumi.String(fmt.Sprintf("/ecs/%s-%s-supertokens", projectName, environment)),
		RetentionInDays: pulumi.Int(7),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-supertokens-logs", projectName, environment)),
			"Service":     pulumi.String("supertokens"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Migration Log Group
	migrationLogGroup, err := cloudwatch.NewLogGroup(ctx, fmt.Sprintf("%s-%s-migration-logs", projectName, environment), &cloudwatch.LogGroupArgs{
		Name:            pulumi.String(fmt.Sprintf("/ecs/%s-%s-migration", projectName, environment)),
		RetentionInDays: pulumi.Int(30), // Keep migration logs longer
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-migration-logs", projectName, environment)),
			"Service":     pulumi.String("migration"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	return &LogGroups{
		APILogGroup:         apiLogGroup,
		WorkerLogGroup:      workerLogGroup,
		SuperTokensLogGroup: supertokensLogGroup,
		MigrationLogGroup:   migrationLogGroup,
	}, nil
}
