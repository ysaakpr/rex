package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SecurityGroups struct {
	ALBSG         *ec2.SecurityGroup
	ECSSG         *ec2.SecurityGroup
	RdsSG         *ec2.SecurityGroup
	RedisSG       *ec2.SecurityGroup
	SuperTokensSG *ec2.SecurityGroup
}

func createSecurityGroups(ctx *pulumi.Context, projectName, environment string, vpcID pulumi.IDOutput, tags pulumi.StringMap) (*SecurityGroups, error) {
	// ALB Security Group
	albSG, err := ec2.NewSecurityGroup(ctx, fmt.Sprintf("%s-%s-alb-sg", projectName, environment), &ec2.SecurityGroupArgs{
		VpcId:       vpcID,
		Description: pulumi.String("Security group for Application Load Balancer"),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(80),
				ToPort:     pulumi.Int(80),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
			&ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(443),
				ToPort:     pulumi.Int(443),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-alb-sg", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// ECS Security Group (for all ECS tasks)
	ecsSG, err := ec2.NewSecurityGroup(ctx, fmt.Sprintf("%s-%s-ecs-sg", projectName, environment), &ec2.SecurityGroupArgs{
		VpcId:       vpcID,
		Description: pulumi.String("Security group for ECS tasks"),
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-ecs-sg", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Allow ALB to access ECS tasks
	_, err = ec2.NewSecurityGroupRule(ctx, fmt.Sprintf("%s-%s-ecs-from-alb", projectName, environment), &ec2.SecurityGroupRuleArgs{
		Type:                  pulumi.String("ingress"),
		FromPort:              pulumi.Int(0),
		ToPort:                pulumi.Int(65535),
		Protocol:              pulumi.String("tcp"),
		SourceSecurityGroupId: albSG.ID(),
		SecurityGroupId:       ecsSG.ID(),
	})
	if err != nil {
		return nil, err
	}

	// Allow ECS tasks to communicate with each other
	_, err = ec2.NewSecurityGroupRule(ctx, fmt.Sprintf("%s-%s-ecs-internal", projectName, environment), &ec2.SecurityGroupRuleArgs{
		Type:            pulumi.String("ingress"),
		FromPort:        pulumi.Int(0),
		ToPort:          pulumi.Int(65535),
		Protocol:        pulumi.String("tcp"),
		Self:            pulumi.Bool(true),
		SecurityGroupId: ecsSG.ID(),
	})
	if err != nil {
		return nil, err
	}

	// RDS Security Group
	rdsSG, err := ec2.NewSecurityGroup(ctx, fmt.Sprintf("%s-%s-rds-sg", projectName, environment), &ec2.SecurityGroupArgs{
		VpcId:       vpcID,
		Description: pulumi.String("Security group for RDS Aurora"),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:       pulumi.String("tcp"),
				FromPort:       pulumi.Int(5432),
				ToPort:         pulumi.Int(5432),
				SecurityGroups: pulumi.StringArray{ecsSG.ID().ToStringOutput()},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-rds-sg", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Redis Security Group
	redisSG, err := ec2.NewSecurityGroup(ctx, fmt.Sprintf("%s-%s-redis-sg", projectName, environment), &ec2.SecurityGroupArgs{
		VpcId:       vpcID,
		Description: pulumi.String("Security group for ElastiCache Redis"),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:       pulumi.String("tcp"),
				FromPort:       pulumi.Int(6379),
				ToPort:         pulumi.Int(6379),
				SecurityGroups: pulumi.StringArray{ecsSG.ID().ToStringOutput()},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-redis-sg", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// SuperTokens Security Group (if needed separately)
	supertokensSG, err := ec2.NewSecurityGroup(ctx, fmt.Sprintf("%s-%s-supertokens-sg", projectName, environment), &ec2.SecurityGroupArgs{
		VpcId:       vpcID,
		Description: pulumi.String("Security group for SuperTokens service"),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:       pulumi.String("tcp"),
				FromPort:       pulumi.Int(3567),
				ToPort:         pulumi.Int(3567),
				SecurityGroups: pulumi.StringArray{ecsSG.ID().ToStringOutput()},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-supertokens-sg", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	return &SecurityGroups{
		ALBSG:         albSG,
		ECSSG:         ecsSG,
		RdsSG:         rdsSG,
		RedisSG:       redisSG,
		SuperTokensSG: supertokensSG,
	}, nil
}
