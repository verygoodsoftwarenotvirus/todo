package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io/ioutil"
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
	if err != nil {
		logger.Error(err, "dumping response")
	}
	logger = logger.WithValue("response", string(bs))
	logger.Info("response received")

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	var base64Encoded, gobEncoded []byte

	base64Encoded, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading response")
	}

	logger = logger.WithValue("base_64_enc_length", len(base64Encoded))
	logger.Debug("blah 1")

	gobEncoded, err = base64.URLEncoding.DecodeString(string(base64Encoded))
	if err != nil {
		return nil, errors.Wrap(err, "decoding base64 response")
	}

	logger = logger.WithValue("gob_enc_length", len(gobEncoded))
	logger.Debug("blah 2")

	err = gob.NewDecoder(bytes.NewReader(gobEncoded)).Decode(&data)
	if err != nil {
		return nil, errors.Wrap(err, "decoding gob")
	}

	logger.Debug("blah 3")

	return data, nil
}
