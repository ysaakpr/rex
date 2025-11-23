package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ALBResources struct {
	ALB                    *lb.LoadBalancer
	APITargetGroup         *lb.TargetGroup
	SuperTokensTargetGroup *lb.TargetGroup
	HTTPListener           *lb.Listener
	HTTPSListener          *lb.Listener
	DNSName                pulumi.StringOutput
	ARN                    pulumi.StringOutput
}

func createLoadBalancer(ctx *pulumi.Context, projectName, environment string, network *NetworkingResources,
	securityGroups *SecurityGroups, certificateArn string, tags pulumi.StringMap) (*ALBResources, error) {

	// Create Application Load Balancer
	alb, err := lb.NewLoadBalancer(ctx, fmt.Sprintf("%s-%s-alb", projectName, environment), &lb.LoadBalancerArgs{
		Name:                     pulumi.String(fmt.Sprintf("%s-%s-alb", projectName, environment)),
		Internal:                 pulumi.Bool(false),
		LoadBalancerType:         pulumi.String("application"),
		SecurityGroups:           pulumi.StringArray{securityGroups.ALBSG.ID().ToStringOutput()},
		Subnets:                  network.PublicSubnetIDs,
		EnableDeletionProtection: pulumi.Bool(environment != "dev"), // Disable in dev
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-alb", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create Target Groups
	apiTargetGroup, err := lb.NewTargetGroup(ctx, fmt.Sprintf("%s-%s-api-tg", projectName, environment), &lb.TargetGroupArgs{
		Name:       pulumi.String(fmt.Sprintf("%s-%s-api-tg", projectName, environment)),
		Port:       pulumi.Int(8080),
		Protocol:   pulumi.String("HTTP"),
		VpcId:      network.VpcID.ToStringOutput(),
		TargetType: pulumi.String("ip"),
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Enabled:            pulumi.Bool(true),
			HealthyThreshold:   pulumi.Int(2),
			UnhealthyThreshold: pulumi.Int(3),
			Timeout:            pulumi.Int(5),
			Interval:           pulumi.Int(30),
			Path:               pulumi.String("/health"),
			Matcher:            pulumi.String("200"),
		},
		DeregistrationDelay: pulumi.Int(30),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-api-tg", projectName, environment)),
			"Service":     pulumi.String("api"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Note: Frontend target group removed - frontend is now served via AWS Amplify
	// ALB only handles backend API and auth routes

	supertokensTargetGroup, err := lb.NewTargetGroup(ctx, fmt.Sprintf("%s-%s-supertokens-tg", projectName, environment), &lb.TargetGroupArgs{
		Name:       pulumi.String(fmt.Sprintf("%s-%s-st-tg", projectName, environment)),
		Port:       pulumi.Int(3567),
		Protocol:   pulumi.String("HTTP"),
		VpcId:      network.VpcID.ToStringOutput(),
		TargetType: pulumi.String("ip"),
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Enabled:            pulumi.Bool(true),
			HealthyThreshold:   pulumi.Int(2),
			UnhealthyThreshold: pulumi.Int(3),
			Timeout:            pulumi.Int(5),
			Interval:           pulumi.Int(30),
			Path:               pulumi.String("/hello"),
			Matcher:            pulumi.String("200"),
		},
		DeregistrationDelay: pulumi.Int(30),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-supertokens-tg", projectName, environment)),
			"Service":     pulumi.String("supertokens"),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create HTTP Listener (always)
	// Default action returns fixed response since frontend is on Amplify
	httpListener, err := lb.NewListener(ctx, fmt.Sprintf("%s-%s-http-listener", projectName, environment), &lb.ListenerArgs{
		LoadBalancerArn: alb.Arn,
		Port:            pulumi.Int(80),
		Protocol:        pulumi.String("HTTP"),
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type: pulumi.String("fixed-response"),
				FixedResponse: &lb.ListenerDefaultActionFixedResponseArgs{
					ContentType: pulumi.String("text/plain"),
					MessageBody: pulumi.String("Backend API - Use /api or /auth paths"),
					StatusCode:  pulumi.String("200"),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Create listener rules for routing
	_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-api-rule", projectName, environment), &lb.ListenerRuleArgs{
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
					Values: pulumi.StringArray{
						pulumi.String("/api/*"),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-supertokens-rule", projectName, environment), &lb.ListenerRuleArgs{
		ListenerArn: httpListener.Arn,
		Priority:    pulumi.Int(90),
		Actions: lb.ListenerRuleActionArray{
			&lb.ListenerRuleActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: supertokensTargetGroup.Arn,
			},
		},
		Conditions: lb.ListenerRuleConditionArray{
			&lb.ListenerRuleConditionArgs{
				PathPattern: &lb.ListenerRuleConditionPathPatternArgs{
					Values: pulumi.StringArray{
						pulumi.String("/auth/*"),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// HTTPS Listener (if certificate provided)
	var httpsListener *lb.Listener
	if certificateArn != "" {
		httpsListener, err = lb.NewListener(ctx, fmt.Sprintf("%s-%s-https-listener", projectName, environment), &lb.ListenerArgs{
			LoadBalancerArn: alb.Arn,
			Port:            pulumi.Int(443),
			Protocol:        pulumi.String("HTTPS"),
			SslPolicy:       pulumi.String("ELBSecurityPolicy-TLS-1-2-2017-01"),
			CertificateArn:  pulumi.String(certificateArn),
			DefaultActions: lb.ListenerDefaultActionArray{
				&lb.ListenerDefaultActionArgs{
					Type: pulumi.String("fixed-response"),
					FixedResponse: &lb.ListenerDefaultActionFixedResponseArgs{
						ContentType: pulumi.String("text/plain"),
						MessageBody: pulumi.String("Backend API - Use /api or /auth paths"),
						StatusCode:  pulumi.String("200"),
					},
				},
			},
		})
		if err != nil {
			return nil, err
		}

		// Add routing rules for HTTPS
		_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-https-api-rule", projectName, environment), &lb.ListenerRuleArgs{
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
						Values: pulumi.StringArray{
							pulumi.String("/api/*"),
						},
					},
				},
			},
		})
		if err != nil {
			return nil, err
		}

		_, err = lb.NewListenerRule(ctx, fmt.Sprintf("%s-%s-https-supertokens-rule", projectName, environment), &lb.ListenerRuleArgs{
			ListenerArn: httpsListener.Arn,
			Priority:    pulumi.Int(90),
			Actions: lb.ListenerRuleActionArray{
				&lb.ListenerRuleActionArgs{
					Type:           pulumi.String("forward"),
					TargetGroupArn: supertokensTargetGroup.Arn,
				},
			},
			Conditions: lb.ListenerRuleConditionArray{
				&lb.ListenerRuleConditionArgs{
					PathPattern: &lb.ListenerRuleConditionPathPatternArgs{
						Values: pulumi.StringArray{
							pulumi.String("/auth/*"),
						},
					},
				},
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return &ALBResources{
		ALB:                    alb,
		APITargetGroup:         apiTargetGroup,
		SuperTokensTargetGroup: supertokensTargetGroup,
		HTTPListener:           httpListener,
		HTTPSListener:          httpsListener,
		DNSName:                alb.DnsName,
		ARN:                    alb.Arn,
	}, nil
}
