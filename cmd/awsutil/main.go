package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	CLI_NAME = "awsutil"
)

func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()
	errBind := v.BindPFlags(cmd.Flags())
	if errBind != nil {
		return v, fmt.Errorf("error binding flag set to viper: %w\n", errBind)
	}
	v.SetEnvPrefix(CLI_NAME) // Enforces all env vars to require "AWSLOGIN_", making them unique
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	return v, nil
}

func main() {
	rootCommand := &cobra.Command{
		Use:                   fmt.Sprintf("%s [flags]", CLI_NAME),
		DisableFlagsInUseLine: true,
		Short:                 "This is a set of utilities for managing AWS accounts",
	}

	bootstrapGovcloudCommand := &cobra.Command{
		Use:                   `bootstrap-govcloud`,
		DisableFlagsInUseLine: true,
		Short:                 "Bootstrap a new AWS govcloud account",
		SilenceErrors:         true,
		SilenceUsage:          true,
		RunE:                  bootstrapGovcloud,
	}
	initBootstrapGovcloudFlags(bootstrapGovcloudCommand.Flags())

	bootstrapUserCommand := &cobra.Command{
		Use:                   `bootstrap-user`,
		DisableFlagsInUseLine: true,
		Short:                 "Bootstrap an IAM user",
		SilenceErrors:         true,
		SilenceUsage:          true,
		RunE:                  bootstrapUser,
	}
	initBootstrapUserFlags(bootstrapUserCommand.Flags())

	resetUserPasswordCommand := &cobra.Command{
		Use:                   `reset-user-password`,
		DisableFlagsInUseLine: true,
		Short:                 "Reset the password for an IAM user",
		SilenceErrors:         true,
		SilenceUsage:          true,
		RunE:                  resetUserPassword,
	}
	initResetUserPasswordFlags(resetUserPasswordCommand.Flags())

	versionCommand := &cobra.Command{
		Use:                   `version`,
		DisableFlagsInUseLine: true,
		Short:                 "gitlab POC on events",
		SilenceErrors:         true,
		SilenceUsage:          true,
		RunE:                  printVersion,
	}

	rootCommand.AddCommand(
		bootstrapGovcloudCommand,
		bootstrapUserCommand,
		resetUserPasswordCommand,
		versionCommand,
	)

	if err := rootCommand.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", CLI_NAME, err.Error())
		_, _ = fmt.Fprintf(os.Stderr, "Try %s --help for more information.\n", CLI_NAME)
		os.Exit(1)
	}
}
