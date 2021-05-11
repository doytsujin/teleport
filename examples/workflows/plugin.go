/*
Copyright 2019 Gravitational, Inc.

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

// Example whitelist based access plugin.
//
// This plugin approves/denies access requests based on a simple whitelist
// of usernames. Requests from whitelisted users are approved, all others
// are denied.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/client/workflows"

	"github.com/gravitational/trace"
	"github.com/pelletier/go-toml"
)

// eprintln prints an optionally formatted string to stderr.
func eprintln(msg string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, a...)
	fmt.Fprintf(os.Stderr, "\n")
}

func main() {
	ctx := context.Background()
	cfg, err := loadConfig("config.toml")
	if err != nil {
		eprintln("ERROR: %s", err)
		os.Exit(1)
	}
	if err := run(ctx, cfg); err != nil {
		eprintln("ERROR: %s", err)
		os.Exit(1)
	}
}

type Config struct {
	Addr         string   `toml:"addr"`
	IdentityFile string   `toml:"identity_file"`
	Whitelist    []string `toml:"whitelist"`
}

func loadConfig(filepath string) (*Config, error) {
	t, err := toml.LoadFile(filepath)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	conf := &Config{}
	if err := t.Unmarshal(conf); err != nil {
		return nil, trace.Wrap(err)
	}
	return conf, nil
}

func run(ctx context.Context, cfg *Config) error {
	client, err := client.New(ctx, client.Config{
		Addrs: []string{cfg.Addr},
		Credentials: []client.Credentials{
			client.LoadIdentityFile(cfg.IdentityFile),
		},
		InsecureAddressDiscovery: true,
		DialTimeout:              time.Second * 5,
	})
	if err != nil {
		return trace.Wrap(err)
	}
	defer client.Close()

	plugin := workflows.NewPlugin(ctx, "example", client)

	// Register a watcher for pending access requests.
	watcher, err := plugin.WatchRequests(ctx, workflows.Filter{
		State: workflows.StatePending,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	if err := watcher.WaitInit(ctx, 5*time.Second); err != nil {
		return trace.Wrap(err)
	}
	eprintln("watcher initialized...")
	defer watcher.Close()

	for {
		select {
		case event := <-watcher.Events():
			req, op := event.Request, event.Type
			switch op {
			case workflows.OpPut:
				// OpPut indicates that a request has been created or updated.  Since we specified
				// StatePending in our filter, only pending requests should appear here.
				eprintln("Handling request: %+v", req)
				whitelisted := false
			CheckWhitelist:
				for _, user := range cfg.Whitelist {
					if req.User == user {
						whitelisted = true
						break CheckWhitelist
					}
				}
				params := workflows.RequestUpdate{
					Annotations: map[string][]string{
						"strategy": {"whitelist"},
					},
				}
				if whitelisted {
					eprintln("User %q in whitelist, approving request...", req.User)
					params.State = workflows.StateApproved
					params.Reason = "user in whitelist"
				} else {
					eprintln("User %q not in whitelist, denying request...", req.User)
					params.State = workflows.StateDenied
					params.Reason = "user not in whitelist"
				}
				if err := plugin.SetRequestState(ctx, req.ID, "delegator", params); err != nil {
					return trace.Wrap(err)
				}
				eprintln("ok.")
			case workflows.OpDelete:
				// request has been removed (expired).
				// Only the ID is non-zero in this case.
				// Due to some limitations in Teleport's event system, filters
				// don't really work with OpDelete events. As such, we may get
				// OpDelete events for requests that would not typically match
				// the filter argument we supplied above.
				eprintln("Request %s has automatically expired.", req.ID)
			default:
				return trace.BadParameter("unexpected event operation %s", op)
			}
		case <-watcher.Done():
			return watcher.Error()
		}
	}
}
