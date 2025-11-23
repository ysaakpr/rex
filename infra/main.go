package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load configuration
		cfg := config.New(ctx, "")
		projectName := cfg.Get("projectName")
		if projectName == "" {
			projectName = "rex-backend"
		}
		environment := cfg.Get("environment")
		if environment == "" {
			environment = "dev"
		}
		vpcCidr := cfg.Get("vpcCidr")
		if vpcCidr == "" {
			vpcCidr = "10.0.0.0/16"
		}

		// Get secrets
		dbMasterPassword := cfg.RequireSecret("dbMasterPassword")
		supertokensApiKey := cfg.RequireSecret("supertokensApiKey")

		dbMasterUsername := cfg.Get("dbMasterUsername")
		if dbMasterUsername == "" {
			dbMasterUsername = "rexadmin"
		}

		domainName := cfg.Get("domainName")
		certificateArn := cfg.Get("certificateArn")

		// GitHub configuration for Amplify
		githubRepo := cfg.Get("githubRepo")
		if githubRepo == "" {
			githubRepo = "https://github.com/yourusername/rex-backend" // Update this with your actual repo
		}
		githubBranch := cfg.Get("githubBranch")
		if githubBranch == "" {
			githubBranch = "main"
		}
		githubToken := cfg.Get("githubToken")
		if githubToken == "" {
			githubToken = "" // Optional for public repos
		}

		// Tags to apply to all resources
		tags := pulumi.StringMap{
			"Project":     pulumi.String(projectName),
			"Environment": pulumi.String(environment),
			"ManagedBy":   pulumi.String("pulumi"),
		}

		// Create VPC and networking
		network, err := createNetworking(ctx, projectName, environment, vpcCidr, tags)
		if err != nil {
			return err
		}

		// Create Security Groups
		securityGroups, err := createSecurityGroups(ctx, projectName, environment, network.VpcID, tags)
		if err != nil {
			return err
		}

		// Create Aurora RDS Cluster (with 2 databases)
		database, err := createDatabase(ctx, projectName, environment, network, securityGroups, dbMasterUsername, dbMasterPassword, tags)
		if err != nil {
			return err
		}

		// Create ElastiCache Redis
		redis, err := createRedis(ctx, projectName, environment, network, securityGroups, tags)
		if err != nil {
			return err
		}

		// Create Secrets Manager secrets
		secrets, err := createSecrets(ctx, projectName, environment, database, redis, dbMasterPassword, supertokensApiKey, tags)
		if err != nil {
			return err
		}

		// Create ECR repositories
		repositories, err := createECRRepositories(ctx, projectName, environment, tags)
		if err != nil {
			return err
		}

		// Create ECS Cluster
		cluster, err := createECSCluster(ctx, projectName, environment, tags)
		if err != nil {
			return err
		}

		// Create Application Load Balancer
		alb, err := createLoadBalancer(ctx, projectName, environment, network, securityGroups, certificateArn, tags)
		if err != nil {
			return err
		}

		// Create IAM roles for ECS tasks
		roles, err := createIAMRoles(ctx, projectName, environment, secrets, tags)
		if err != nil {
			return err
		}

		// Create CloudWatch log groups
		logs, err := createLogGroups(ctx, projectName, environment, tags)
		if err != nil {
			return err
		}

		// Create ECS Task Definitions and Services (backend only - no frontend)
		services, err := createECSServices(ctx, projectName, environment, cluster, network, securityGroups,
			alb, roles, logs, repositories, secrets, database, redis, supertokensApiKey, tags)
		if err != nil {
			return err
		}

		// Create Amplify App for Frontend
		amplifyApp, err := createAmplifyApp(ctx, projectName, environment, alb, githubRepo, githubBranch, githubToken, tags)
		if err != nil {
			return err
		}

		// Create migration task definition
		migrationTask, err := createMigrationTask(ctx, projectName, environment, network, roles, logs,
			repositories, secrets, database, tags)
		if err != nil {
			return err
		}

		// Export important values
		ctx.Export("vpcId", network.VpcID)
		ctx.Export("publicSubnetIds", network.PublicSubnetIDs)
		ctx.Export("privateSubnetIds", network.PrivateSubnetIDs)
		ctx.Export("albDnsName", alb.DNSName)
		ctx.Export("albArn", alb.ARN)
		ctx.Export("rdsClusterEndpoint", database.ClusterEndpoint)
		ctx.Export("rdsReaderEndpoint", database.ReaderEndpoint)
		ctx.Export("redisEndpoint", redis.Endpoint)
		ctx.Export("ecsClusterName", cluster.ClusterName)
		ctx.Export("ecsClusterArn", cluster.ClusterARN)
		ctx.Export("apiServiceName", services.APIServiceName)
		ctx.Export("workerServiceName", services.WorkerServiceName)
		ctx.Export("supertokensServiceName", services.SuperTokensServiceName)
		ctx.Export("migrationTaskDefinitionArn", migrationTask.TaskDefinitionARN)
		ctx.Export("apiRepositoryUrl", repositories.APIRepoURL)
		ctx.Export("workerRepositoryUrl", repositories.WorkerRepoURL)

		// Amplify exports
		ctx.Export("amplifyAppId", amplifyApp.AppID)
		ctx.Export("amplifyAppArn", amplifyApp.AppARN)
		ctx.Export("amplifyDefaultDomain", amplifyApp.DefaultDomain)
		ctx.Export("amplifyBranchUrl", amplifyApp.BranchURL)
		ctx.Export("frontendUrl", amplifyApp.BranchURL)

		// API URL for reference
		if domainName != "" {
			ctx.Export("apiUrl", pulumi.Sprintf("https://%s/api", domainName))
		} else {
			ctx.Export("apiUrl", pulumi.Sprintf("http://%s/api", alb.DNSName))
		}

		return nil
	})
}
