package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Simple networking for all-in-one mode (no NAT Gateway needed)
func createNetworkingSimple(ctx *pulumi.Context, projectName, environment, vpcCidr string, tags pulumi.StringMap) (*NetworkingResources, error) {
	// Create VPC
	vpc, err := ec2.NewVpc(ctx, fmt.Sprintf("%s-%s-vpc", projectName, environment), &ec2.VpcArgs{
		CidrBlock:          pulumi.String(vpcCidr),
		EnableDnsHostnames: pulumi.Bool(true),
		EnableDnsSupport:   pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-vpc", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create Internet Gateway (all we need for public subnet)
	igw, err := ec2.NewInternetGateway(ctx, fmt.Sprintf("%s-%s-igw", projectName, environment), &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-igw", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Get availability zones
	azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available"),
	})
	if err != nil {
		return nil, err
	}

	// Create public subnets only (2 for high availability)
	var publicSubnets []*ec2.Subnet
	var publicSubnetIDs []pulumi.IDOutput
	for i := 0; i < 2; i++ {
		subnet, err := ec2.NewSubnet(ctx, fmt.Sprintf("%s-%s-public-subnet-%d", projectName, environment, i+1), &ec2.SubnetArgs{
			VpcId:               vpc.ID(),
			CidrBlock:           pulumi.String(fmt.Sprintf("10.0.%d.0/24", i)),
			AvailabilityZone:    pulumi.String(azs.Names[i]),
			MapPublicIpOnLaunch: pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String(fmt.Sprintf("%s-%s-public-subnet-%d", projectName, environment, i+1)),
				"Type":        pulumi.String("public"),
				"Project":     tags["Project"],
				"Environment": tags["Environment"],
				"ManagedBy":   tags["ManagedBy"],
			},
		})
		if err != nil {
			return nil, err
		}
		publicSubnets = append(publicSubnets, subnet)
		publicSubnetIDs = append(publicSubnetIDs, subnet.ID())
	}

	// Create public route table (route to Internet Gateway)
	publicRouteTable, err := ec2.NewRouteTable(ctx, fmt.Sprintf("%s-%s-public-rt", projectName, environment), &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String("0.0.0.0/0"),
				GatewayId: igw.ID(),
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-public-rt", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Associate public subnets with public route table
	for i, subnet := range publicSubnets {
		_, err = ec2.NewRouteTableAssociation(ctx, fmt.Sprintf("%s-%s-public-rta-%d", projectName, environment, i+1), &ec2.RouteTableAssociationArgs{
			SubnetId:     subnet.ID(),
			RouteTableId: publicRouteTable.ID(),
		})
		if err != nil {
			return nil, err
		}
	}

	// Convert IDOutputs to StringArrayOutput for public subnets
	publicSubnetIDStrings := make([]pulumi.StringOutput, len(publicSubnetIDs))
	for i, id := range publicSubnetIDs {
		publicSubnetIDStrings[i] = id.ToStringOutput()
	}
	publicSubnetIDsArray := pulumi.ToStringArrayOutput(publicSubnetIDStrings)

	// No private subnets or NAT Gateway for all-in-one mode
	return &NetworkingResources{
		VpcID:            vpc.ID(),
		PublicSubnetIDs:  publicSubnetIDsArray,
		PrivateSubnetIDs: pulumi.ToStringArray([]string{}).ToStringArrayOutput(), // Empty array
		InternetGateway:  igw,
		NatGateway:       nil, // No NAT Gateway needed!
	}, nil
}
