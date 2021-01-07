package kube

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/util/exec"
	"regexp"
	"strconv"
)

type ProbeCommandType string

const (
	ProbeCommandTypeCurl   ProbeCommandType = "ProbeCommandTypeCurl"
	ProbeCommandTypeWget   ProbeCommandType = "ProbeCommandTypeWget"
	ProbeCommandTypeNetcat ProbeCommandType = "ProbeCommandTypeNetcat"
)

type ProbeResult struct {
	Out      string
	ErrorOut string
	Err      string
	ExitCode int
}

type ProbeCommand interface {
	Command() []string
	ParseOutput(out string, errorOut string, execErr error) *ProbeResult
}

type CurlCommand struct {
	TimeoutSeconds int
	URL            string
}

func (cc *CurlCommand) Command() []string {
	return []string{"curl", "-I", "--connect-timeout", fmt.Sprintf("%d", cc.TimeoutSeconds), cc.URL}
}

func (cc *CurlCommand) ParseOutput(out string, errorOut string, execErr error) *ProbeResult {
	if execErr == nil {
		return &ProbeResult{
			Out:      out,
			ErrorOut: errorOut,
			Err:      "",
			ExitCode: 0,
		}
	}

	switch e := execErr.(type) {
	case exec.CodeExitError:
		return &ProbeResult{
			Out:      out,
			ErrorOut: errorOut,
			Err:      e.Err.Error(),
			ExitCode: e.Code,
		}
	default:
		// Goal: distinguish between "we weren't able to run the command, for whatever reason" and
		//   "we were able to run the command just fine, but the command itself failed"
		// TODO Not sure if this is accomplishing that goal correctly
		if out != "" || errorOut != "" {
			log.Warningf("ignoring error for command '%s' : %+v", cc.Command(), execErr)

			curlRegexp := regexp.MustCompile(`curl: \((\d+)\)`)

			matches := curlRegexp.FindStringSubmatch(out)
			if len(matches) == 0 {
				panic(errors.Errorf("expected a match for %s", out))
			}
			curlExitCode, err := strconv.Atoi(matches[1])
			if err != nil {
				log.Fatalf("unable to parse '%s' to int: %+v", matches[1], err)
			}

			return &ProbeResult{
				Out:      out,
				ErrorOut: errorOut,
				Err:      "",
				ExitCode: curlExitCode,
			}
		}
		return &ProbeResult{
			Out:      out,
			ErrorOut: errorOut,
			Err:      execErr.Error(),
			ExitCode: -1,
		}
	}
}

type NetcatCommand struct {
	TimeoutSeconds int
	ToAddress      string
	ToPort         int
}

func (nc *NetcatCommand) Command() []string {
	return []string{"nc", "-v", "-z", "-w", fmt.Sprintf("%d", nc.TimeoutSeconds), nc.ToAddress, fmt.Sprintf("%d", nc.ToPort)}
}

func (nc *NetcatCommand) ParseOutput(out string, errorOut string, execErr error) *ProbeResult {
	if execErr == nil {
		return &ProbeResult{
			Out:      out,
			ErrorOut: errorOut,
			Err:      "",
			ExitCode: 0,
		}
	}

	switch e := execErr.(type) {
	case exec.CodeExitError:
		return &ProbeResult{
			Out:      out,
			ErrorOut: errorOut,
			Err:      e.Err.Error(),
			ExitCode: e.Code,
		}
	default:
		return &ProbeResult{
			Out:      out,
			ErrorOut: errorOut,
			Err:      e.Error(),
			ExitCode: -1,
		}
		//ncRegexp := regexp.MustCompile(`command terminated with exit code (\d+)`)
		//matches := ncRegexp.FindStringSubmatch(execErr.Error())
		//if len(matches) == 0 {
		//	panic(errors.Errorf("expected a match for '%s'", execErr.Error()))
		//}
		//exitCode, err := strconv.Atoi(matches[1])
		//if err != nil {
		//	log.Fatalf("unable to parse '%s' to int: %+v", matches[1], err)
		//}
		//
		//return &ProbeResult{
		//	Out:      out,
		//	ErrorOut: errorOut,
		//	Err:      execErr,
		//	ExitCode: exitCode,
		//}
	}
}

type WgetCommand struct {
	TimeoutSeconds int
	URL            string
}

func (wc *WgetCommand) Command() []string {
	// note some versions of wget want -s for spider mode, others, -S
	return []string{"wget", "--spider", "--tries", "1", "--timeout", fmt.Sprintf("%d", wc.TimeoutSeconds), wc.URL}
}

func (wc *WgetCommand) ParseOutput(out string, errorOut string, execErr error) *ProbeResult {
	panic("TODO")
}
