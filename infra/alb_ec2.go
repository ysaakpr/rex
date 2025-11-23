package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Create ALB target groups for EC2 instance
func createEC2TargetGroups(ctx *pulumi.Context, projectName, environment string, network *NetworkingResources,
	albResources *ALBResources, instance *ec2.Instance, tags pulumi.StringMap) error {

	// API Target Group (port 8080)
	apiTargetGroup, err := lb.NewTargetGroup(ctx, fmt.Sprintf("%s-%s-api-tg", projectName, environment), &lb.TargetGroupArgs{
		Name:       pulumi.String(fmt.Sprintf("%s-%s-api-tg", projectName, environment)),
		Port:       pulumi.Int(8080),
		Protocol:   pulumi.String("HTTP"),
		VpcId:      network.VpcID,
		TargetType: pulumi.String("instance"),
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Enabled:            pulumi.Bool(true),
			HealthyThreshold:   pulumi.Int(2),
			UnhealthyThreshold: pulumi.Int(3),
			Timeout:            pulumi.Int(5),
			Interval:           pulumi.Int(30),
			Path:               pulumi.String("/health"),
			Matcher:            pulumi.String("200"),
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-api-tg", projectName, environment)),
			"Service":     pulumi.String("api"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return err
	}

	// Register API instance
	_, err = lb.NewTargetGroupAttachment(ctx, fmt.Sprintf("%s-%s-api-attachment", projectName, environment), &lb.TargetGroupAttachmentArgs{
		TargetGroupArn: apiTargetGroup.Arn,
		TargetId:       instance.ID(),
		Port:           pulumi.Int(8080),
	})
	if err != nil {
		return err
	}

	// SuperTokens Target Group (port 3567)
	supertokensTargetGroup, err := lb.NewTargetGroup(ctx, fmt.Sprintf("%s-%s-supertokens-tg", projectName, environment), &lb.TargetGroupArgs{
		Name:       pulumi.String(fmt.Sprintf("%s-%s-st-tg", projectName, environment)),
		Port:       pulumi.Int(3567),
		Protocol:   pulumi.String("HTTP"),
		VpcId:      network.VpcID,
		TargetType: pulumi.String("instance"),
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Enabled:            pulumi.Bool(true),
			HealthyThreshold:   pulumi.Int(2),
			UnhealthyThreshold: pulumi.Int(3),
			Timeout:            pulumi.Int(5),
			Interval:           pulumi.Int(30),
			Path:               pulumi.String("/hello"),
			Matcher:            pulumi.String("200"),
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-st-tg", projectName, environment)),
			"Service":     pulumi.String("supertokens"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return err
	}

	// Register SuperTokens instance
	_, err = lb.NewTargetGroupAttachment(ctx, fmt.Sprintf("%s-%s-supertokens-attachment", projectName, environment), &lb.TargetGroupAttachmentArgs{
		TargetGroupArn: supertokensTargetGroup.Arn,
		TargetId:       instance.ID(),
		Port:           pulumi.Int(3567),
	})
	if err != nil {
		return err
	}

	// API Rule - /api/*
	_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-api-rule", projectName, environment), &lb.ListenerRuleArgs{
		ListenerArn: albResources.HTTPListener.Arn,
		Priority:    pulumi.Int(100),
		Actions: lb.ListenerRuleActionArray{
			&lb.ListenerRuleActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: apiTargetGroup.Arn,
			},
		},
		Conditions: lb.ListenerRuleConditionArray{
			&lb.ListenerRuleConditionArgs{
				PathPattern: &lb.ListenerRuleConditionPathPatternArgs{
					Values: pulumi.StringArray{pulumi.String("/api/*")},
				},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-api-rule", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return err
	}

	// SuperTokens Rule - /auth/*
	_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-supertokens-rule", projectName, environment), &lb.ListenerRuleArgs{
		ListenerArn: albResources.HTTPListener.Arn,
		Priority:    pulumi.Int(200),
		Actions: lb.ListenerRuleActionArray{
			&lb.ListenerRuleActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: supertokensTargetGroup.Arn,
			},
		},
		Conditions: lb.ListenerRuleConditionArray{
			&lb.ListenerRuleConditionArgs{
				PathPattern: &lb.ListenerRuleConditionPathPatternArgs{
					Values: pulumi.StringArray{pulumi.String("/auth/*")},
				},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-st-rule", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return err
	}

	return nil
}
