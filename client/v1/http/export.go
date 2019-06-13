package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/http/httputil"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

// BuildExportDataRequest builds an http Request for exporting a user's data
func (c *V1Client) BuildExportDataRequest(ctx context.Context) (*http.Request, error) {
	uri := c.BuildURL(nil, "data", "export")

	return http.NewRequest(http.MethodGet, uri, nil)
}

// ExportData retrieves a data export and loads it as an object
func (c *V1Client) ExportData(ctx context.Context) (*models.DataExport, error) {
	var data *models.DataExport
	logger := c.logger.WithValue("function_name", "ExportData")

	req, err := c.BuildExportDataRequest(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	res, err := c.executeRawRequest(ctx, c.authedClient, req)
	if err != nil {
		return nil, errors.Wrap(err, "executing request")
	}

	bs, err := httputil.DumpResponse(res, true)
	logger = logger.WithValue("response", string(bs))
	logger.Info("response received")

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	var base64Encoded, gobEncoded []byte

	l, err := base64.NewDecoder(base64.URLEncoding, res.Body).Read(base64Encoded)
	if err != nil {
		return nil, errors.Wrap(err, "reading response")
	}
	logger = logger.
		WithValue("l", l).
		WithValue("base_32_enc_length", len(base64Encoded)).
		WithValue("gob_enc_length", len(gobEncoded))
	logger.Debug("blah 1")

	if _, encErr := base64.URLEncoding.Decode(gobEncoded, base64Encoded); encErr != nil {
		return nil, errors.Wrap(encErr, "damnit")
	}

	logger = logger.
		WithValue("l", l).
		WithValue("base_32_enc_length", len(base64Encoded)).
		WithValue("gob_enc_length", len(gobEncoded))
	logger.Debug("blah 2")

	err = gob.NewDecoder(bytes.NewReader(base64Encoded)).Decode(&data)
	if err != nil {
		return nil, errors.Wrap(err, "decoding gob")
	}

	logger = logger.
		WithValue("l", l).
		WithValue("base_32_enc_length", len(base64Encoded)).
		WithValue("gob_enc_length", len(gobEncoded))
	logger.Debug("blah 3")

	return data, nil
}
