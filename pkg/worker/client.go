package worker

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"time"
)

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
