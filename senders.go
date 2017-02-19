package arangolite

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type sender interface {
	Send(cli *http.Client, req *http.Request) (Result, error)
}

type basicSender struct{}

func (s *basicSender) Send(cli *http.Client, req *http.Request) (Result, error) {
	res, err := cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "the database HTTP request failed")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		res.Body.Close()
		return nil, errors.Errorf("the database HTTP request failed, status code %d", res.StatusCode)
	}

	raw, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not read the database response")
	}
	parsed := parsedResult{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, errors.Wrap(err, "database response decoding failed")
	}

	if parsed.Error {
		err := errors.Wrap(errors.New(parsed.ErrorMessage), "the database returned an error")
		switch {
		case strings.Contains(parsed.ErrorMessage, "unique constraint violated"):
			err = withErrUnique(err)
		case strings.Contains(parsed.ErrorMessage, "not found") || strings.Contains(parsed.ErrorMessage, "unknown collection"):
			err = withErrNotFound(err)
		case strings.Contains(parsed.ErrorMessage, "duplicate name"):
			err = withErrDuplicate(err)
		}
		return nil, err
	}

	return &result{raw: raw, parsed: parsed}, nil
}

type parsedResult struct {
	Error        bool            `json:"error"`
	ErrorMessage string          `json:"errorMessage"`
	Content      json.RawMessage `json:"result"`
	Cached       bool            `json:"cached"`
	HasMore      bool            `json:"hasMore"`
	ID           string          `json:"id"`
}

type result struct {
	raw    json.RawMessage
	parsed parsedResult
}

func (r *result) Raw() json.RawMessage {
	return r.raw
}

func (r *result) RawContent() json.RawMessage {
	return r.parsed.Content
}

func (r *result) HasMore() bool {
	return r.parsed.HasMore
}

func (r *result) Cursor() string {
	return r.parsed.ID
}

func (r *result) Unmarshal(v interface{}) error {
	return json.Unmarshal(r.raw, v)
}
