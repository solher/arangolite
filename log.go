package arangolite

import (
	"context"
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
	if s.verbosity == LogDebug {
		r, _ := httputil.DumpRequestOut(req, true)
		s.logger.Println("Request:")
		s.logger.Println(string(r))
	}

	now := time.Now()

	res, err := s.sender.Send(ctx, cli, req)
	if err != nil {
		s.logger.Printf("Send error: %s\n", err.Error())
		return nil, err
	}
	if res.parsed.Error {
		s.logger.Printf("Database error: %s\n", res.parsed.ErrorMessage)
		return res, nil
	}

	s.logger.Printf("Success in %v:\n", time.Since(now))
	if s.verbosity == LogDebug {
		s.logger.Println(res.raw)
	}

	return res, nil
}
