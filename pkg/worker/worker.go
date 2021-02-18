package worker

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"os/exec"
)

var (
	protocols = map[v1.Protocol]bool{
		v1.ProtocolUDP:  true,
		v1.ProtocolSCTP: true,
		v1.ProtocolTCP:  true,
	}
)

func RunWorker(jobs string) (string, error) {
	log.Tracef("received jobs: %s", jobs)
	var batch Batch
	err := json.Unmarshal([]byte(jobs), &batch)
	if err != nil {
		return "", errors.Wrapf(err, "unable to unmarshal json")
	}

	if err := batch.IsValid(); err != nil {
		return "", err
	}

	results, err := batch.Issue()
	if err != nil {
		return "", errors.Wrapf(err, "unable to issue requests")
	}

	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", errors.Wrapf(err, "unable to marshal json")
	}
	return string(jsonBytes), nil
}

type Batch struct {
	Namespace string
	Pod       string
	Container string
	Requests  []*Request
}

func (b *Batch) IsValid() error {
	for _, r := range b.Requests {
		if !protocols[r.Protocol] {
			return errors.Errorf("invalid protocol %s", r.Protocol)
		}
	}
	return nil
}

func (b *Batch) Issue() ([]*Result, error) {
	var results []*Result
	for _, r := range b.Requests {
		results = append(results, r.Issue())
	}
	return results, nil
}

type Result struct {
	Request *Request
	Output  string
	Error   string
}

func (r *Result) IsSuccess() bool {
	return r.Error == ""
}

type Request struct {
	Key      string
	Protocol v1.Protocol
	Host     string
	Port     int
}

func (r *Request) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func (r *Request) Command() []string {
	switch r.Protocol {
	case v1.ProtocolSCTP:
		return []string{"/agnhost", "connect", r.Address(), "--timeout=1s", "--protocol=sctp"}
	case v1.ProtocolTCP:
		return []string{"/agnhost", "connect", r.Address(), "--timeout=1s", "--protocol=tcp"}
	case v1.ProtocolUDP:
		return []string{"/agnhost", "connect", r.Address(), "--timeout=1s", "--protocol=udp"}
	default:
		panic(errors.Errorf("protocol %s not supported", r.Protocol))
	}
}

func (r *Request) Issue() *Result {
	command := r.Command()
	name, args := command[0], command[1:]
	out, err := utils.CommandRun(exec.Command(name, args...))
	var errString string
	if err != nil {
		errString = err.Error()
	}
	return &Result{
		Request: r,
		Output:  out,
		Error:   errString,
	}
}

// Can't run as a server over http for two reasons:
// 1. the container program needs to be /agnhost serving on a port/protocol
// 2. the invocation needs to be a kubectl exec to avoid getting blocked as collateral damage by a network policy
//func SetupHTTPServer() {
//	http.HandleFunc("/batch", func(w http.ResponseWriter, r *http.Request) {
//		log.Debugf("received request: %s to %s", r.Method, r.URL.String())
//		if r.Method == "POST" {
//			body, err := ioutil.ReadAll(r.Body)
//			if err != nil {
//				HandleError(w, r, errors.Wrapf(err, "unable to read request body"), 400)
//				return
//			}
//			log.Tracef("received body: %s", body)
//			var batch Batch
//			err = json.Unmarshal(body, &batch)
//			if err != nil {
//				HandleError(w, r, errors.Wrapf(err, "unable to unmarshal JSON"), 400)
//				return
//			}
//
//			results, err := batch.Issue()
//			if err != nil {
//				HandleError(w, r, err, 400)
//				return
//			}
//
//			jsonBytes, err := json.MarshalIndent(results, "", "  ")
//			if err != nil {
//				HandleError(w, r, errors.Wrapf(err, "unable to marshal JSON"), 500)
//				return
//			}
//			header := w.Header()
//			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
//			_, err = fmt.Fprint(w, string(jsonBytes))
//			if err != nil {
//				HandleError(w, r, errors.Wrapf(err, "unable to write response"), 500)
//			}
//		} else {
//			NotFound(w, r)
//		}
//	})
//}
//
//func NotFound(w http.ResponseWriter, r *http.Request) {
//	log.Errorf("HTTPResponder not found from request %+v", r)
//	http.NotFound(w, r)
//}
//
//func HandleError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
//	log.Errorf("HTTPResponder error %s with code %d from request %+v", err.Error(), statusCode, r)
//	http.Error(w, err.Error(), statusCode)
//}
