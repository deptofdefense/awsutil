package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/99designs/aws-vault/v6/prompt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	flagBootstrapGovcloudProfile    = "profile"
	flagBootstrapGovcloudUser       = "user"
	flagBootstrapGovcloudGroupName  = "group-name"
	flagBootstrapGovcloudPolicyName = "policy-name"
)

func initBootstrapGovcloudFlags(flag *pflag.FlagSet) {
	flag.String(flagBootstrapGovcloudProfile, "", "The AWS Profile name to use from ~/.aws/config")
	flag.String(flagBootstrapGovcloudUser, "", "The name of the first IAM user to create")
	flag.String(flagBootstrapGovcloudGroupName, "admin", "The name of the IAM group used for administration")
	flag.String(flagBootstrapGovcloudPolicyName, "admin", "The name of the IAM policy used for administration")
}

func checkBootstrapGovcloudConfig(v *viper.Viper) error {
	awsProfile := v.GetString(flagBootstrapGovcloudProfile)
	if len(awsProfile) == 0 {
		return errors.New("The AWS Profile should not be empty")
	}
	userName := v.GetString(flagBootstrapGovcloudUser)
	if len(userName) == 0 {
		return errors.New("The IAM user name should not be empty")
	}
	groupName := v.GetString(flagBootstrapGovcloudGroupName)
	if len(groupName) == 0 {
		return errors.New("The IAM group name should not be empty")
	}
	policyName := v.GetString(flagBootstrapGovcloudPolicyName)
	if len(policyName) == 0 {
		return errors.New("The IAM policy name should not be empty")
	}
	return nil
}

type PolicyDocument struct {
	Version   string
	Statement []StatementEntry
}

type StatementEntry struct {
	Effect   string
	Action   []string
	Resource string
}

func bootstrapGovcloud(cmd *cobra.Command, args []string) error {

	v, errViper := initViper(cmd)
	if errViper != nil {
		return fmt.Errorf("error initializing viper: %w\n", errViper)
	}

	if errConfig := checkBootstrapGovcloudConfig(v); errConfig != nil {
		return errConfig
	}

	awsProfile := v.GetString(flagBootstrapGovcloudProfile)
	userName := v.GetString(flagBootstrapGovcloudUser)
	groupName := v.GetString(flagBootstrapGovcloudGroupName)
	policyName := v.GetString(flagBootstrapGovcloudPolicyName)

	awsCfg, errCfg := config.LoadDefaultConfig(context.TODO(),
		// Specify the shared configuration profile to load.
		config.WithSharedConfigProfile(awsProfile),
	)
	if errCfg != nil {
		return errCfg
	}

	svcIAM := iam.NewFromConfig(awsCfg)

	fmt.Printf("Create Group %s\n", groupName)
	_, errCreateGroup := svcIAM.CreateGroup(context.TODO(), &iam.CreateGroupInput{
		GroupName: &groupName,
		Path:      aws.String("/"),
	})
	if errCreateGroup != nil {
		return errCreateGroup
	}

	fmt.Printf("Create Policy %s in Group %s\n", policyName, groupName)
	policy := PolicyDocument{
		Version: "2012-10-17",
		Statement: []StatementEntry{
			{
				Effect:   "Allow",
				Action:   []string{"*"},
				Resource: "*",
			},
		},
	}
	policyBytes, errMarshal := json.Marshal(&policy)
	if errMarshal != nil {
		return errMarshal
	}
	policyDocument := string(policyBytes)
	_, errPutGroupPolicy := svcIAM.PutGroupPolicy(context.TODO(), &iam.PutGroupPolicyInput{
		GroupName:      &groupName,
		PolicyName:     &policyName,
		PolicyDocument: &policyDocument,
	})
	if errPutGroupPolicy != nil {
		return errPutGroupPolicy
	}

	fmt.Printf("Create User %s\n", userName)
	_, errCreateUser := svcIAM.CreateUser(context.TODO(), &iam.CreateUserInput{
		UserName: &userName,
		Path:     aws.String("/"),
	})
	if errCreateUser != nil {
		return errCreateUser
	}

	fmt.Printf("Add User %s to Group %s\n", userName, groupName)
	_, errAddUserToGroup := svcIAM.AddUserToGroup(context.TODO(), &iam.AddUserToGroupInput{
		GroupName: &groupName,
		UserName:  &userName,
	})
	if errAddUserToGroup != nil {
		return errAddUserToGroup
	}

	fmt.Println("List Users")
	listUsersOutput, errListUsers := svcIAM.ListUsers(context.TODO(), &iam.ListUsersInput{
		PathPrefix: aws.String("/"),
	})
	if errListUsers != nil {
		return errListUsers
	}
	for _, user := range listUsersOutput.Users {
		fmt.Println(user.UserName, user.Arn)
	}

	pass, errTerminalSecretPrompt := prompt.TerminalSecretPrompt(fmt.Sprintf("\nEnter a new password for %s: ", userName))
	if errTerminalSecretPrompt != nil {
		return errTerminalSecretPrompt
	}

	fmt.Printf("\nSet password for user %s\n", userName)
	_, errCreateLoginProfile := svcIAM.CreateLoginProfile(context.TODO(), &iam.CreateLoginProfileInput{
		UserName: &userName,
		Password: &pass,
	})
	if errCreateLoginProfile != nil {
		return errCreateLoginProfile
	}
	fmt.Println("COMPLETE!")

	return nil
}
