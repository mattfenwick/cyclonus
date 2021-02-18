package worker

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func Run() {
	command := SetupRootCommand()
	if err := errors.Wrapf(command.Execute(), "run root command"); err != nil {
		log.Fatalf("unable to run root command: %+v", err)
		os.Exit(1)
	}
}

type Args struct {
	//Verbosity string
	Jobs        string
	Concurrency int
}

func SetupRootCommand() *cobra.Command {
	args := &Args{}
	command := &cobra.Command{
		Use:   "cyclonus-worker",
		Short: "thin wrapper around 'agnhost connect' for issuing batches of connectivity requests",
		Run: func(cmd *cobra.Command, as []string) {
			RunWorkerCommand(args)
		},
	}

	//command.Flags().StringVarP(&args.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	command.Flags().IntVar(&args.Concurrency, "concurrency", 10, "number of jobs to simultaneously run")

	command.Flags().StringVar(&args.Jobs, "jobs", "", "JSON-formatted string of jobs")
	utils.DoOrDie(command.MarkFlagRequired("jobs"))

	return command
}

func RunWorkerCommand(args *Args) {
	//utils.DoOrDie(utils.SetUpLogger(args.Verbosity))

	out, err := RunWorker(args.Jobs, args.Concurrency)
	utils.DoOrDie(err)
	fmt.Printf("%s\n", out)
}
