package client

import (
	"strconv"
	"sync/atomic"

	"github.com/go-resty/resty/v2"
)

// GraphVersionHeader carries the Console permission-graph version. The server
// stamps it on every response; echoing the highest value seen back on the next
// request makes Console hold authorization until its own cache has caught up,
// which gives a CLI run read-your-writes causality even when consecutive
// requests land on different Console nodes.
const GraphVersionHeader = "X-Cdk-Graph-Version"

// graphVersionTracker keeps the highest graph version observed so far. Applies
// run concurrently on a shared client (--max-parallel), so it must be safe for
// use from several goroutines.
type graphVersionTracker struct {
	version atomic.Int64
}

// observe records v when it is newer than what we already saw.
func (t *graphVersionTracker) observe(v int64) {
	for {
		current := t.version.Load()
		if v <= current {
			return
		}
		if t.version.CompareAndSwap(current, v) {
			return
		}
	}
}

func (t *graphVersionTracker) current() int64 {
	return t.version.Load()
}

// installGraphVersionEcho wires the tracker into a resty client so that every
// call made through it participates, whatever command issued it.
func installGraphVersionEcho(restyClient *resty.Client, tracker *graphVersionTracker) {
	restyClient.OnBeforeRequest(func(_ *resty.Client, request *resty.Request) error {
		if version := tracker.current(); version > 0 {
			request.SetHeader(GraphVersionHeader, strconv.FormatInt(version, 10))
		}
		return nil
	})
	restyClient.OnAfterResponse(func(_ *resty.Client, response *resty.Response) error {
		header := response.Header().Get(GraphVersionHeader)
		if header == "" {
			return nil
		}
		// A header we cannot read is not worth failing a request over.
		if version, err := strconv.ParseInt(header, 10, 64); err == nil {
			tracker.observe(version)
		}
		return nil
	})
}
