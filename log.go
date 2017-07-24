package arangolite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

// LogVerbosity is the logging verbosity.
type LogVerbosity int

const (
	// LogSummary prints a simple summary of the exchanges with the database.
	LogSummary LogVerbosity = iota
	// LogDebug prints all the sent and received http requests.
	LogDebug
)

// newLoggingSender returns a logging wrapper around a sender.
func newLoggingSender(sender sender, logger *log.Logger, verbosity LogVerbosity) sender {
	return &loggingSender{
		sender:    sender,
		logger:    logger,
		verbosity: verbosity,
	}
}

type loggingSender struct {
	sender    sender
	logger    *log.Logger
	verbosity LogVerbosity
}

func (s *loggingSender) Send(ctx context.Context, cli *http.Client, req *http.Request) (*response, error) {
	dump := bytes.NewBuffer(nil)
	defer func() { s.logger.Print(dump.String()) }()

	dump.WriteString("\nRequest:")
	switch s.verbosity {
	case LogSummary:
		dump.WriteString(fmt.Sprintf(" %s %s \n", req.Method, req.URL.EscapedPath()))
	case LogDebug:
		dump.WriteString("\n")
		r, _ := httputil.DumpRequestOut(req, true)
		dump.Write(r)
		dump.WriteString("\n\n")
	}

	now := time.Now()
	res, err := s.sender.Send(ctx, cli, req)
	if err != nil {
		dump.WriteString("Send error: ")
		dump.WriteString(err.Error())
		dump.WriteString("\n")
		return nil, err
	}
	if res.parsed.Error {
		dump.WriteString("Database error: ")
		dump.WriteString(res.parsed.ErrorMessage)
		dump.WriteString("\n")
		return res, nil
	}

	dump.WriteString("Success in ")
	dump.WriteString(time.Since(now).String())
	if s.verbosity == LogDebug {
		// buf := bytes.NewBuffer(nil)
		// raw := res.raw
		dump.WriteString(":\n")
		if err := json.Indent(dump, res.raw, "", "\t"); err != nil {
			dump.Write(res.raw)
		}
	}
	dump.WriteString("\n")

	return res, nil
}
