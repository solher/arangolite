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
	printQuery  bool
	printResult bool
}

func newLogger() *logger {
	l := &logger{}
	l.Logger = *log.New(os.Stdout, "", 0)
	l.printQuery = true
	l.printResult = true

	return l
}

func (l *logger) Options(enabled, printQuery, printResult bool) *logger {
	var out *os.File

	if enabled {
		out = os.Stdout
	}

	l.Logger = *log.New(out, "", 0)

	l.printQuery = printQuery
	l.printResult = printResult

	return l
}

func (l *logger) LogBegin(msg, url string, jsonQuery []byte) {
	l.Printf("\n[Arangolite] %s %s %s | URL: %s", blue, msg, reset, url)

	if l.printQuery {
		l.Println("    " + string(indentJSON(jsonQuery)))
	}
}

func (l *logger) LogResult(result *result, start time.Time, in, out chan interface{}) {
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

	if result.Cached {
		l.Printf("\n[Arangolite] %s RESULT %s | %s CACHED %s | Execution: %v | Batches: %d",
			blue, reset, yellow, reset, execTime, batchNb)
	} else {
		l.Printf("\n[Arangolite] %s RESULT %s | Execution: %v | Batches: %d",
			blue, reset, execTime, batchNb)
	}

	if l.printResult {
		content := "    " + string(indentJSON([]byte(result.Content)))
		if batchNb > 1 {
			content += "\n\n    Result has been truncated to the first batch"
		}
		l.Println(content)
	}
}

func (l *logger) LogError(errMsg string, execTime time.Duration) {
	l.Printf("\n[Arangolite] %s RESULT %s | Execution: %v\n    ERROR: %s",
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
