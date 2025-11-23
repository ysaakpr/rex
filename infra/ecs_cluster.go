package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ECSClusterResources struct {
	Cluster     *ecs.Cluster
	ClusterName pulumi.StringOutput
	ClusterARN  pulumi.StringOutput
}

func createECSCluster(ctx *pulumi.Context, projectName, environment string, tags pulumi.StringMap) (*ECSClusterResources, error) {
	// Create ECS Cluster
	cluster, err := ecs.NewCluster(ctx, fmt.Sprintf("%s-%s-cluster", projectName, environment), &ecs.ClusterArgs{
		Name: pulumi.String(fmt.Sprintf("%s-%s-cluster", projectName, environment)),
		Settings: ecs.ClusterSettingArray{
			&ecs.ClusterSettingArgs{
				Name:  pulumi.String("containerInsights"),
				Value: pulumi.String("enabled"),
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-cluster", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	return &ECSClusterResources{
		Cluster:     cluster,
		ClusterName: cluster.Name,
		ClusterARN:  cluster.Arn,
	}, nil
}
