package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type MigrationTaskResources struct {
	TaskDefinition    *ecs.TaskDefinition
	TaskDefinitionARN pulumi.StringOutput
}

type MigrationInput struct {
	DatabaseEndpoint pulumi.StringOutput
	MasterUsername   string
	MainDBName       string
	MasterPassword   pulumi.StringOutput
}

func createMigrationTask(ctx *pulumi.Context, projectName, environment string, network *NetworkingResources,
	roles *IAMRoles, logs *LogGroups, repositories *ECRRepositories, secrets *SecretsResources,
	database *DatabaseResources, tags pulumi.StringMap) (*MigrationTaskResources, error) {

	containerDef := pulumi.All(repositories.APIRepoURL, database.ClusterEndpoint,
		secrets.DatabaseSecretArn, logs.MigrationLogGroup.Name).ApplyT(
		func(args []interface{}) (string, error) {
			imageURL := args[0].(string)
			dbEndpoint := args[1].(string)
			secretArn := args[2].(string)
			logGroup := args[3].(string)

			containers := []map[string]interface{}{
				{
					"name":  "migration",
					"image": fmt.Sprintf("%s:latest", imageURL),
					"command": []string{
						"/app/migrate",
						"up",
					},
					"environment": []map[string]interface{}{
						{"name": "APP_ENV", "value": environment},
						{"name": "DB_HOST", "value": dbEndpoint},
						{"name": "DB_PORT", "value": "5432"},
						{"name": "DB_USER", "value": database.MasterUsername},
						{"name": "DB_NAME", "value": database.MainDBName},
						{"name": "DB_SSL_MODE", "value": "require"},
					},
					"secrets": []map[string]interface{}{
						{
							"name":      "DB_PASSWORD",
							"valueFrom": fmt.Sprintf("%s:password::", secretArn),
						},
					},
					"logConfiguration": map[string]interface{}{
						"logDriver": "awslogs",
						"options": map[string]interface{}{
							"awslogs-group":         logGroup,
							"awslogs-region":        "ap-south-1",
							"awslogs-stream-prefix": "migration",
						},
					},
				},
			}

			jsonData, err := json.Marshal(containers)
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		},
	).(pulumi.StringOutput)

	taskDefinition, err := ecs.NewTaskDefinition(ctx, fmt.Sprintf("%s-%s-migration-task", projectName, environment), &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(fmt.Sprintf("%s-%s-migration", projectName, environment)),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		Cpu:                     pulumi.String("256"),
		Memory:                  pulumi.String("512"),
		ExecutionRoleArn:        roles.TaskExecutionRoleArn,
		TaskRoleArn:             roles.TaskRoleArn,
		ContainerDefinitions:    containerDef,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-migration-task", projectName, environment)),
			"Service":     pulumi.String("migration"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	return &MigrationTaskResources{
		TaskDefinition:    taskDefinition,
		TaskDefinitionARN: taskDefinition.Arn,
	}, nil
}

func createMigrationTaskFromInput(ctx *pulumi.Context, projectName, environment string, network *NetworkingResources,
	roles *IAMRoles, logs *LogGroups, repositories *ECRRepositories, secrets *SecretsResources,
	input *MigrationInput, tags pulumi.StringMap) (*MigrationTaskResources, error) {

	containerDef := pulumi.All(repositories.APIRepoURL, input.DatabaseEndpoint,
		secrets.DatabaseSecretArn, logs.MigrationLogGroup.Name).ApplyT(
		func(args []interface{}) (string, error) {
			imageURL := args[0].(string)
			dbEndpoint := args[1].(string)
			secretArn := args[2].(string)
			logGroup := args[3].(string)

			containers := []map[string]interface{}{
				{
					"name":  "migration",
					"image": fmt.Sprintf("%s:latest", imageURL),
					"command": []string{
						"/app/migrate",
						"up",
					},
					"environment": []map[string]interface{}{
						{"name": "APP_ENV", "value": environment},
						{"name": "DB_HOST", "value": dbEndpoint},
						{"name": "DB_PORT", "value": "5432"},
						{"name": "DB_USER", "value": input.MasterUsername},
						{"name": "DB_NAME", "value": input.MainDBName},
						{"name": "DB_SSL_MODE", "value": "require"},
					},
					"secrets": []map[string]interface{}{
						{
							"name":      "DB_PASSWORD",
							"valueFrom": fmt.Sprintf("%s:password::", secretArn),
						},
					},
					"logConfiguration": map[string]interface{}{
						"logDriver": "awslogs",
						"options": map[string]interface{}{
							"awslogs-group":         logGroup,
							"awslogs-region":        "ap-south-1",
							"awslogs-stream-prefix": "migration",
						},
					},
				},
			}

			jsonData, err := json.Marshal(containers)
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		},
	).(pulumi.StringOutput)

	taskDefinition, err := ecs.NewTaskDefinition(ctx, fmt.Sprintf("%s-%s-migration-task", projectName, environment), &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(fmt.Sprintf("%s-%s-migration", projectName, environment)),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		Cpu:                     pulumi.String("256"),
		Memory:                  pulumi.String("512"),
		ExecutionRoleArn:        roles.TaskExecutionRoleArn,
		TaskRoleArn:             roles.TaskRoleArn,
		ContainerDefinitions:    containerDef,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-migration-task", projectName, environment)),
			"Service":     pulumi.String("migration"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	return &MigrationTaskResources{
		TaskDefinition:    taskDefinition,
		TaskDefinitionARN: taskDefinition.Arn,
	}, nil
}
