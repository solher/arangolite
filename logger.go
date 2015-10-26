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
	enabled, printQuery, printResult bool
}

func newLogger() *logger {
	l := &logger{}
	l.Logger = *log.New(os.Stdout, "", 0)
	l.enabled = true
	l.printQuery = true
	l.printResult = true

	return l
}

func (l *logger) Options(enabled, printQuery, printResult bool) *logger {
	l.enabled = enabled
	l.printQuery = printQuery
	l.printResult = printResult

	return l
}

func (l *logger) LogBegin(msg, method, url string, jsonQuery []byte) {
	if !l.enabled {
		return
	}

	l.Printf("\n[Arangolite] %s %s %s | %s %s", blue, msg, reset, method, url)

	if l.printQuery && jsonQuery != nil && len(jsonQuery) != 0 {
		l.Println("    " + string(indentJSON(jsonQuery)))
	}
}

func (l *logger) LogResult(cached bool, start time.Time, in, out chan interface{}) {
	batchNb := 0
	firstBatch := []byte{}

	for {
		tmp := <-in
		out <- tmp

		switch t := tmp.(type) {
		case json.RawMessage:
			batchNb++
			if batchNb == 1 {
				firstBatch = t
			}
			continue
		}
		break
	}

	if !l.enabled {
		return
	}

	execTime := time.Now().Sub(start)

	if cached {
		l.Printf("\n[Arangolite] %s RESULT %s | %s CACHED %s | Execution: %v | Batches: %d",
			blue, reset, yellow, reset, execTime, batchNb)
	} else {
		l.Printf("\n[Arangolite] %s RESULT %s | Execution: %v | Batches: %d",
			blue, reset, execTime, batchNb)
	}

	if l.printResult {
		content := "    " + string(indentJSON(firstBatch))
		if batchNb > 1 {
			content += "\n\n    Result has been truncated to the first batch"
		}
		l.Println(content)
	}
}

func (l *logger) LogError(errMsg string, start time.Time) {
	if !l.enabled {
		return
	}

	l.Printf("\n[Arangolite] %s ERROR %s | Execution: %v | Message: %s",
		red, reset, time.Now().Sub(start), errMsg)
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
	json.Indent(b, in, "    ", "  ")

	return b.Bytes()
}
