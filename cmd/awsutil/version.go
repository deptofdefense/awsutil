package main

import (
	"fmt"

	"github.com/deptofdefense/awsutil/pkg/version"

	"github.com/spf13/cobra"
)

func printVersion(cmd *cobra.Command, args []string) error {
	fmt.Println(version.Full())
	return nil
}
