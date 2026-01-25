// Package http contains a client to use for http requests.
package http

import (
	"github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	*retryablehttp.Client
}

func New() *Client {
	client := retryablehttp.NewClient()
	return &Client{
		Client: client,
	}
}
