package worker

import (
	"encoding/json"

	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Client struct {
	Kubernetes kube.IKubernetes
}

func (c *Client) Batch(b *Batch) ([]*Result, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to marshal json")
	}
	command := []string{"/worker", "--jobs", string(bytes)}
	logrus.Infof("issuing %s worker command with %d requests", b.Key(), len(b.Requests))
	stdout, stderr, commandErr, err := c.Kubernetes.ExecuteRemoteCommand(b.Namespace, b.Pod, b.Container, command)
	logrus.Tracef("%s worker stdout:\n%s\nworker stderr:\n%s\n", b.Key(), stdout, stderr)

	if err != nil {
		return nil, err
	} else if commandErr != nil {
		return nil, commandErr
	}

	var results []*Result
	err = json.Unmarshal([]byte(stdout), &results)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal json")
	}

	if len(results) != len(b.Requests) {
		return results, errors.Errorf("expected %d results, but got only %d", len(b.Requests), len(results))
	}

	return results, nil
}

/*
type Client struct {
	Resty *resty.Client
}

func NewDefaultClient(host string) *Client {
	return NewClient(host, DefaultPort)
}

func NewClient(host string, port int) *Client {
	return &Client{Resty: resty.New().
		SetHostURL(net.JoinHostPort(host, strconv.Itoa(port))).
		//SetDebug(false).
		SetTimeout(30 * time.Second)}
}

func (c *Client) Batch(b *Batch) ([]*Result, error) {
	var results []*Result
	_, err := utils.IssueRequest(c.Resty, "POST", "/batch", b, &results)
	return results, err
}
*/
