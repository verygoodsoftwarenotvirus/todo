package main

import (
	"math"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type (
	requestBuilder func() (*http.Request, error)

	builder struct {
		ReqBuilder  requestBuilder
		Probability float64
	}

	jobPool struct {
		RequestBuilders []builder
	}
)

func (p *jobPool) GetRandomRequestBuilder() requestBuilder {
	var (
		index = float64(rand.Intn(len(p.RequestBuilders)))
		sum   float64
		i     int
	)

	for sum < index {
		i++
		sum += p.RequestBuilders[i].Probability
	}

	return p.RequestBuilders[int(math.Max(0, float64(i-1)))].ReqBuilder
}
