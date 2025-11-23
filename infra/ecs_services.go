package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/servicediscovery"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ECSServicesResources struct {
	APIService             *ecs.Service
	WorkerService          *ecs.Service
	SuperTokensService     *ecs.Service
	APIServiceName         pulumi.StringOutput
	WorkerServiceName      pulumi.StringOutput
	SuperTokensServiceName pulumi.StringOutput
}

func createECSServices(ctx *pulumi.Context, projectName, environment string, cluster *ECSClusterResources,
	network *NetworkingResources, securityGroups *SecurityGroups, alb *ALBResources, roles *IAMRoles,
	logs *LogGroups, repositories *ECRRepositories, secrets *SecretsResources, database *DatabaseResources,
	redis *RedisResources, supertokensApiKey pulumi.StringOutput, tags pulumi.StringMap) (*ECSServicesResources, error) {

	// Create private namespace for service discovery
	namespace, err := servicediscovery.NewPrivateDnsNamespace(ctx, fmt.Sprintf("%s-%s-namespace", projectName, environment), &servicediscovery.PrivateDnsNamespaceArgs{
		Name: pulumi.String(fmt.Sprintf("%s-%s.local", projectName, environment)),
		Vpc:  network.VpcID.ToStringOutput(),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-namespace", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create service discovery service for SuperTokens
	supertokensDiscovery, err := servicediscovery.NewService(ctx, fmt.Sprintf("%s-%s-supertokens-discovery", projectName, environment), &servicediscovery.ServiceArgs{
		Name: pulumi.String("supertokens"),
		DnsConfig: &servicediscovery.ServiceDnsConfigArgs{
			NamespaceId: namespace.ID(),
			DnsRecords: servicediscovery.ServiceDnsConfigDnsRecordArray{
				&servicediscovery.ServiceDnsConfigDnsRecordArgs{
					Ttl:  pulumi.Int(10),
					Type: pulumi.String("A"),
				},
			},
		},
		HealthCheckCustomConfig: &servicediscovery.ServiceHealthCheckCustomConfigArgs{
			FailureThreshold: pulumi.Int(1),
		},
	})
	if err != nil {
		return nil, err
	}

	// SuperTokens Task Definition
	supertokensTaskDef, err := createSuperTokensTaskDefinition(ctx, projectName, environment, roles, logs, database, secrets, supertokensApiKey, tags)
	if err != nil {
		return nil, err
	}

	// SuperTokens Service
	supertokensService, err := ecs.NewService(ctx, fmt.Sprintf("%s-%s-supertokens-service", projectName, environment), &ecs.ServiceArgs{
		Name:           pulumi.String(fmt.Sprintf("%s-%s-supertokens", projectName, environment)),
		Cluster:        cluster.ClusterARN,
		TaskDefinition: supertokensTaskDef.Arn,
		DesiredCount:   pulumi.Int(1),
		LaunchType:     pulumi.String("FARGATE"),
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			Subnets:        network.PrivateSubnetIDs,
			SecurityGroups: pulumi.StringArray{securityGroups.SuperTokensSG.ID().ToStringOutput()},
		},
		LoadBalancers: ecs.ServiceLoadBalancerArray{
			&ecs.ServiceLoadBalancerArgs{
				TargetGroupArn: alb.SuperTokensTargetGroup.Arn,
				ContainerName:  pulumi.String("supertokens"),
				ContainerPort:  pulumi.Int(3567),
			},
		},
		ServiceRegistries: &ecs.ServiceServiceRegistriesArgs{
			RegistryArn: supertokensDiscovery.Arn,
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-supertokens-service", projectName, environment)),
			"Service":     pulumi.String("supertokens"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	}, pulumi.DependsOn([]pulumi.Resource{alb.HTTPListener}))
	if err != nil {
		return nil, err
	}

	// API Task Definition
	apiTaskDef, err := createAPITaskDefinition(ctx, projectName, environment, roles, logs, repositories, database, redis, secrets, namespace, tags)
	if err != nil {
		return nil, err
	}

	// API Service
	apiService, err := ecs.NewService(ctx, fmt.Sprintf("%s-%s-api-service", projectName, environment), &ecs.ServiceArgs{
		Name:           pulumi.String(fmt.Sprintf("%s-%s-api", projectName, environment)),
		Cluster:        cluster.ClusterARN,
		TaskDefinition: apiTaskDef.Arn,
		DesiredCount:   pulumi.Int(2), // 2 instances for HA
		LaunchType:     pulumi.String("FARGATE"),
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			Subnets:        network.PrivateSubnetIDs,
			SecurityGroups: pulumi.StringArray{securityGroups.ECSSG.ID().ToStringOutput()},
		},
		LoadBalancers: ecs.ServiceLoadBalancerArray{
			&ecs.ServiceLoadBalancerArgs{
				TargetGroupArn: alb.APITargetGroup.Arn,
				ContainerName:  pulumi.String("api"),
				ContainerPort:  pulumi.Int(8080),
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-api-service", projectName, environment)),
			"Service":     pulumi.String("api"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	}, pulumi.DependsOn([]pulumi.Resource{supertokensService, alb.HTTPListener}))
	if err != nil {
		return nil, err
	}

	// Worker Task Definition
	workerTaskDef, err := createWorkerTaskDefinition(ctx, projectName, environment, roles, logs, repositories, database, redis, secrets, tags)
	if err != nil {
		return nil, err
	}

	// Worker Service
	workerService, err := ecs.NewService(ctx, fmt.Sprintf("%s-%s-worker-service", projectName, environment), &ecs.ServiceArgs{
		Name:           pulumi.String(fmt.Sprintf("%s-%s-worker", projectName, environment)),
		Cluster:        cluster.ClusterARN,
		TaskDefinition: workerTaskDef.Arn,
		DesiredCount:   pulumi.Int(1), // Single worker for background jobs
		LaunchType:     pulumi.String("FARGATE"),
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			Subnets:        network.PrivateSubnetIDs,
			SecurityGroups: pulumi.StringArray{securityGroups.ECSSG.ID().ToStringOutput()},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-worker-service", projectName, environment)),
			"Service":     pulumi.String("worker"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Note: Frontend is now deployed via AWS Amplify, not ECS
	// See amplify.go for frontend deployment configuration

	return &ECSServicesResources{
		APIService:             apiService,
		WorkerService:          workerService,
		SuperTokensService:     supertokensService,
		APIServiceName:         apiService.Name,
		WorkerServiceName:      workerService.Name,
		SuperTokensServiceName: supertokensService.Name,
	}, nil
}

func createSuperTokensTaskDefinition(ctx *pulumi.Context, projectName, environment string, roles *IAMRoles,
	logs *LogGroups, database *DatabaseResources, secrets *SecretsResources, supertokensApiKey pulumi.StringOutput, tags pulumi.StringMap) (*ecs.TaskDefinition, error) {

	containerDef := pulumi.All(database.ClusterEndpoint, secrets.DatabaseSecretArn, logs.SuperTokensLogGroup.Name).ApplyT(
		func(args []interface{}) (string, error) {
			endpoint := args[0].(string)
			secretArn := args[1].(string)
			logGroup := args[2].(string)

			containers := []map[string]interface{}{
				{
					"name":  "supertokens",
					"image": "registry.supertokens.io/supertokens/supertokens-postgresql:7.0",
					"portMappings": []map[string]interface{}{
						{
							"containerPort": 3567,
							"protocol":      "tcp",
						},
					},
					"environment": []map[string]interface{}{
						{
							"name":  "POSTGRESQL_CONNECTION_URI",
							"value": fmt.Sprintf("postgresql://%s@%s:5432/%s", database.MasterUsername, endpoint, database.SuperTokensDBName),
						},
					},
					"secrets": []map[string]interface{}{
						{
							"name":      "POSTGRESQL_PASSWORD",
							"valueFrom": fmt.Sprintf("%s:password::", secretArn),
						},
						{
							"name":      "API_KEYS",
							"valueFrom": fmt.Sprintf("%s:api_key::", secretArn),
						},
					},
					"logConfiguration": map[string]interface{}{
						"logDriver": "awslogs",
						"options": map[string]interface{}{
							"awslogs-group":         logGroup,
							"awslogs-region":        "us-east-1",
							"awslogs-stream-prefix": "supertokens",
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

	return ecs.NewTaskDefinition(ctx, fmt.Sprintf("%s-%s-supertokens-task", projectName, environment), &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(fmt.Sprintf("%s-%s-supertokens", projectName, environment)),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		Cpu:                     pulumi.String("512"),
		Memory:                  pulumi.String("1024"),
		ExecutionRoleArn:        roles.TaskExecutionRoleArn,
		TaskRoleArn:             roles.TaskRoleArn,
		ContainerDefinitions:    containerDef,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-supertokens-task", projectName, environment)),
			"Service":     pulumi.String("supertokens"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
}

func createAPITaskDefinition(ctx *pulumi.Context, projectName, environment string, roles *IAMRoles,
	logs *LogGroups, repositories *ECRRepositories, database *DatabaseResources, redis *RedisResources,
	secrets *SecretsResources, namespace *servicediscovery.PrivateDnsNamespace, tags pulumi.StringMap) (*ecs.TaskDefinition, error) {

	containerDef := pulumi.All(repositories.APIRepoURL, database.ClusterEndpoint, redis.Endpoint,
		secrets.DatabaseSecretArn, logs.APILogGroup.Name, namespace.Name).ApplyT(
		func(args []interface{}) (string, error) {
			imageURL := args[0].(string)
			dbEndpoint := args[1].(string)
			redisEndpoint := args[2].(string)
			secretArn := args[3].(string)
			logGroup := args[4].(string)
			namespaceName := args[5].(string)

			containers := []map[string]interface{}{
				{
					"name":  "api",
					"image": fmt.Sprintf("%s:latest", imageURL),
					"portMappings": []map[string]interface{}{
						{
							"containerPort": 8080,
							"protocol":      "tcp",
						},
					},
					"environment": []map[string]interface{}{
						{"name": "APP_ENV", "value": environment},
						{"name": "APP_PORT", "value": "8080"},
						{"name": "DB_HOST", "value": dbEndpoint},
						{"name": "DB_PORT", "value": "5432"},
						{"name": "DB_USER", "value": database.MasterUsername},
						{"name": "DB_NAME", "value": database.MainDBName},
						{"name": "DB_SSL_MODE", "value": "require"},
						{"name": "REDIS_HOST", "value": redisEndpoint},
						{"name": "REDIS_PORT", "value": "6379"},
						{"name": "SUPERTOKENS_CONNECTION_URI", "value": fmt.Sprintf("http://supertokens.%s:3567", namespaceName)},
						{"name": "LOG_LEVEL", "value": "info"},
						{"name": "LOG_FORMAT", "value": "json"},
					},
					"secrets": []map[string]interface{}{
						{
							"name":      "DB_PASSWORD",
							"valueFrom": fmt.Sprintf("%s:password::", secretArn),
						},
						{
							"name":      "SUPERTOKENS_API_KEY",
							"valueFrom": fmt.Sprintf("%s:api_key::", secretArn),
						},
					},
					"logConfiguration": map[string]interface{}{
						"logDriver": "awslogs",
						"options": map[string]interface{}{
							"awslogs-group":         logGroup,
							"awslogs-region":        "us-east-1",
							"awslogs-stream-prefix": "api",
						},
					},
					"healthCheck": map[string]interface{}{
						"command":     []string{"CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"},
						"interval":    30,
						"timeout":     5,
						"retries":     3,
						"startPeriod": 60,
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

	return ecs.NewTaskDefinition(ctx, fmt.Sprintf("%s-%s-api-task", projectName, environment), &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(fmt.Sprintf("%s-%s-api", projectName, environment)),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		Cpu:                     pulumi.String("512"),
		Memory:                  pulumi.String("1024"),
		ExecutionRoleArn:        roles.TaskExecutionRoleArn,
		TaskRoleArn:             roles.TaskRoleArn,
		ContainerDefinitions:    containerDef,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-api-task", projectName, environment)),
			"Service":     pulumi.String("api"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
}

func createWorkerTaskDefinition(ctx *pulumi.Context, projectName, environment string, roles *IAMRoles,
	logs *LogGroups, repositories *ECRRepositories, database *DatabaseResources, redis *RedisResources,
	secrets *SecretsResources, tags pulumi.StringMap) (*ecs.TaskDefinition, error) {

	containerDef := pulumi.All(repositories.WorkerRepoURL, database.ClusterEndpoint, redis.Endpoint,
		secrets.DatabaseSecretArn, logs.WorkerLogGroup.Name).ApplyT(
		func(args []interface{}) (string, error) {
			imageURL := args[0].(string)
			dbEndpoint := args[1].(string)
			redisEndpoint := args[2].(string)
			secretArn := args[3].(string)
			logGroup := args[4].(string)

			containers := []map[string]interface{}{
				{
					"name":  "worker",
					"image": fmt.Sprintf("%s:latest", imageURL),
					"environment": []map[string]interface{}{
						{"name": "APP_ENV", "value": environment},
						{"name": "DB_HOST", "value": dbEndpoint},
						{"name": "DB_PORT", "value": "5432"},
						{"name": "DB_USER", "value": database.MasterUsername},
						{"name": "DB_NAME", "value": database.MainDBName},
						{"name": "DB_SSL_MODE", "value": "require"},
						{"name": "REDIS_HOST", "value": redisEndpoint},
						{"name": "REDIS_PORT", "value": "6379"},
						{"name": "LOG_LEVEL", "value": "info"},
						{"name": "LOG_FORMAT", "value": "json"},
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
							"awslogs-region":        "us-east-1",
							"awslogs-stream-prefix": "worker",
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

	return ecs.NewTaskDefinition(ctx, fmt.Sprintf("%s-%s-worker-task", projectName, environment), &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(fmt.Sprintf("%s-%s-worker", projectName, environment)),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		Cpu:                     pulumi.String("512"),
		Memory:                  pulumi.String("1024"),
		ExecutionRoleArn:        roles.TaskExecutionRoleArn,
		TaskRoleArn:             roles.TaskRoleArn,
		ContainerDefinitions:    containerDef,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-worker-task", projectName, environment)),
			"Service":     pulumi.String("worker"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
}

// Frontend is now deployed via AWS Amplify instead of ECS
// See amplify.go for frontend deployment configuration
// This function has been removed as it's no longer needed
