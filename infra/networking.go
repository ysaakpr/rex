package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type NetworkingResources struct {
	VpcID            pulumi.IDOutput
	PublicSubnetIDs  pulumi.StringArrayOutput
	PrivateSubnetIDs pulumi.StringArrayOutput
	InternetGateway  *ec2.InternetGateway
	NatGateway       *ec2.NatGateway
}

func createNetworking(ctx *pulumi.Context, projectName, environment, vpcCidr string, tags pulumi.StringMap) (*NetworkingResources, error) {
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

	// Create Internet Gateway
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
	azs, err := ec2.GetAvailabilityZones(ctx, &ec2.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available"),
	})
	if err != nil {
		return nil, err
	}

	// Create public subnets (2 for high availability)
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

	// Create private subnets (2 for high availability)
	var privateSubnets []*ec2.Subnet
	var privateSubnetIDs []pulumi.IDOutput
	for i := 0; i < 2; i++ {
		subnet, err := ec2.NewSubnet(ctx, fmt.Sprintf("%s-%s-private-subnet-%d", projectName, environment, i+1), &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(fmt.Sprintf("10.0.%d.0/24", i+10)),
			AvailabilityZone: pulumi.String(azs.Names[i]),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String(fmt.Sprintf("%s-%s-private-subnet-%d", projectName, environment, i+1)),
				"Type":        pulumi.String("private"),
				"Project":     tags["Project"],
				"Environment": tags["Environment"],
				"ManagedBy":   tags["ManagedBy"],
			},
		})
		if err != nil {
			return nil, err
		}
		privateSubnets = append(privateSubnets, subnet)
		privateSubnetIDs = append(privateSubnetIDs, subnet.ID())
	}

	// Allocate Elastic IP for NAT Gateway
	eip, err := ec2.NewEip(ctx, fmt.Sprintf("%s-%s-nat-eip", projectName, environment), &ec2.EipArgs{
		Domain: pulumi.String("vpc"),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-nat-eip", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create NAT Gateway in first public subnet
	natGateway, err := ec2.NewNatGateway(ctx, fmt.Sprintf("%s-%s-nat-gateway", projectName, environment), &ec2.NatGatewayArgs{
		SubnetId:     publicSubnets[0].ID(),
		AllocationId: eip.ID(),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-nat-gateway", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create public route table
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

	// Create private route table
	privateRouteTable, err := ec2.NewRouteTable(ctx, fmt.Sprintf("%s-%s-private-rt", projectName, environment), &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock:    pulumi.String("0.0.0.0/0"),
				NatGatewayId: natGateway.ID(),
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-private-rt", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Associate private subnets with private route table
	for i, subnet := range privateSubnets {
		_, err = ec2.NewRouteTableAssociation(ctx, fmt.Sprintf("%s-%s-private-rta-%d", projectName, environment, i+1), &ec2.RouteTableAssociationArgs{
			SubnetId:     subnet.ID(),
			RouteTableId: privateRouteTable.ID(),
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

	// Convert IDOutputs to StringArrayOutput for private subnets
	privateSubnetIDStrings := make([]pulumi.StringOutput, len(privateSubnetIDs))
	for i, id := range privateSubnetIDs {
		privateSubnetIDStrings[i] = id.ToStringOutput()
	}
	privateSubnetIDsArray := pulumi.ToStringArrayOutput(privateSubnetIDStrings)

	return &NetworkingResources{
		VpcID:            vpc.ID(),
		PublicSubnetIDs:  publicSubnetIDsArray,
		PrivateSubnetIDs: privateSubnetIDsArray,
		InternetGateway:  igw,
		NatGateway:       natGateway,
	}, nil
}
