/*
Copyright 2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package workflows

import (
	"context"
	"sync"
	"time"

	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/types"

	"github.com/gravitational/trace"
)

// RequestWatcher is a stream watcher for Teleport access requests.
type RequestWatcher interface {
	// WaitInit waits for the Init operation to complete
	// or returns an error if the watcher fails to init.
	WaitInit(ctx context.Context, timeout time.Duration) error
	// Events returns a channel of RequestEvents.
	Events() <-chan RequestEvent
	// Done returns a channel that is closed when the watcher is done.
	Done() <-chan struct{}
	// Error returns the last error of the requestWatcher.
	Error() error
	// close the watcher.
	Close() error
}

// requestWatcher is a thin wrapper around types.Watcher
// used for watching access request events.
type requestWatcher struct {
	types.Watcher
	initC   chan struct{}
	eventsC chan RequestEvent
	emux    sync.Mutex
	err     error
}

// RequestEvent is a request event.
type RequestEvent struct {
	// Type is the operation type of the event.
	Type OpType
	// Request is the payload of the event. If Type
	// is OpDelete, only the ID field will be filled.
	Request Request
}

// NewRequestWatcher creates a new RequestWatcher using the given client and filter.
func NewRequestWatcher(ctx context.Context, clt *client.Client, filter Filter) (RequestWatcher, error) {
	eventWatcher, err := clt.NewWatcher(ctx,
		types.Watch{
			Kinds: []types.WatchKind{
				{
					Kind:   types.KindAccessRequest,
					Filter: filter.IntoMap(),
				},
			},
		},
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	w := &requestWatcher{
		Watcher: eventWatcher,
		initC:   make(chan struct{}),
		eventsC: make(chan RequestEvent),
	}
	go w.receiveEvents()
	return w, nil
}

// receiveEvents recieves events from the client stream and sends
// the associated RequestEvents to the requestWatcher's eventC channel.
func (w *requestWatcher) receiveEvents() {
	for {
		event := <-w.Watcher.Events()
		var req Request
		switch event.Type {
		case types.OpInit:
			close(w.initC)
			continue
		case types.OpPut:
			r, ok := event.Resource.(types.AccessRequest)
			if !ok {
				w.setError(trace.Errorf("unexpected resource type %T", event.Resource))
				return
			}
			req = requestFromAccessRequest(r)
		case types.OpDelete:
			req = Request{
				ID: event.Resource.GetName(),
			}
		default:
			w.setError(trace.Errorf("unexpected event op type %s", event.Type))
			return
		}
		w.eventsC <- RequestEvent{
			Type:    event.Type,
			Request: req,
		}
	}
}

// WaitInit waits for the Init operation to complete
// or returns an error if the watcher fails to init.
func (w *requestWatcher) WaitInit(ctx context.Context, timeout time.Duration) error {
	select {
	case <-w.initC:
		return nil
	case <-time.After(timeout):
		return trace.ConnectionProblem(nil, "requestWatcher initialization timed out")
	case <-w.Done():
		return w.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Events returns a channel of RequestEvents.
func (w *requestWatcher) Events() <-chan RequestEvent {
	return w.eventsC
}

// Error returns the last error of the requestWatcher.
func (w *requestWatcher) Error() error {
	w.emux.Lock()
	defer w.emux.Unlock()
	if w.err != nil {
		return w.err
	}
	return w.Watcher.Error()
}

// setError sets the requestWatcher error.
func (w *requestWatcher) setError(err error) {
	w.emux.Lock()
	defer w.emux.Unlock()
	w.err = err
}
