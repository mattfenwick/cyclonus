package worker

import (
	"encoding/json"
	"os/exec"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

var (
	protocols = map[v1.Protocol]bool{
		v1.ProtocolUDP:  true,
		v1.ProtocolSCTP: true,
		v1.ProtocolTCP:  true,
	}
)

func RunWorker(jobs string, concurrency int) (string, error) {
	var batch Batch
	err := json.Unmarshal([]byte(jobs), &batch)
	if err != nil {
		return "", errors.Wrapf(err, "unable to unmarshal json from '%s'", jobs)
	}

	if err := batch.IsValid(); err != nil {
		return "", err
	}

	results := IssueBatch(&batch, concurrency)

	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", errors.Wrapf(err, "unable to marshal json")
	}
	return string(jsonBytes), nil
}

func IssueBatch(batch *Batch, concurrency int) []*Result {
	requestChan := make(chan *Request)
	resultChan := make(chan *Result, len(batch.Requests))
	for i := 0; i < concurrency; i++ {
		go worker(requestChan, resultChan)
	}
	for _, b := range batch.Requests {
		requestChan <- b
	}
	close(requestChan)

	var resultSlice []*Result
	for i := 0; i < len(batch.Requests); i++ {
		resultSlice = append(resultSlice, <-resultChan)
	}
	return resultSlice
}

func worker(requests <-chan *Request, results chan<- *Result) {
	for request := range requests {
		results <- IssueRequestWithRetries(request, 1)
	}
}

func IssueRequestWithRetries(r *Request, retries int) *Result {
	result := IssueRequest(r)
	for i := 0; i < retries && !result.IsSuccess(); i++ {
		result = IssueRequest(r)
	}
	return result
}

func IssueRequest(r *Request) *Result {
	command := r.Command()
	name, args := command[0], command[1:]
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	var errString string
	if err != nil {
		errString = errors.Wrapf(err, "unable to run command '%s'", cmd.String()).Error()
	}
	return &Result{
		Request: r,
		Output:  string(out),
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
