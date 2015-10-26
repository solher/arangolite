package arangolite

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestLogger runs tests on the Arangolite logger.
func TestLogger(t *testing.T) {
	// a := assert.New(t)
	r := require.New(t)
	output := bytes.NewBuffer(nil)
	logger := newLogger()
	logger.Options(false, false, false).SetOutput(output)
	in := make(chan interface{}, 2)
	out := make(chan interface{}, 2)

	logger.LogBegin("TEST", "POST", "http://test.fr", []byte("QUERY"))
	r.Empty(output.String())
	output.Reset()

	go logger.LogResult(false, time.Now(), in, out)
	in <- json.RawMessage("Hello world !")
	in <- nil
	<-out
	<-out
	r.Empty(output.String())
	output.Reset()

	logger.LogError("ERROR", time.Now())
	r.Empty(output.String())
	output.Reset()

	logger.Options(true, false, false)

	logger.LogBegin("TEST", "POST", "http://test.fr", []byte(`{"foo":"bar"}`))
	r.Contains(output.String(), "TEST")
	r.Contains(output.String(), "http://test.fr")
	r.NotContains(output.String(), "foo")
	output.Reset()

	go logger.LogResult(true, time.Now(), in, out)
	in <- json.RawMessage{}
	in <- nil
	<-out
	<-out
	r.NotContains(output.String(), "[]")
	output.Reset()

	logger.LogError("ERROR", time.Now())
	r.Contains(output.String(), "ERROR")
	output.Reset()

	logger.Options(true, true, true)

	logger.LogBegin("TEST", "POST", "http://test.fr", []byte(`{"foo":"bar"}`))
	r.Contains(output.String(), "TEST")
	r.Contains(output.String(), "http://test.fr")
	r.Contains(output.String(), "foo")
	output.Reset()

	go logger.LogResult(false, time.Now(), in, out)
	in <- json.RawMessage(`{"foo":"bar"}`)
	in <- json.RawMessage{}
	in <- nil
	<-out
	<-out
	<-out
	time.Sleep(time.Millisecond)
	r.Contains(output.String(), "foo")
	output.Reset()
}
