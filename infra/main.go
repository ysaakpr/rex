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

		// Low-cost mode configuration
		lowCostStr := cfg.Get("lowcost")
		lowCost := lowCostStr == "true"

		// All-in-one mode configuration (overrides lowcost)
		allInOneStr := cfg.Get("allinone")
		allInOne := allInOneStr == "true"

		// Tags to apply to all resources
		tags := pulumi.StringMap{
			"Project":     pulumi.String(projectName),
			"Environment": pulumi.String(environment),
			"ManagedBy":   pulumi.String("pulumi"),
		}

		// Check for all-in-one mode first (simplest architecture)
		if allInOne {
			ctx.Log.Info("All-in-one mode enabled: deploying everything on single EC2 Spot instance with Docker Compose (no ALB, no NAT)", nil)

			// Create VPC and networking (without NAT Gateway for all-in-one)
			network, err := createNetworkingSimple(ctx, projectName, environment, vpcCidr, tags)
			if err != nil {
				return err
			}

			// Create ECR repositories first (needed for Docker images)
			repositories, err := createECRRepositories(ctx, projectName, environment, tags)
			if err != nil {
				return err
			}

			// Create all-in-one EC2 instance with Docker Compose
			// Note: No ALB or NAT Gateway needed
			allInOneRes, err := createAllInOneEC2(ctx, projectName, environment, network,
				dbMasterUsername, dbMasterPassword, supertokensApiKey, repositories, tags)
			if err != nil {
				return err
			}

			// Export all-in-one instance info
			ctx.Export("deploymentMode", pulumi.String("allinone"))
			ctx.Export("allInOneInstanceId", allInOneRes.Instance.ID())
			ctx.Export("allInOnePublicIp", allInOneRes.PublicIP)
			ctx.Export("allInOnePrivateIp", allInOneRes.PrivateIP)
			ctx.Export("allInOnePublicDns", allInOneRes.PublicDNS)
			ctx.Export("apiRepositoryUrl", repositories.APIRepoURL)
			ctx.Export("workerRepositoryUrl", repositories.WorkerRepoURL)

			// Service URLs (via nginx reverse proxy with HTTPS using Elastic IP)
			ctx.Export("apiUrl", pulumi.Sprintf("https://%s/api", allInOneRes.PublicIP))
			ctx.Export("baseUrl", pulumi.Sprintf("https://%s", allInOneRes.PublicIP))
			ctx.Export("httpUrl", pulumi.Sprintf("http://%s", allInOneRes.PublicIP))
			ctx.Export("elasticIp", allInOneRes.PublicIP)

			// Connection instructions
			ctx.Export("connectionInfo", pulumi.Sprintf(`
All-in-One Deployment - Nginx Reverse Proxy with HTTPS:
- Elastic IP (static): %s
- Base URL: https://%s
- API: https://%s/api
- SuperTokens: https://%s/auth
- Health Check: http://%s/health (HTTP allowed for monitoring)

Elastic IP Benefits:
  ✓ Static IP persists across instance restarts/replacements
  ✓ Spot instance interruptions don't change your IP
  ✓ Bookmark and share without URL changes
  ✓ No NAT Gateway needed (saves $32/month)

Nginx configuration:
  - HTTP (port 80): Redirects to HTTPS (except /health and Let's Encrypt)
  - HTTPS (port 443): Main application access
  - Internal proxying:
    /api/* → api:8080 (HTTP within Docker network)
    /auth/* → supertokens:3567 (HTTP within Docker network)

SSL Certificate:
  - Self-signed certificate for IP address (browser will warn)
  - For trusted cert, use Let's Encrypt IP certificate support

SSH Access: aws ssm start-session --target %s
Or use Elastic IP: %s
`, allInOneRes.PublicIP, allInOneRes.PublicIP, allInOneRes.PublicIP, allInOneRes.PublicIP, allInOneRes.PublicIP,
				allInOneRes.Instance.ID(), allInOneRes.PublicIP))

			ctx.Export("frontendUrl", pulumi.String("Not deployed - deploy separately or use manual Amplify setup"))

			return nil
		}

		// For standard/low-cost modes, create full networking with NAT Gateway
		ctx.Log.Info("Creating VPC with NAT Gateway (required for private subnets)", nil)

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

		// Database and Redis endpoints (will be populated based on mode)
		var databaseEndpoint pulumi.StringOutput
		var redisEndpoint pulumi.StringOutput
		var databaseMasterUsername string
		var databaseMainDBName string
		var databaseSuperTokensDBName string

		if lowCost {
			// Create low-cost EC2 instance with PostgreSQL and Redis
			ctx.Log.Info("Low-cost mode enabled: deploying self-hosted PostgreSQL and Redis on EC2 Spot instance", nil)

			lowCostResources, err := createLowCostEC2(ctx, projectName, environment, network, securityGroups.ECSSG,
				dbMasterUsername, dbMasterPassword, tags)
			if err != nil {
				return err
			}

			databaseEndpoint = lowCostResources.DatabaseHost
			redisEndpoint = lowCostResources.RedisHost
			databaseMasterUsername = lowCostResources.MasterUsername
			databaseMainDBName = lowCostResources.MainDBName
			databaseSuperTokensDBName = lowCostResources.SuperTokensDBName

			// Export low-cost instance info
			ctx.Export("lowCostInstanceId", lowCostResources.Instance.ID())
			ctx.Export("lowCostInstancePrivateIp", lowCostResources.Instance.PrivateIp)
		} else {
			// Create Aurora RDS Cluster (with 2 databases)
			ctx.Log.Info("Standard mode: deploying managed RDS Aurora and ElastiCache", nil)

			database, err := createDatabase(ctx, projectName, environment, network, securityGroups, dbMasterUsername, dbMasterPassword, tags)
			if err != nil {
				return err
			}

			// Create ElastiCache Redis
			redis, err := createRedis(ctx, projectName, environment, network, securityGroups, tags)
			if err != nil {
				return err
			}

			databaseEndpoint = database.ClusterEndpoint
			redisEndpoint = redis.Endpoint
			databaseMasterUsername = database.MasterUsername
			databaseMainDBName = database.MainDBName
			databaseSuperTokensDBName = database.SuperTokensDBName

			// Export standard mode info
			ctx.Export("rdsClusterEndpoint", database.ClusterEndpoint)
			ctx.Export("rdsReaderEndpoint", database.ReaderEndpoint)
		}

		// Create Secrets Manager secrets with the endpoints we have
		secretsInput := &SecretsInput{
			DatabaseEndpoint:  databaseEndpoint,
			RedisEndpoint:     redisEndpoint,
			MasterUsername:    databaseMasterUsername,
			MasterPassword:    dbMasterPassword,
			MainDBName:        databaseMainDBName,
			SuperTokensDBName: databaseSuperTokensDBName,
			SuperTokensAPIKey: supertokensApiKey,
		}
		secrets, err := createSecretsFromInput(ctx, projectName, environment, secretsInput, tags)
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
		servicesInput := &ECSServicesInput{
			DatabaseEndpoint:  databaseEndpoint,
			RedisEndpoint:     redisEndpoint,
			MasterUsername:    databaseMasterUsername,
			MainDBName:        databaseMainDBName,
			SuperTokensDBName: databaseSuperTokensDBName,
			MasterPassword:    dbMasterPassword,
			SuperTokensAPIKey: supertokensApiKey,
		}
		services, err := createECSServicesFromInput(ctx, projectName, environment, cluster, network, securityGroups,
			alb, roles, logs, repositories, secrets, servicesInput, tags)
		if err != nil {
			return err
		}

		// Create Amplify App for Frontend (only if GitHub token is provided)
		var amplifyApp *AmplifyResources
		if githubToken != "" {
			ctx.Log.Info("GitHub token provided: deploying Amplify frontend", nil)
			amplifyApp, err = createAmplifyApp(ctx, projectName, environment, alb, githubRepo, githubBranch, githubToken, tags)
			if err != nil {
				return err
			}
		} else {
			ctx.Log.Warn("No GitHub token provided: skipping Amplify deployment. Set rex-backend:githubToken to deploy frontend.", nil)
		}

		// Create migration task definition
		migrationInput := &MigrationInput{
			DatabaseEndpoint: databaseEndpoint,
			MasterUsername:   databaseMasterUsername,
			MainDBName:       databaseMainDBName,
			MasterPassword:   dbMasterPassword,
		}
		migrationTask, err := createMigrationTaskFromInput(ctx, projectName, environment, network, roles, logs,
			repositories, secrets, migrationInput, tags)
		if err != nil {
			return err
		}

		// Export important values
		ctx.Export("vpcId", network.VpcID)
		ctx.Export("publicSubnetIds", network.PublicSubnetIDs)
		ctx.Export("privateSubnetIds", network.PrivateSubnetIDs)
		ctx.Export("albDnsName", alb.DNSName)
		ctx.Export("albArn", alb.ARN)
		ctx.Export("databaseEndpoint", databaseEndpoint)
		ctx.Export("redisEndpoint", redisEndpoint)
		ctx.Export("deploymentMode", pulumi.String(map[bool]string{true: "lowcost", false: "managed"}[lowCost]))
		ctx.Export("ecsClusterName", cluster.ClusterName)
		ctx.Export("ecsClusterArn", cluster.ClusterARN)
		ctx.Export("apiServiceName", services.APIServiceName)
		ctx.Export("workerServiceName", services.WorkerServiceName)
		ctx.Export("supertokensServiceName", services.SuperTokensServiceName)
		ctx.Export("migrationTaskDefinitionArn", migrationTask.TaskDefinitionARN)
		ctx.Export("apiRepositoryUrl", repositories.APIRepoURL)
		ctx.Export("workerRepositoryUrl", repositories.WorkerRepoURL)

		// Amplify exports (only if deployed)
		if amplifyApp != nil {
			ctx.Export("amplifyAppId", amplifyApp.AppID)
			ctx.Export("amplifyAppArn", amplifyApp.AppARN)
			ctx.Export("amplifyDefaultDomain", amplifyApp.DefaultDomain)
			ctx.Export("amplifyBranchUrl", amplifyApp.BranchURL)
			ctx.Export("frontendUrl", amplifyApp.BranchURL)
		} else {
			ctx.Export("frontendUrl", pulumi.String("Not deployed - set rex-backend:githubToken to deploy"))
		}

		// API URL for reference
		if domainName != "" {
			ctx.Export("apiUrl", pulumi.Sprintf("https://%s/api", domainName))
		} else {
			ctx.Export("apiUrl", pulumi.Sprintf("http://%s/api", alb.DNSName))
		}

		return nil
	})
}
