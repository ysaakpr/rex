package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/secretsmanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SecretsResources struct {
	DatabaseSecret       *secretsmanager.Secret
	DatabaseSecretArn    pulumi.StringOutput
	SuperTokensSecret    *secretsmanager.Secret
	SuperTokensSecretArn pulumi.StringOutput
}

type SecretsInput struct {
	DatabaseEndpoint  pulumi.StringOutput
	RedisEndpoint     pulumi.StringOutput
	MasterUsername    string
	MasterPassword    pulumi.StringOutput
	MainDBName        string
	SuperTokensDBName string
	SuperTokensAPIKey pulumi.StringOutput
}

func createSecrets(ctx *pulumi.Context, projectName, environment string, database *DatabaseResources,
	redis *RedisResources, dbPassword, supertokensApiKey pulumi.StringOutput, tags pulumi.StringMap) (*SecretsResources, error) {

	// Create database connection secret
	dbSecret, err := secretsmanager.NewSecret(ctx, fmt.Sprintf("%s-%s-db-secret", projectName, environment), &secretsmanager.SecretArgs{
		Name:        pulumi.String(fmt.Sprintf("%s-%s-db-secret", projectName, environment)),
		Description: pulumi.String("Database connection details"),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-db-secret", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Build database secret value
	dbSecretValue := pulumi.All(database.ClusterEndpoint, dbPassword, redis.Endpoint).ApplyT(
		func(args []interface{}) (string, error) {
			endpoint := args[0].(string)
			password := args[1].(string)
			redisEndpoint := args[2].(string)

			secretData := map[string]interface{}{
				"host":               endpoint,
				"port":               "5432",
				"username":           database.MasterUsername,
				"password":           password,
				"dbname":             database.MainDBName,
				"supertokens_dbname": database.SuperTokensDBName,
				"sslmode":            "require",
				"redis_host":         redisEndpoint,
				"redis_port":         "6379",
			}

			jsonData, err := json.Marshal(secretData)
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		},
	).(pulumi.StringOutput)

	_, err = secretsmanager.NewSecretVersion(ctx, fmt.Sprintf("%s-%s-db-secret-version", projectName, environment), &secretsmanager.SecretVersionArgs{
		SecretId:     dbSecret.ID(),
		SecretString: dbSecretValue,
	})
	if err != nil {
		return nil, err
	}

	// Create SuperTokens configuration secret
	superTokensSecret, err := secretsmanager.NewSecret(ctx, fmt.Sprintf("%s-%s-supertokens-secret", projectName, environment), &secretsmanager.SecretArgs{
		Name:        pulumi.String(fmt.Sprintf("%s-%s-supertokens-secret", projectName, environment)),
		Description: pulumi.String("SuperTokens configuration"),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-supertokens-secret", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	superTokensSecretValue := supertokensApiKey.ApplyT(func(apiKey string) (string, error) {
		secretData := map[string]interface{}{
			"api_key": apiKey,
		}
		jsonData, err := json.Marshal(secretData)
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
	}).(pulumi.StringOutput)

	_, err = secretsmanager.NewSecretVersion(ctx, fmt.Sprintf("%s-%s-supertokens-secret-version", projectName, environment), &secretsmanager.SecretVersionArgs{
		SecretId:     superTokensSecret.ID(),
		SecretString: superTokensSecretValue,
	})
	if err != nil {
		return nil, err
	}

	return &SecretsResources{
		DatabaseSecret:       dbSecret,
		DatabaseSecretArn:    dbSecret.Arn,
		SuperTokensSecret:    superTokensSecret,
		SuperTokensSecretArn: superTokensSecret.Arn,
	}, nil
}

func createSecretsFromInput(ctx *pulumi.Context, projectName, environment string, input *SecretsInput, tags pulumi.StringMap) (*SecretsResources, error) {

	// Create database connection secret
	// Use v2 suffix to avoid conflicts with deleted secrets
	dbSecret, err := secretsmanager.NewSecret(ctx, fmt.Sprintf("%s-%s-db-secret-v2", projectName, environment), &secretsmanager.SecretArgs{
		Name:                        pulumi.String(fmt.Sprintf("%s-%s-db-secret-v2", projectName, environment)),
		Description:                 pulumi.String("Database connection details (v2)"),
		RecoveryWindowInDays:        pulumi.Int(0), // Allow immediate deletion
		ForceOverwriteReplicaSecret: pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-db-secret-v2", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Build database secret value
	dbSecretValue := pulumi.All(input.DatabaseEndpoint, input.MasterPassword, input.RedisEndpoint).ApplyT(
		func(args []interface{}) (string, error) {
			endpoint := args[0].(string)
			password := args[1].(string)
			redisEndpoint := args[2].(string)

			secretData := map[string]interface{}{
				"host":               endpoint,
				"port":               "5432",
				"username":           input.MasterUsername,
				"password":           password,
				"dbname":             input.MainDBName,
				"supertokens_dbname": input.SuperTokensDBName,
				"sslmode":            "require",
				"redis_host":         redisEndpoint,
				"redis_port":         "6379",
			}

			jsonData, err := json.Marshal(secretData)
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		},
	).(pulumi.StringOutput)

	_, err = secretsmanager.NewSecretVersion(ctx, fmt.Sprintf("%s-%s-db-secret-version", projectName, environment), &secretsmanager.SecretVersionArgs{
		SecretId:     dbSecret.ID(),
		SecretString: dbSecretValue,
	})
	if err != nil {
		return nil, err
	}

	// Create SuperTokens configuration secret
	// Use v2 suffix to avoid conflicts with deleted secrets
	superTokensSecret, err := secretsmanager.NewSecret(ctx, fmt.Sprintf("%s-%s-supertokens-secret-v2", projectName, environment), &secretsmanager.SecretArgs{
		Name:                        pulumi.String(fmt.Sprintf("%s-%s-supertokens-secret-v2", projectName, environment)),
		Description:                 pulumi.String("SuperTokens configuration (v2)"),
		RecoveryWindowInDays:        pulumi.Int(0), // Allow immediate deletion
		ForceOverwriteReplicaSecret: pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-supertokens-secret-v2", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	superTokensSecretValue := input.SuperTokensAPIKey.ApplyT(func(apiKey string) (string, error) {
		secretData := map[string]interface{}{
			"api_key": apiKey,
		}
		jsonData, err := json.Marshal(secretData)
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
	}).(pulumi.StringOutput)

	_, err = secretsmanager.NewSecretVersion(ctx, fmt.Sprintf("%s-%s-supertokens-secret-version", projectName, environment), &secretsmanager.SecretVersionArgs{
		SecretId:     superTokensSecret.ID(),
		SecretString: superTokensSecretValue,
	})
	if err != nil {
		return nil, err
	}

	return &SecretsResources{
		DatabaseSecret:       dbSecret,
		DatabaseSecretArn:    dbSecret.Arn,
		SuperTokensSecret:    superTokensSecret,
		SuperTokensSecretArn: superTokensSecret.Arn,
	}, nil
}
