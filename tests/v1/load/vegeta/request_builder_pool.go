package main

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/http_client/v1"

	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type (
	requestBuilder func() (*http.Request, error)

	builder struct {
		name       string
		ReqBuilder requestBuilder
		Weight     int
	}

	jobPool struct {
		totalWeight     int
		RequestBuilders map[string]builder
	}
)

func buildJobPoolFromClient(client *client.V1Client) *jobPool {
	jp := buildJobPool(
		builder{
			name: "BuildGetItemsRequest",
			ReqBuilder: func() (*http.Request, error) {
				return client.BuildGetItemsRequest(context.Background(), nil)
			},
			Weight: 1,
		},
		builder{
			name: "BuildGetOAuth2ClientsRequest",
			ReqBuilder: func() (*http.Request, error) {
				return client.BuildGetOAuth2ClientsRequest(context.Background(), nil)
			},
			Weight: 1,
		},
		builder{
			name: "BuildGetUsersRequest",
			ReqBuilder: func() (*http.Request, error) {
				return client.BuildGetUsersRequest(context.Background(), nil)
			},
			Weight: 3,
		},
	)
	return jp
}

func buildJobPool(inputs ...builder) *jobPool {
	jp := &jobPool{
		RequestBuilders: make(map[string]builder),
	}

	if inputs != nil {
		for _, b := range inputs {
			jp.RequestBuilders[b.name] = b
			jp.totalWeight += b.Weight
		}
	}

	return jp
}

func (p *jobPool) randomWeightedSelect() (requestBuilder, error) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(p.totalWeight)

	for _, rb := range p.RequestBuilders {
		r -= rb.Weight
		if r <= 0 {
			return rb.ReqBuilder, nil
		}
	}

	return nil, errors.New("No task selected")
}

func (p *jobPool) Read(retVal []byte) (int, error) {
	rb, err := p.randomWeightedSelect()
	if err != nil {
		return -1, errors.Wrap(err, "selecting task")
	}

	req, err := rb()
	if err != nil || req == nil {
		return -1, errors.Wrap(err, "building request")
	}

	target, err := reqToTarget(req)
	if err != nil {
		return -1, errors.Wrap(err, "converting request")
	}

	b, _ := json.Marshal(target)

	return bytes.NewReader(append(b, []byte("\n")[0])).Read(retVal)
}
