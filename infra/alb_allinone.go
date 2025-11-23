package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AllInOneALBResources struct {
	ALB                    *lb.LoadBalancer
	APITargetGroup         *lb.TargetGroup
	SuperTokensTargetGroup *lb.TargetGroup
	HTTPListener           *lb.Listener
	HTTPSListener          *lb.Listener
	DNSName                pulumi.StringOutput
	ARN                    pulumi.StringOutput
}

// createAllInOneALB creates an ALB specifically for the all-in-one EC2 deployment
// Target groups are configured for instance targets, not IP targets
func createAllInOneALB(ctx *pulumi.Context, projectName, environment string,
	network *NetworkingResources, securityGroups *SecurityGroups,
	instance *ec2.Instance, certificateArn string, tags pulumi.StringMap) (*AllInOneALBResources, error) {

	// Create Application Load Balancer
	alb, err := lb.NewLoadBalancer(ctx, fmt.Sprintf("%s-%s-allinone-alb", projectName, environment), &lb.LoadBalancerArgs{
		Name:                     pulumi.String(fmt.Sprintf("%s-%s-alb", projectName, environment)),
		Internal:                 pulumi.Bool(false),
		LoadBalancerType:         pulumi.String("application"),
		SecurityGroups:           pulumi.StringArray{securityGroups.ALBSG.ID().ToStringOutput()},
		Subnets:                  network.PublicSubnetIDs,
		EnableDeletionProtection: pulumi.Bool(false), // Always false for all-in-one (dev/test)
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-alb", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
			"Mode":        pulumi.String("allinone"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Create API Target Group (for EC2 instance)
	apiTargetGroup, err := lb.NewTargetGroup(ctx, fmt.Sprintf("%s-%s-allinone-api-tg", projectName, environment), &lb.TargetGroupArgs{
		Name:       pulumi.String(fmt.Sprintf("%s-%s-api-tg", projectName, environment)),
		Port:       pulumi.Int(8080),
		Protocol:   pulumi.String("HTTP"),
		VpcId:      network.VpcID.ToStringOutput(),
		TargetType: pulumi.String("instance"), // EC2 instance target
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
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
			"Mode":        pulumi.String("allinone"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Create SuperTokens Target Group (for EC2 instance)
	superTokensTargetGroup, err := lb.NewTargetGroup(ctx, fmt.Sprintf("%s-%s-allinone-supertokens-tg", projectName, environment), &lb.TargetGroupArgs{
		Name:       pulumi.String(fmt.Sprintf("%s-%s-st-tg", projectName, environment)),
		Port:       pulumi.Int(3567),
		Protocol:   pulumi.String("HTTP"),
		VpcId:      network.VpcID.ToStringOutput(),
		TargetType: pulumi.String("instance"), // EC2 instance target
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
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
			"Mode":        pulumi.String("allinone"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Register the EC2 instance with both target groups
	_, err = lb.NewTargetGroupAttachment(ctx, fmt.Sprintf("%s-%s-api-attachment", projectName, environment), &lb.TargetGroupAttachmentArgs{
		TargetGroupArn: apiTargetGroup.Arn,
		TargetId:       instance.ID(),
		Port:           pulumi.Int(8080),
	})
	if err != nil {
		return nil, err
	}

	_, err = lb.NewTargetGroupAttachment(ctx, fmt.Sprintf("%s-%s-st-attachment", projectName, environment), &lb.TargetGroupAttachmentArgs{
		TargetGroupArn: superTokensTargetGroup.Arn,
		TargetId:       instance.ID(),
		Port:           pulumi.Int(3567),
	})
	if err != nil {
		return nil, err
	}

	// Create HTTP Listener
	httpListener, err := lb.NewListener(ctx, fmt.Sprintf("%s-%s-allinone-http-listener", projectName, environment), &lb.ListenerArgs{
		LoadBalancerArn: alb.Arn,
		Port:            pulumi.Int(80),
		Protocol:        pulumi.String("HTTP"),
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type: pulumi.String("fixed-response"),
				FixedResponse: &lb.ListenerDefaultActionFixedResponseArgs{
					ContentType: pulumi.String("text/plain"),
					MessageBody: pulumi.String("Not Found"),
					StatusCode:  pulumi.String("404"),
				},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-http-listener", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create API Listener Rule
	_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-allinone-api-rule", projectName, environment), &lb.ListenerRuleArgs{
		ListenerArn: httpListener.Arn,
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
		return nil, err
	}

	// Create SuperTokens Listener Rule
	_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-allinone-st-rule", projectName, environment), &lb.ListenerRuleArgs{
		ListenerArn: httpListener.Arn,
		Priority:    pulumi.Int(200),
		Actions: lb.ListenerRuleActionArray{
			&lb.ListenerRuleActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: superTokensTargetGroup.Arn,
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
		return nil, err
	}

	// Optional: Create HTTPS Listener if certificate is provided
	var httpsListener *lb.Listener
	if certificateArn != "" {
		httpsListener, err = lb.NewListener(ctx, fmt.Sprintf("%s-%s-allinone-https-listener", projectName, environment), &lb.ListenerArgs{
			LoadBalancerArn: alb.Arn,
			Port:            pulumi.Int(443),
			Protocol:        pulumi.String("HTTPS"),
			CertificateArn:  pulumi.String(certificateArn),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type: pulumi.String("fixed-response"),
					FixedResponse: &lb.ListenerDefaultActionFixedResponseArgs{
						ContentType: pulumi.String("text/plain"),
						MessageBody: pulumi.String("Not Found"),
						StatusCode:  pulumi.String("404"),
					},
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String(fmt.Sprintf("%s-%s-https-listener", projectName, environment)),
				"Project":     tags["Project"],
				"Environment": tags["Environment"],
				"ManagedBy":   tags["ManagedBy"],
			},
		})
		if err != nil {
			return nil, err
		}

		// HTTPS API Rule
		_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-allinone-https-api-rule", projectName, environment), &lb.ListenerRuleArgs{
			ListenerArn: httpsListener.Arn,
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
				"Name":        pulumi.String(fmt.Sprintf("%s-%s-https-api-rule", projectName, environment)),
				"Project":     tags["Project"],
				"Environment": tags["Environment"],
				"ManagedBy":   tags["ManagedBy"],
			},
		})
		if err != nil {
			return nil, err
		}

		// HTTPS SuperTokens Rule
		_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-allinone-https-st-rule", projectName, environment), &lb.ListenerRuleArgs{
			ListenerArn: httpsListener.Arn,
			Priority:    pulumi.Int(200),
			Actions: lb.ListenerRuleActionArray{
				&lb.ListenerRuleActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: superTokensTargetGroup.Arn,
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
				"Name":        pulumi.String(fmt.Sprintf("%s-%s-https-st-rule", projectName, environment)),
				"Project":     tags["Project"],
				"Environment": tags["Environment"],
				"ManagedBy":   tags["ManagedBy"],
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return &AllInOneALBResources{
		ALB:                    alb,
		APITargetGroup:         apiTargetGroup,
		SuperTokensTargetGroup: superTokensTargetGroup,
		HTTPListener:           httpListener,
		HTTPSListener:          httpsListener,
		DNSName:                alb.DnsName,
		ARN:                    alb.Arn,
	}, nil
}

