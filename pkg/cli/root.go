package cli

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func RunRootCommand() {
	command := SetupRootCommand()
	if err := errors.Wrapf(command.Execute(), "run root command"); err != nil {
		log.Fatalf("unable to run root command: %+v", err)
		os.Exit(1)
	}
}

type Flags struct {
	Verbosity string
}

func SetupRootCommand() *cobra.Command {
	flags := &Flags{}
	command := &cobra.Command{
		Use:   "cyclonus",
		Short: "explain, probe, and query network policies",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return SetUpLogger(flags.Verbosity)
		},
	}

	command.PersistentFlags().StringVarP(&flags.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	command.AddCommand(SetupAnalyzePoliciesCommand())
	command.AddCommand(SetupQueryTrafficCommand())
	command.AddCommand(SetupSyntheticProbeConnectivityCommand())
	command.AddCommand(SetupQueryTargetsCommand())
	command.AddCommand(SetupGenerateCommand())
	command.AddCommand(SetupProbeCommand())

	// TODO
	//command.AddCommand(setupQueryPeersCommand())

	return command
}
