// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package newsman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	// GET is a simple alias to ensure a certain continuity in HTTP verb usage
	GET method = http.MethodGet
	// HEAD is a simple alias to ensure a certain continuity in HTTP verb usage
	HEAD method = http.MethodHead
	// POST is a simple alias to ensure a certain continuity in HTTP verb usage
	POST method = http.MethodPost
	// PUT is a simple alias to ensure a certain continuity in HTTP verb usage
	PUT method = http.MethodPut
	// PATCH is a simple alias to ensure a certain continuity in HTTP verb usage
	PATCH method = http.MethodPatch
	// DELETE is a simple alias to ensure a certain continuity in HTTP verb usage
	DELETE method = http.MethodDelete
)

type (
	method string

	// WebhookConfig represents a config instruction
	WebhookConfig struct {
		Method      string
		ContentType string
		URL         string
	}

	// WebhookListener is a middleman between the websocket connection and the newsman.
	WebhookListener struct {
		config         *WebhookConfig
		listenerConfig *ListenerConfig
		httpClient     *http.Client
		errFunc        func(error)
		incoming       chan Event
	}
)

func nilErrFunc(_ error) {}

func buildDefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

// NewWebhookListener constructs a new config listener
func NewWebhookListener(errFunc func(error), config *WebhookConfig, listenerConfig *ListenerConfig) Listener {
	if errFunc == nil {
		errFunc = nilErrFunc
	}

	whl := &WebhookListener{
		listenerConfig: listenerConfig,
		config:         config,
		errFunc:        errFunc,
		httpClient:     buildDefaultHTTPClient(),
		incoming:       make(chan Event, 256),
	}

	return whl
}

// Listen sends messages from the newsman to the websocket connection.
//
// A goroutine running Listen is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (l *WebhookListener) Listen() {
	for {
		select {
		case event, ok := <-l.incoming:
			if !ok {
				return // The channel is closed.
			}

			if err := l.sendWebhook(event); err != nil {
				l.errFunc(err)
			}

			// Add queued chat messages to the current websocket message.
			n := len(l.incoming)
			for i := 0; i < n; i++ {
				x := <-l.incoming
				if err := l.sendWebhook(x); err != nil {
					l.errFunc(err)
				}
			}
		}
	}
}

// Channel implements our listener interface
func (l *WebhookListener) Channel() chan Event {
	return l.incoming
}

// Config implements our listener interface
func (l *WebhookListener) Config() *ListenerConfig {
	return l.listenerConfig
}

func (l *WebhookListener) bodyBuilder(in interface{}) io.Reader {
	switch strings.TrimSpace(strings.ToLower(l.config.ContentType)) {
	// case "application/json":
	default:
		out, _ := json.Marshal(in)
		return bytes.NewReader(out)
	}
}

func (l *WebhookListener) sendWebhook(event Event) error {
	req, err := http.NewRequest(l.config.Method, l.config.URL, l.bodyBuilder(event.Data))
	if err != nil {
		log.Printf("error encountered executing config: %v", err)
		return err
	}

	if l.config.ContentType != "" {
		req.Header.Set("Content-type", l.config.ContentType)
	}

	res, err := l.httpClient.Do(req)
	if err != nil {
		log.Printf("error encountered executing config: %v", err)
		return err
	}

	if res.StatusCode >= http.StatusBadRequest && l.errFunc != nil {
		l.errFunc(fmt.Errorf("status code: %d", res.StatusCode))
	}
	return nil
}
