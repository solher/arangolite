package arangolite

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func longRequest(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Waiting 2 seconds...")
	time.Sleep(200 * time.Millisecond)
	panic("The request was not canceled")
}

func runWebserver() {
	go func() {
		http.HandleFunc("/", longRequest)
		http.ListenAndServe(":9999", nil)
	}()
}

func TestSendCanBeCanceled(t *testing.T) {
	runWebserver()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:9999", nil)

	sender := basicSender{}
	parent := context.Background()
	ctx, cancel := context.WithTimeout(parent, 100*time.Millisecond)
	defer cancel()

	resp, err := sender.Send(ctx, client, req)

	assertEqual(t, err.Error(), "the database HTTP request failed: Get http://localhost:9999: context deadline exceeded")
	assertTrue(t, resp == nil, "The response of a canceled request should be nil")
}
