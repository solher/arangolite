package arangolite

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"strings"

	"github.com/pkg/errors"
)

type sender interface {
	Send(ctx context.Context, cli *http.Client, req *http.Request) (*response, error)
}

type basicSender struct{}

func (s *basicSender) Send(ctx context.Context, cli *http.Client, req *http.Request) (*response, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		break
	}
	res, err := cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "the database HTTP request failed")
	}

	raw, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not read the database response")
	}

	parsed := parsedResponse{}
	if strings.Contains(res.Header.Get("Content-Type"), "application/json") && raw[0] != '[' {
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return nil, errors.Wrap(err, "could not decode the json database response")
		}
	}

	raw = []byte(strings.TrimSpace(string(raw)))
	return &response{statusCode: res.StatusCode, raw: raw, parsed: parsed}, nil
}

type parsedResponse struct {
	Error        bool            `json:"error"`
	ErrorMessage string          `json:"errorMessage"`
	ErrorNum     int             `json:"errorNum"`
	Result       json.RawMessage `json:"result"`
	HasMore      bool            `json:"hasMore"`
	ID           string          `json:"id"`
}

type response struct {
	statusCode int
	raw        json.RawMessage
	parsed     parsedResponse
}

func (r *response) Raw() json.RawMessage {
	return r.raw
}

func (r *response) RawResult() json.RawMessage {
	return r.parsed.Result
}

func (r *response) StatusCode() int {
	return r.statusCode
}

func (r *response) HasMore() bool {
	return r.parsed.HasMore
}

func (r *response) Cursor() string {
	return r.parsed.ID
}

func (r *response) Unmarshal(v interface{}) error {
	if err := json.Unmarshal(r.raw, v); err != nil {
		return errors.Wrap(err, "response unmarshalling failed")
	}
	return nil
}

func (r *response) UnmarshalResult(v interface{}) error {
	if err := json.Unmarshal(r.parsed.Result, v); err != nil {
		return errors.Wrap(err, "response result unmarshalling failed")
	}
	return nil
}
