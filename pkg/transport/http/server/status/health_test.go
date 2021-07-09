package status

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	assertions "github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	assert := assertions.New(t)
	ready := make(chan struct{})
	h := newHealthHandler(ready)
	assert.NotNil(h)

	req := httptest.NewRequest(http.MethodGet, "http://localhost/readiness", http.NoBody)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(http.StatusServiceUnavailable, resp.StatusCode)
	assert.Empty("", string(body))

	ready <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	req = httptest.NewRequest(http.MethodGet, "http://localhost/readiness", http.NoBody)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Empty(string(body))

	req = httptest.NewRequest(http.MethodGet, "http://localhost/liveness", http.NoBody)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(resp.StatusCode, http.StatusOK)
	assert.Empty(string(body))
}
