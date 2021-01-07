package utils

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

func CommandRun(cmd *exec.Cmd) (string, error) {
	if len(cmd.Dir) == 0 {
		dir, _ := os.Getwd()
		log.Infof("running command: '%s' in current directory '%s'", cmd.String(), dir)
	} else {
		log.Infof("running command: '%s' in directory '%s'", cmd.String(), cmd.Dir)
	}
	cmdOutput, err := cmd.CombinedOutput()
	cmdOutputStr := string(cmdOutput)
	log.Tracef("command: '%s' output:\n%s", cmd.String(), cmdOutput)
	return cmdOutputStr, errors.Wrapf(err, "unable to run command '%s': %s", cmd.String(), cmdOutputStr)
}

func CommandRunAndPrint(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// can't use CommandRun(cmd) here -- attaching to os pipes interferes with cmd.CombinedOutput()
	log.Infof("running command '%s' with pipes attached in directory: '%s'", cmd.String(), cmd.Dir)
	return errors.Wrapf(cmd.Run(), "unable to run command '%s'", cmd.String())
}
