package internal

import (
	"net/http"
	"time"
)

type Requester struct {
	HttpClient *http.Client
}

func NewRequester() *Requester {
	return &Requester{
		HttpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}
