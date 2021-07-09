package server

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

const (
	defaultListenAddr      = ":8080"
	defaultShutdownTimeout = 5 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultReadTimeout     = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
)

var srvNum = new(uint32)

var defaultOpts = &options{
	listenAddr: defaultListenAddr,
	timeouts: &Timeouts{
		Shutdown: defaultShutdownTimeout,
		Read:     defaultReadTimeout,
		Write:    defaultWriteTimeout,
		Idle:     defaultIdleTimeout,
	},
}

type Option func(*options)

// noop
var noopOpt = func(o *options) {}

type options struct {
	name       string
	listenAddr string
	handler    http.Handler
	logger     logger.Logger
	timeouts   *Timeouts
}

// evaluateOptions returns defaultServerOptions, replaced with values of given.
func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOpts
	for _, o := range opts {
		o(optCopy)
	}

	if optCopy.name == "" {
		optCopy.name = fmt.Sprintf("№%d", atomic.AddUint32(srvNum, 1))
	}

	if optCopy.logger == nil {
		// don't initialize it as global variable because it would be evaluated before
		// global logger is configured.
		optCopy.logger = logger.GetLogger()
	}
	optCopy.logger = optCopy.logger.With(logger.Fields{"server": optCopy.name})

	return optCopy
}

func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

func WithHandler(h http.Handler) Option {
	return func(o *options) {
		o.handler = h
	}
}

func WithLogger(l logger.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

// WithAddr configures listen address.
func WithAddr(addr string) Option {
	if addr == "" {
		return noopOpt
	}

	return func(opts *options) {
		opts.listenAddr = addr
	}
}

// represents configuration of server timeouts.
type Timeouts struct {
	Shutdown,
	Read,
	Write,
	Idle time.Duration
}

func WithTimeouts(timeouts Timeouts) Option {
	return func(opts *options) {
		ts := opts.timeouts
		if timeouts.Idle != 0 {
			ts.Idle = timeouts.Idle
		}
		if timeouts.Read != 0 {
			ts.Read = timeouts.Read
		}
		if timeouts.Write != 0 {
			ts.Write = timeouts.Write
		}
		if timeouts.Shutdown != 0 {
			ts.Shutdown = timeouts.Shutdown
		}
	}
}
