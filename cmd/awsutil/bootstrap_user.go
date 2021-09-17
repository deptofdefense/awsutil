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
	flagBootstrapUserUser = "user"
)

func initBootstrapUserFlags(flag *pflag.FlagSet) {
	flag.String(flagBootstrapUserUser, "", "The name of the IAM user to update")
}

func checkBootstrapUserConfig(v *viper.Viper) error {
	userName := v.GetString(flagBootstrapUserUser)
	if len(userName) == 0 {
		return errors.New("The IAM user name should not be empty")
	}
	return nil
}

func bootstrapUser(cmd *cobra.Command, args []string) error {

	v, errViper := initViper(cmd)
	if errViper != nil {
		return fmt.Errorf("error initializing viper: %w\n", errViper)
	}

	if errConfig := checkBootstrapUserConfig(v); errConfig != nil {
		return errConfig
	}

	userName := v.GetString(flagBootstrapUserUser)

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

	_, errGetUser := svcIAM.GetUser(context.TODO(), &iam.GetUserInput{
		UserName: &userName,
	})
	if errGetUser != nil {
		fmt.Printf("User %s has not yet been provisioned", userName)
		return errGetUser
	}

	_, errCreateLoginProfile := svcIAM.CreateLoginProfile(context.TODO(), &iam.CreateLoginProfileInput{
		UserName: &userName,
		Password: password,
	})
	if errCreateLoginProfile != nil {
		return errCreateLoginProfile
	}

	listAccountAliasesOutput, errListAccountAliases := svcIAM.ListAccountAliases(context.TODO(), &iam.ListAccountAliasesInput{})
	if errListAccountAliases != nil {
		return errListAccountAliases
	}

	var loginUrl, securityCredsUrl string
	if len(listAccountAliasesOutput.AccountAliases) > 0 {

		alias := listAccountAliasesOutput.AccountAliases[0]
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "us-gov-east-1" || awsRegion == "us-gov-west-1" {
			loginUrl = fmt.Sprintf("https://%s.signin.amazonaws-us-gov.com/console", alias)
			securityCredsUrl = fmt.Sprintf("https://console.amazonaws-us-gov.com/iam/home?region=%s#/users/%s?section=security_credentials", awsRegion, userName)
		} else {
			loginUrl = fmt.Sprintf("https://%s.signin.aws.amazon.com/console", alias)
			securityCredsUrl = fmt.Sprintf("https://console.aws.amazon.com/iam/home?#/users/%s?section=security_credentials", userName)
		}
	} else {
		loginUrl = "https://console.aws.amazon.com/"
	}

	fmt.Printf("Login URL: %s\n", loginUrl)
	fmt.Printf("Username: %s\n", userName)
	fmt.Printf("Password: %s\n", *password)
	fmt.Println(`Please follow these steps:
1. Log in to the console with your temporary password.
2. Create an MFA. Save MFA to 1Password as One Time Password (OTP)\n`)
	if len(securityCredsUrl) > 0 {
		fmt.Printf("\tURL: %s\n", securityCredsUrl)
	}
	fmt.Println(`3. Log out of the AWS Console
4. Log in to the console with your new MFA
5. Reset your password (min 20 chars, requires upper and lowercase, numbers and symbols)
6. Assume the IAM Role you wish to use.

NOTE: You will not be able to do anything in the account unless you log in with MFA and assume the AWS IAM role for your project.\n`)

	return nil
}
