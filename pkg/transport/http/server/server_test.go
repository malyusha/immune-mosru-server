package server

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

func TestNew(t *testing.T) {
	shutdownTimeout := time.Millisecond * 100

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	srv := New(
		WithLogger(new(logger.Empty)),
		WithAddr(":10050"),
		WithHandler(handler),
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := srv.Start(ctx)
	go func() {
		assert.NoError(t, <-done)
	}()

	time.Sleep(time.Millisecond * 300) // wait till server runs
	req, err := http.NewRequest(http.MethodGet, "http://localhost:10050", http.NoBody)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, resp.Body.Close())
	assert.NoError(t, err)

	assert.Equal(t, "OK", string(b))

	// This should stop server from accepting all further incoming connections
	cancel()
	time.Sleep(shutdownTimeout) // just to be sure, that server is down

	resp, err = http.DefaultClient.Do(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
