package cmd

import "github.com/spf13/cobra"

// markFlagRequired marks the provided flag as required panicking if it was not found.
func markFlagRequired(command *cobra.Command, flag string) {
	err := command.MarkFlagRequired(flag)
	if err != nil {
		panic(err)
	}
}
