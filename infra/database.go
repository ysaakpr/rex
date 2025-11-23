package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DatabaseResources struct {
	Cluster           *rds.Cluster
	ClusterInstance   *rds.ClusterInstance
	ClusterEndpoint   pulumi.StringOutput
	ReaderEndpoint    pulumi.StringOutput
	MasterUsername    string
	MainDBName        string
	SuperTokensDBName string
}

func createDatabase(ctx *pulumi.Context, projectName, environment string, network *NetworkingResources,
	securityGroups *SecurityGroups, masterUsername string, masterPassword pulumi.StringOutput, tags pulumi.StringMap) (*DatabaseResources, error) {

	mainDBName := "rex_backend"
	supertokensDBName := "supertokens"

	// Create DB Subnet Group
	dbSubnetGroup, err := rds.NewSubnetGroup(ctx, fmt.Sprintf("%s-%s-db-subnet", projectName, environment), &rds.SubnetGroupArgs{
		SubnetIds: network.PrivateSubnetIDs,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-db-subnet", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Prepare cluster configuration
	clusterArgs := &rds.ClusterArgs{
		ClusterIdentifier:   pulumi.String(fmt.Sprintf("%s-%s-aurora-cluster", projectName, environment)),
		Engine:              pulumi.String("aurora-postgresql"),
		EngineMode:          pulumi.String("provisioned"),
		EngineVersion:       pulumi.String("14.10"), // PostgreSQL 14 - most stable for Serverless v2
		DatabaseName:        pulumi.String(mainDBName),
		MasterUsername:      pulumi.String(masterUsername),
		MasterPassword:      masterPassword,
		DbSubnetGroupName:   dbSubnetGroup.Name,
		VpcSecurityGroupIds: pulumi.StringArray{securityGroups.RdsSG.ID().ToStringOutput()},
		Serverlessv2ScalingConfiguration: &rds.ClusterServerlessv2ScalingConfigurationArgs{
			MaxCapacity: pulumi.Float64(2.0), // 2 ACUs max
			MinCapacity: pulumi.Float64(0.5), // 0.5 ACUs min
		},
		BackupRetentionPeriod: pulumi.Int(7),
		PreferredBackupWindow: pulumi.String("03:00-04:00"),
		SkipFinalSnapshot:     pulumi.Bool(environment == "dev"), // Skip snapshot in dev
		ApplyImmediately:      pulumi.Bool(true),
		EnabledCloudwatchLogsExports: pulumi.StringArray{
			pulumi.String("postgresql"),
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-aurora-cluster", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	}

	// Only set FinalSnapshotIdentifier if we're NOT skipping final snapshot
	if environment != "dev" {
		clusterArgs.FinalSnapshotIdentifier = pulumi.String(fmt.Sprintf("%s-%s-final-snapshot-%d", projectName, environment, 1))
	}

	// Create Aurora Serverless v2 Cluster
	cluster, err := rds.NewCluster(ctx, fmt.Sprintf("%s-%s-aurora-cluster", projectName, environment), clusterArgs)
	if err != nil {
		return nil, err
	}

	// Create Aurora Serverless v2 Instance
	clusterInstance, err := rds.NewClusterInstance(ctx, fmt.Sprintf("%s-%s-aurora-instance-1", projectName, environment), &rds.ClusterInstanceArgs{
		Identifier:         pulumi.String(fmt.Sprintf("%s-%s-aurora-instance-1", projectName, environment)),
		ClusterIdentifier:  cluster.ID(),
		InstanceClass:      pulumi.String("db.serverless"),
		Engine:             pulumi.String("aurora-postgresql"),
		EngineVersion:      pulumi.String("14.10"), // Match cluster version
		PubliclyAccessible: pulumi.Bool(false),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-aurora-instance-1", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	return &DatabaseResources{
		Cluster:           cluster,
		ClusterInstance:   clusterInstance,
		ClusterEndpoint:   cluster.Endpoint,
		ReaderEndpoint:    cluster.ReaderEndpoint,
		MasterUsername:    masterUsername,
		MainDBName:        mainDBName,
		SuperTokensDBName: supertokensDBName,
	}, nil
}
