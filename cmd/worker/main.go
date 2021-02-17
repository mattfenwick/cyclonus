package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	port = 23456
)

func main() {
	log.SetLevel(log.DebugLevel)

	stop := make(chan struct{})
	SetupHTTPServer()

	addr := fmt.Sprintf(":%d", port)
	go func() {
		log.Infof("starting HTTP server on port %d", port)
		utils.DoOrDie(http.ListenAndServe(addr, nil))
	}()

	<-stop
}

func SetupHTTPServer() {
	http.HandleFunc("/batch", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("received request: %s to %s", r.Method, r.URL.String())
		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				HandleError(w, r, errors.Wrapf(err, "unable to read request body"), 400)
				return
			}
			log.Tracef("received body: %s", body)
			var batch Batch
			err = json.Unmarshal(body, &batch)
			if err != nil {
				HandleError(w, r, errors.Wrapf(err, "unable to unmarshal JSON"), 400)
				return
			}

			results, err := batch.Issue()
			if err != nil {
				HandleError(w, r, err, 400)
				return
			}

			jsonBytes, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				HandleError(w, r, errors.Wrapf(err, "unable to marshal JSON"), 500)
				return
			}
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			_, err = fmt.Fprint(w, string(jsonBytes))
			if err != nil {
				HandleError(w, r, errors.Wrapf(err, "unable to write response"), 500)
			}
		} else {
			NotFound(w, r)
		}
	})
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	log.Errorf("HTTPResponder not found from request %+v", r)
	http.NotFound(w, r)
}

func HandleError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	log.Errorf("HTTPResponder error %s with code %d from request %+v", err.Error(), statusCode, r)
	http.Error(w, err.Error(), statusCode)
}

type Batch struct {
	Requests []*Request
}

func (b *Batch) Issue() ([]*Result, error) {
	var results []*Result
	for _, r := range b.Requests {
		result, err := r.Issue()
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

type Result struct {
	Request *Request
	Output  string
	Error   string
}

type Request struct {
	Protocol v1.Protocol
	Address  string
}

func (r *Request) Command() ([]string, error) {
	switch r.Protocol {
	case v1.ProtocolSCTP:
		return []string{"/agnhost", "connect", r.Address, "--timeout=1s", "--protocol=sctp"}, nil
	case v1.ProtocolTCP:
		return []string{"/agnhost", "connect", r.Address, "--timeout=1s", "--protocol=tcp"}, nil
	case v1.ProtocolUDP:
		return []string{"/agnhost", "connect", r.Address, "--timeout=1s", "--protocol=udp"}, nil
	default:
		return nil, errors.Errorf("protocol %s not supported", r.Protocol)
	}
}

func (r *Request) Issue() (*Result, error) {
	command, err := r.Command()
	if err != nil {
		return nil, err
	}
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
	}, nil
}

type Client struct {
	Resty *resty.Client
}

func NewClient(host string) *Client {
	return &Client{Resty: resty.New().
		SetHostURL(fmt.Sprintf("%s:%d", host, port)).
		//SetDebug(false).
		SetTimeout(30 * time.Second)}
}

func (c *Client) Batch(b *Batch) ([]*Result, error) {
	var results []*Result
	_, err := utils.IssueRequest(c.Resty, "POST", "/batch", b, &results)
	return results, err
}
