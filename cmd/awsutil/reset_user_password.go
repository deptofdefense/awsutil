package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	flagResetUserPasswordUser = "user"
)

func initResetUserPasswordFlags(flag *pflag.FlagSet) {
	flag.String(flagResetUserPasswordUser, "", "The name of the IAM user to update")
}

func checkResetUserPasswordConfig(v *viper.Viper) error {
	userName := v.GetString(flagResetUserPasswordUser)
	if len(userName) == 0 {
		return errors.New("The IAM user name should not be empty")
	}
	return nil
}

func resetUserPassword(cmd *cobra.Command, args []string) error {

	v, errViper := initViper(cmd)
	if errViper != nil {
		return fmt.Errorf("error initializing viper: %w\n", errViper)
	}

	if errConfig := checkResetUserPasswordConfig(v); errConfig != nil {
		return errConfig
	}

	userName := v.GetString(flagResetUserPasswordUser)

	awsCfg, errCfg := config.LoadDefaultConfig(context.TODO())
	if errCfg != nil {
		return errCfg
	}

	svcIAM := iam.NewFromConfig(awsCfg)
	svcSecretsManager := secretsmanager.NewFromConfig(awsCfg)

	getRandomPasswordOutput, errGetRandomPassword := svcSecretsManager.GetRandomPassword(context.TODO(), &secretsmanager.GetRandomPasswordInput{
		PasswordLength:          24,
		RequireEachIncludedType: true,
	})
	if errGetRandomPassword != nil {
		return errGetRandomPassword
	}
	password := getRandomPasswordOutput.RandomPassword

	_, errUpdateLoginProfile := svcIAM.UpdateLoginProfile(context.TODO(), &iam.UpdateLoginProfileInput{
		UserName: &userName,
		Password: password,
	})
	if errUpdateLoginProfile != nil {
		return errUpdateLoginProfile
	}

	listAccountAliasesOutput, errListAccountAliases := svcIAM.ListAccountAliases(context.TODO(), &iam.ListAccountAliasesInput{})
	if errListAccountAliases != nil {
		return errListAccountAliases
	}

	var loginUrl string
	if len(listAccountAliasesOutput.AccountAliases) > 0 {

		alias := listAccountAliasesOutput.AccountAliases[0]
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "us-gov-east-1" || awsRegion == "us-gov-west-1" {
			loginUrl = fmt.Sprintf("https://%s.signin.amazonaws-us-gov.com/console", alias)
		} else {
			loginUrl = fmt.Sprintf("https://%s.signin.aws.amazon.com/console", alias)
		}
	} else {
		loginUrl = "https://console.aws.amazon.com/"
	}

	fmt.Printf("Login URL: %s\n", loginUrl)
	fmt.Printf("Username: %s\n", userName)
	fmt.Printf("Password: %s\n", *password)
	fmt.Println(`Please follow these steps:
1. Log in to the console with your new password"
2. Reset your password\n`)

	return nil
}
