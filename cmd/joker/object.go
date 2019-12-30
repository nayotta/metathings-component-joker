package main

import "github.com/spf13/cobra"

var (
	ObjectCmd = &cobra.Command{
		Use:   "object",
		Short: "Simple Storage toolkits",
	}
)

func init() {
	RootCmd.AddCommand(ObjectCmd)
}
