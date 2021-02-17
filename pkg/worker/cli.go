package worker

import (
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

const (
	DefaultPort = 23456
)

func Run() {
	command := SetupRootCommand()
	if err := errors.Wrapf(command.Execute(), "run root command"); err != nil {
		log.Fatalf("unable to run root command: %+v", err)
		os.Exit(1)
	}
}

type Args struct {
	Verbosity string
	Port      int
}

func SetupRootCommand() *cobra.Command {
	args := &Args{}
	command := &cobra.Command{
		Use:   "cyclonus-worker",
		Short: "thin wrapper around agnhost for issuing batches of connectivity requests",
		Run: func(cmd *cobra.Command, as []string) {
			RunWorker(args)
		},
	}

	command.Flags().StringVarP(&args.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")
	command.Flags().IntVar(&args.Port, "port", DefaultPort, "port to run server on")

	return command
}

func RunWorker(args *Args) {
	utils.DoOrDie(utils.SetUpLogger(args.Verbosity))

	RunServer(args.Port)
}
