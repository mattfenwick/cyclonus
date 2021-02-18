package worker

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	Kubernetes *kube.Kubernetes
}

func (c *Client) Batch(b *Batch) ([]*Result, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to marshal json")
	}
	command := []string{"/worker", "--jobs", fmt.Sprintf("'%s'", bytes)}
	stdout, stderr, commandErr, err := c.Kubernetes.ExecuteRemoteCommand(b.Namespace, b.Pod, b.Container, command)
	log.Debugf("%+v worker stdout:\n%s\nworker stderr:\n%s\n", b, stdout, stderr)

	if err != nil {
		return nil, err
	} else if commandErr != nil {
		return nil, commandErr
	}

	var results []*Result
	err = json.Unmarshal([]byte(stdout), &results)
	return results, errors.Wrapf(err, "unable to unmarshal json")
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
		SetHostURL(fmt.Sprintf("%s:%d", host, port)).
		//SetDebug(false).
		SetTimeout(30 * time.Second)}
}

func (c *Client) Batch(b *Batch) ([]*Result, error) {
	var results []*Result
	_, err := utils.IssueRequest(c.Resty, "POST", "/batch", b, &results)
	return results, err
}
*/
