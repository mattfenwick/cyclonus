package utils

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

func CommandRun(cmd *exec.Cmd) (string, error) {
	if len(cmd.Dir) == 0 {
		dir, _ := os.Getwd()
		log.Debugf("running command: '%s' in current directory '%s'", cmd.String(), dir)
	} else {
		log.Debugf("running command: '%s' in directory '%s'", cmd.String(), cmd.Dir)
	}
	cmdOutput, err := cmd.CombinedOutput()
	cmdOutputStr := string(cmdOutput)
	log.Tracef("command: '%s' output:\n%s", cmd.String(), cmdOutput)
	return cmdOutputStr, errors.Wrapf(err, "unable to run command '%s': %s", cmd.String(), cmdOutputStr)
}

func CommandExtendEnvironment(cmd *exec.Cmd, env map[string]string) {
	cmd.Env = os.Environ()
	for key, val := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}
}

func CommandRunAndCaptureProgress(cmd *exec.Cmd) error {
	_, _ = cmd.StdinPipe()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// can't use CommandRun(cmd) here -- attaching to os pipes interferes with cmd.CombinedOutput()
	log.Infof("running command '%s' with pipes attached in directory '%s' and with env \n%+v\n", cmd.String(), cmd.Dir, cmd.Env)
	return errors.Wrapf(cmd.Run(), "unable to run command '%s'", cmd.String())
}
