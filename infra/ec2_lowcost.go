package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type LowCostResources struct {
	Instance          *ec2.Instance
	SecurityGroup     *ec2.SecurityGroup
	DatabaseHost      pulumi.StringOutput
	RedisHost         pulumi.StringOutput
	MasterUsername    string
	MainDBName        string
	SuperTokensDBName string
	MasterPassword    pulumi.StringOutput
}

func createLowCostEC2(ctx *pulumi.Context, projectName, environment string, network *NetworkingResources,
	ecsSG *ec2.SecurityGroup, masterUsername string, masterPassword pulumi.StringOutput, tags pulumi.StringMap) (*LowCostResources, error) {

	mainDBName := "rex_backend"
	supertokensDBName := "supertokens"

	// Create Security Group for the EC2 instance (PostgreSQL + Redis)
	ec2SG, err := ec2.NewSecurityGroup(ctx, fmt.Sprintf("%s-%s-lowcost-sg", projectName, environment), &ec2.SecurityGroupArgs{
		VpcId:       network.VpcID,
		Description: pulumi.String("Security group for low-cost EC2 instance (PostgreSQL + Redis)"),
		Ingress: ec2.SecurityGroupIngressArray{
			// PostgreSQL access from ECS
			&ec2.SecurityGroupIngressArgs{
				Protocol:       pulumi.String("tcp"),
				FromPort:       pulumi.Int(5432),
				ToPort:         pulumi.Int(5432),
				SecurityGroups: pulumi.StringArray{ecsSG.ID().ToStringOutput()},
				Description:    pulumi.String("PostgreSQL from ECS"),
			},
			// Redis access from ECS
			&ec2.SecurityGroupIngressArgs{
				Protocol:       pulumi.String("tcp"),
				FromPort:       pulumi.Int(6379),
				ToPort:         pulumi.Int(6379),
				SecurityGroups: pulumi.StringArray{ecsSG.ID().ToStringOutput()},
				Description:    pulumi.String("Redis from ECS"),
			},
			// SSH access (optional, for debugging)
			&ec2.SecurityGroupIngressArgs{
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(22),
				ToPort:      pulumi.Int(22),
				CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Description: pulumi.String("SSH access"),
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
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-lowcost-sg", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create IAM role for EC2 instance (for Systems Manager access)
	ec2Role, err := iam.NewRole(ctx, fmt.Sprintf("%s-%s-lowcost-role", projectName, environment), &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {
					"Service": "ec2.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}]
		}`),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-lowcost-role", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Attach SSM policy for Systems Manager access
	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-%s-lowcost-ssm-policy", projectName, environment), &iam.RolePolicyAttachmentArgs{
		Role:      ec2Role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
	})
	if err != nil {
		return nil, err
	}

	// Create instance profile
	instanceProfile, err := iam.NewInstanceProfile(ctx, fmt.Sprintf("%s-%s-lowcost-profile", projectName, environment), &iam.InstanceProfileArgs{
		Role: ec2Role.Name,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-lowcost-profile", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Get the latest Ubuntu 22.04 LTS AMI
	ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
		MostRecent: pulumi.BoolRef(true),
		Owners:     []string{"099720109477"}, // Canonical
		Filters: []ec2.GetAmiFilter{
			{
				Name:   "name",
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
			},
			{
				Name:   "virtualization-type",
				Values: []string{"hvm"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// User data script to install PostgreSQL and Redis
	userData := pulumi.All(masterPassword, pulumi.String(masterUsername), pulumi.String(mainDBName), pulumi.String(supertokensDBName)).
		ApplyT(func(args []interface{}) string {
			password := args[0].(string)
			username := args[1].(string)
			mainDB := args[2].(string)
			supertokensDB := args[3].(string)

			return fmt.Sprintf(`#!/bin/bash
set -e

# Update system
apt-get update
apt-get upgrade -y

# Install PostgreSQL 14
apt-get install -y postgresql-14 postgresql-contrib-14

# Configure PostgreSQL to listen on all interfaces
sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" /etc/postgresql/14/main/postgresql.conf

# Configure PostgreSQL authentication
cat > /etc/postgresql/14/main/pg_hba.conf <<EOF
# TYPE  DATABASE        USER            ADDRESS                 METHOD
local   all             postgres                                peer
local   all             all                                     peer
host    all             all             0.0.0.0/0               md5
host    all             all             ::/0                    md5
EOF

# Restart PostgreSQL
systemctl restart postgresql

# Create database user and databases
sudo -u postgres psql <<EOSQL
CREATE USER %s WITH PASSWORD '%s';
CREATE DATABASE %s OWNER %s;
CREATE DATABASE %s OWNER %s;
ALTER USER %s WITH SUPERUSER;
EOSQL

# Install Redis
apt-get install -y redis-server

# Configure Redis to listen on all interfaces
sed -i 's/bind 127.0.0.1 ::1/bind 0.0.0.0/' /etc/redis/redis.conf
# Disable protected mode (for private network)
sed -i 's/protected-mode yes/protected-mode no/' /etc/redis/redis.conf
# Set maxmemory policy
echo "maxmemory 512mb" >> /etc/redis/redis.conf
echo "maxmemory-policy allkeys-lru" >> /etc/redis/redis.conf

# Restart Redis
systemctl restart redis-server

# Enable services to start on boot
systemctl enable postgresql
systemctl enable redis-server

# Install CloudWatch agent for monitoring (optional)
wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
dpkg -i -E ./amazon-cloudwatch-agent.deb

echo "Setup complete!"
`, username, password, mainDB, username, supertokensDB, username, username)
		}).(pulumi.StringOutput)

	// Create EC2 Spot Instance
	// Deploy in first private subnet
	instance, err := ec2.NewInstance(ctx, fmt.Sprintf("%s-%s-lowcost-instance", projectName, environment), &ec2.InstanceArgs{
		Ami:                pulumi.String(ami.Id),
		InstanceType:       pulumi.String("t3a.small"), // 2 vCPU, 2 GB RAM
		IamInstanceProfile: instanceProfile.Name,
		SubnetId: network.PrivateSubnetIDs.ApplyT(func(subnets []string) string {
			return subnets[0]
		}).(pulumi.StringOutput),
		VpcSecurityGroupIds: pulumi.StringArray{ec2SG.ID().ToStringOutput()},
		UserData:            userData,

		// Spot instance configuration
		InstanceMarketOptions: &ec2.InstanceInstanceMarketOptionsArgs{
			MarketType: pulumi.String("spot"),
			SpotOptions: &ec2.InstanceInstanceMarketOptionsSpotOptionsArgs{
				MaxPrice:                     pulumi.String("0.05"),
				SpotInstanceType:             pulumi.String("persistent"),
				InstanceInterruptionBehavior: pulumi.String("stop"),
			},
		},

		// Root volume configuration
		RootBlockDevice: &ec2.InstanceRootBlockDeviceArgs{
			VolumeSize:          pulumi.Int(30),
			VolumeType:          pulumi.String("gp3"),
			DeleteOnTermination: pulumi.Bool(true),
		},

		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-lowcost-db-redis", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
			"Type":        pulumi.String("lowcost-database"),
		},
	})
	if err != nil {
		return nil, err
	}

	return &LowCostResources{
		Instance:          instance,
		SecurityGroup:     ec2SG,
		DatabaseHost:      instance.PrivateIp,
		RedisHost:         instance.PrivateIp,
		MasterUsername:    masterUsername,
		MainDBName:        mainDBName,
		SuperTokensDBName: supertokensDBName,
		MasterPassword:    masterPassword,
	}, nil
}
