package arangolite

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"time"
)

type logger struct {
	log.Logger
}

func newLogger(enabled bool) *logger {
	var out *os.File

	if enabled {
		out = os.Stdout
	}

	return &logger{*log.New(out, "\n[Arangolite] ", 0)}
}

func (l *logger) logBegin(msg, url string, jsonQuery []byte) {
	l.Printf("%s %s %s | URL: %s\n    %s", blue, msg, reset, url, indentJSON(jsonQuery))
}

func (l *logger) logResult(result *result, start time.Time, in, out chan interface{}) {
	batchNb := 0

	for {
		tmp := <-in
		out <- tmp

		switch tmp.(type) {
		case json.RawMessage:
			batchNb++
			continue
		}
		break
	}

	execTime := time.Now().Sub(start)
	content := string(indentJSON([]byte(result.Content)))
	if len(content) > 5000 {
		content = content[0:5000] + "\n\n    Result has been truncated to 5000 characters"
	}

	if result.Cached {
		l.Printf("%s RESULT %s | %s CACHED %s | Execution: %v | Batches: %d\n    %s",
			blue, reset, yellow, reset, execTime, batchNb, content)
	} else {
		l.Printf("%s RESULT %s | Execution: %v | Batches: %d\n    %s",
			blue, reset, execTime, batchNb, content)
	}
}

func (l *logger) logError(errMsg string, execTime time.Duration) {
	l.Printf("%s RESULT %s | Execution: %v\n    ERROR: %s",
		blue, reset, execTime, errMsg)
}

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

func indentJSON(in []byte) []byte {
	b := &bytes.Buffer{}
	_ = json.Indent(b, in, "    ", "  ")

	return b.Bytes()
}
