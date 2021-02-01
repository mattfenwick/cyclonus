package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	version   = "development"
	gitSHA    = "development"
	buildTime = "development"
)

func SetupVersionCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "version",
		Short: "print out version information",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunVersionCommand()
		},
	}

	return command
}

func RunVersionCommand() {
	fmt.Printf("Version: %s\nGit SHA: %s\nBuild time: %s\n", version, gitSHA, buildTime)
}
