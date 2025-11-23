package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/elasticache"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type RedisResources struct {
	SubnetGroup      *elasticache.SubnetGroup
	ReplicationGroup *elasticache.ReplicationGroup
	Endpoint         pulumi.StringOutput
}

func createRedis(ctx *pulumi.Context, projectName, environment string, network *NetworkingResources,
	securityGroups *SecurityGroups, tags pulumi.StringMap) (*RedisResources, error) {

	// Create ElastiCache Subnet Group
	subnetGroup, err := elasticache.NewSubnetGroup(ctx, fmt.Sprintf("%s-%s-redis-subnet", projectName, environment), &elasticache.SubnetGroupArgs{
		SubnetIds: network.PrivateSubnetIDs,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-redis-subnet", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create Redis Replication Group (cluster mode disabled for simplicity)
	replicationGroup, err := elasticache.NewReplicationGroup(ctx, fmt.Sprintf("%s-%s-redis", projectName, environment), &elasticache.ReplicationGroupArgs{
		ReplicationGroupId:       pulumi.String(fmt.Sprintf("%s-%s-redis", projectName, environment)),
		Description:              pulumi.String("Redis cluster for Rex Backend"),
		Engine:                   pulumi.String("redis"),
		EngineVersion:            pulumi.String("7.0"),
		NodeType:                 pulumi.String("cache.t4g.micro"), // Smallest node for dev
		NumCacheClusters:         pulumi.Int(1),                    // Single node for dev, increase for production
		Port:                     pulumi.Int(6379),
		SubnetGroupName:          subnetGroup.Name,
		SecurityGroupIds:         pulumi.StringArray{securityGroups.RedisSG.ID().ToStringOutput()},
		AtRestEncryptionEnabled:  pulumi.Bool(true),
		TransitEncryptionEnabled: pulumi.Bool(false), // Set to true for production with AUTH
		AutomaticFailoverEnabled: pulumi.Bool(false), // Set to true with multiple nodes
		SnapshotRetentionLimit:   pulumi.Int(5),
		SnapshotWindow:           pulumi.String("03:00-05:00"),
		ApplyImmediately:         pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-redis", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	return &RedisResources{
		SubnetGroup:      subnetGroup,
		ReplicationGroup: replicationGroup,
		Endpoint:         replicationGroup.PrimaryEndpointAddress,
	}, nil
}
