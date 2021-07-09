package middleware

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	assertion "github.com/stretchr/testify/assert"

	"github.com/malyusha/immune-mosru-server/pkg/contextutils"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
)

func TestChain(t *testing.T) {
	assert := assertion.New(t)

	var mw []server.Middleware
	for i := 0; i < 10; i++ {
		i := i
		m := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "%d ", i)
				next.ServeHTTP(w, r)
			})
		}
		mw = append(mw, m)
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "h")
	})

	c := server.Chain(h, mw...)

	req := httptest.NewRequest(http.MethodGet, "http://localhost/", http.NoBody)
	w := httptest.NewRecorder()

	c.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal("0 1 2 3 4 5 6 7 8 9 h", string(body))
}

func TestPassRequestIDToContext(t *testing.T) {
	fakeRequestID := "test"
	assert := assertion.New(t)
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(fakeRequestID, contextutils.GetTraceID(r.Context()))
		w.Write([]byte(contextutils.GetTraceID(r.Context())))
	})

	gen := StringGenerator(func() string { return fakeRequestID })
	c := server.Chain(hf, PassTraceIDToContext(gen))
	req := httptest.NewRequest(http.MethodGet, "http://localhost/", http.NoBody)
	w := httptest.NewRecorder()

	c.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(fakeRequestID, string(body))
}

func TestPassLoggingFieldsToRequestContext(t *testing.T) {
	assert := assertion.New(t)
	var loggerBuffer strings.Builder

	log, _ := logger.NewZerologLogger(&logger.Config{
		Level:         "debug",
		Output:        &loggerBuffer,
		OmitTimestamp: true,
	})

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithContext(r.Context()).Debug("msg")
	})

	extractor := func(ctx context.Context) logger.Fields {
		return logger.Fields{"ctx": ctx.Value("ctx")}
	}

	c := server.Chain(hf, PassLoggingFieldsToRequestContext(extractor))
	ctx := context.WithValue(context.Background(), "ctx", true)
	req := httptest.NewRequest(http.MethodGet, "http://test", http.NoBody).WithContext(ctx)
	w := httptest.NewRecorder()

	c.ServeHTTP(w, req)

	expectLog := `{"level":"debug","ctx":true,"message":"msg"}` + "\n"
	assert.Equal(expectLog, loggerBuffer.String())
}
