package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type IAMRoles struct {
	TaskExecutionRole    *iam.Role
	TaskRole             *iam.Role
	TaskExecutionRoleArn pulumi.StringOutput
	TaskRoleArn          pulumi.StringOutput
}

func createIAMRoles(ctx *pulumi.Context, projectName, environment string, secrets *SecretsResources, tags pulumi.StringMap) (*IAMRoles, error) {
	// Task Execution Role (for pulling images and reading secrets)
	taskExecutionRole, err := iam.NewRole(ctx, fmt.Sprintf("%s-%s-task-execution-role", projectName, environment), &iam.RoleArgs{
		Name: pulumi.String(fmt.Sprintf("%s-%s-task-execution-role", projectName, environment)),
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Action": "sts:AssumeRole",
				"Effect": "Allow",
				"Principal": {
					"Service": "ecs-tasks.amazonaws.com"
				}
			}]
		}`),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-task-execution-role", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Attach AWS managed policy for ECS task execution
	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-%s-task-execution-policy", projectName, environment), &iam.RolePolicyAttachmentArgs{
		Role:      taskExecutionRole.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
	})
	if err != nil {
		return nil, err
	}

	// Create custom policy for accessing Secrets Manager
	secretsPolicy := pulumi.All(secrets.DatabaseSecretArn, secrets.SuperTokensSecretArn).ApplyT(
		func(args []interface{}) (string, error) {
			dbSecretArn := args[0].(string)
			stSecretArn := args[1].(string)

			policyDoc := map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect": "Allow",
						"Action": []string{
							"secretsmanager:GetSecretValue",
						},
						"Resource": []string{
							dbSecretArn,
							stSecretArn,
						},
					},
					{
						"Effect": "Allow",
						"Action": []string{
							"kms:Decrypt",
						},
						"Resource": "*",
					},
				},
			}

			jsonData, err := json.Marshal(policyDoc)
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		},
	).(pulumi.StringOutput)

	_, err = iam.NewRolePolicy(ctx, fmt.Sprintf("%s-%s-secrets-policy", projectName, environment), &iam.RolePolicyArgs{
		Role:   taskExecutionRole.ID(),
		Policy: secretsPolicy,
	})
	if err != nil {
		return nil, err
	}

	// Task Role (for application runtime permissions)
	taskRole, err := iam.NewRole(ctx, fmt.Sprintf("%s-%s-task-role", projectName, environment), &iam.RoleArgs{
		Name: pulumi.String(fmt.Sprintf("%s-%s-task-role", projectName, environment)),
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Action": "sts:AssumeRole",
				"Effect": "Allow",
				"Principal": {
					"Service": "ecs-tasks.amazonaws.com"
				}
			}]
		}`),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-task-role", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Add policies for task role (application permissions)
	// Allow tasks to read secrets during runtime
	_, err = iam.NewRolePolicy(ctx, fmt.Sprintf("%s-%s-task-secrets-policy", projectName, environment), &iam.RolePolicyArgs{
		Role:   taskRole.ID(),
		Policy: secretsPolicy,
	})
	if err != nil {
		return nil, err
	}

	// Allow tasks to write logs
	logsPolicy := pulumi.String(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Action": [
				"logs:CreateLogStream",
				"logs:PutLogEvents"
			],
			"Resource": "*"
		}]
	}`)

	_, err = iam.NewRolePolicy(ctx, fmt.Sprintf("%s-%s-task-logs-policy", projectName, environment), &iam.RolePolicyArgs{
		Role:   taskRole.ID(),
		Policy: logsPolicy,
	})
	if err != nil {
		return nil, err
	}

	return &IAMRoles{
		TaskExecutionRole:    taskExecutionRole,
		TaskRole:             taskRole,
		TaskExecutionRoleArn: taskExecutionRole.Arn,
		TaskRoleArn:          taskRole.Arn,
	}, nil
}
