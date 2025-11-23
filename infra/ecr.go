package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ECRRepositories struct {
	APIRepo       *ecr.Repository
	WorkerRepo    *ecr.Repository
	APIRepoURL    pulumi.StringOutput
	WorkerRepoURL pulumi.StringOutput
}

func createECRRepositories(ctx *pulumi.Context, projectName, environment string, tags pulumi.StringMap) (*ECRRepositories, error) {
	// Lifecycle policy to keep only the last 10 images
	lifecyclePolicy := map[string]interface{}{
		"rules": []map[string]interface{}{
			{
				"rulePriority": 1,
				"description":  "Keep last 10 images",
				"selection": map[string]interface{}{
					"tagStatus":   "any",
					"countType":   "imageCountMoreThan",
					"countNumber": 10,
				},
				"action": map[string]interface{}{
					"type": "expire",
				},
			},
		},
	}
	lifecyclePolicyJSON, _ := json.Marshal(lifecyclePolicy)

	// API Repository
	apiRepo, err := ecr.NewRepository(ctx, fmt.Sprintf("%s-%s-api", projectName, environment), &ecr.RepositoryArgs{
		Name:               pulumi.String(fmt.Sprintf("%s-%s-api", projectName, environment)),
		ImageTagMutability: pulumi.String("MUTABLE"),
		ImageScanningConfiguration: &ecr.RepositoryImageScanningConfigurationArgs{
			ScanOnPush: pulumi.Bool(true),
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-api", projectName, environment)),
			"Service":     pulumi.String("api"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = ecr.NewLifecyclePolicy(ctx, fmt.Sprintf("%s-%s-api-lifecycle", projectName, environment), &ecr.LifecyclePolicyArgs{
		Repository: apiRepo.Name,
		Policy:     pulumi.String(string(lifecyclePolicyJSON)),
	})
	if err != nil {
		return nil, err
	}

	// Worker Repository
	workerRepo, err := ecr.NewRepository(ctx, fmt.Sprintf("%s-%s-worker", projectName, environment), &ecr.RepositoryArgs{
		Name:               pulumi.String(fmt.Sprintf("%s-%s-worker", projectName, environment)),
		ImageTagMutability: pulumi.String("MUTABLE"),
		ImageScanningConfiguration: &ecr.RepositoryImageScanningConfigurationArgs{
			ScanOnPush: pulumi.Bool(true),
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-worker", projectName, environment)),
			"Service":     pulumi.String("worker"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = ecr.NewLifecyclePolicy(ctx, fmt.Sprintf("%s-%s-worker-lifecycle", projectName, environment), &ecr.LifecyclePolicyArgs{
		Repository: workerRepo.Name,
		Policy:     pulumi.String(string(lifecyclePolicyJSON)),
	})
	if err != nil {
		return nil, err
	}

	// Note: Frontend repository removed - frontend is now deployed via AWS Amplify

	return &ECRRepositories{
		APIRepo:       apiRepo,
		WorkerRepo:    workerRepo,
		APIRepoURL:    apiRepo.RepositoryUrl,
		WorkerRepoURL: workerRepo.RepositoryUrl,
	}, nil
}
