package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	assertion "github.com/stretchr/testify/assert"

	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
)

func TestPanicRecoverHandler(t *testing.T) {
	assert := assertion.New(t)
	var loggerBuffer strings.Builder
	log, _ := logger.NewZerologLogger(&logger.Config{
		Level:         "debug",
		Output:        &loggerBuffer,
		OmitTimestamp: true,
	})

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("error")
	})
	c := server.Chain(hf, RecoverOnPanic(log))
	req := httptest.NewRequest(http.MethodGet, "http://localhost/", http.NoBody)
	w := httptest.NewRecorder()

	c.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(resp.StatusCode, http.StatusInternalServerError)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(string(body), `{"message":"Internal Server Error"}`)
	assert.Equal(loggerBuffer.String(), `{"level":"error","panic":"error","message":"panic inside of HTTP"}`+"\n")
}
