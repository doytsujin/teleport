/*
Copyright 2017-2019 Gravitational, Inc.

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

package services

import (
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/lib/utils"
)

// UnmarshalClusterName unmarshals the ClusterName resource from JSON.
func UnmarshalClusterName(bytes []byte, opts ...MarshalOption) (ClusterName, error) {
	if len(bytes) == 0 {
		return nil, trace.BadParameter("missing resource data")
	}

	cfg, err := CollectOptions(opts)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	var h ResourceHeader
	if err = utils.FastUnmarshal(bytes, &h); err != nil {
		return nil, trace.Wrap(err)
	}

	switch h.Version {
	case V2:
		var clusterName ClusterNameV2
		if err := utils.FastUnmarshal(bytes, &clusterName); err != nil {
			return nil, trace.BadParameter(err.Error())
		}
		if err := clusterName.CheckAndSetDefaults(); err != nil {
			return nil, trace.Wrap(err)
		}
		if cfg.ID != 0 {
			clusterName.SetResourceID(cfg.ID)
		}
		if !cfg.Expires.IsZero() {
			clusterName.SetExpiry(cfg.Expires)
		}
		return &clusterName, nil
	default:
		return nil, trace.BadParameter("cluster name resource version %v is not supported", h.Version)
	}
}

// MarshalClusterName marshals the ClusterName resource to JSON.
func MarshalClusterName(clusterName ClusterName, opts ...MarshalOption) ([]byte, error) {
	if err := clusterName.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	cfg, err := CollectOptions(opts)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	switch clusterName := clusterName.(type) {
	case *ClusterNameV2:
		if !cfg.PreserveResourceID {
			// avoid modifying the original object
			// to prevent unexpected data races
			copy := *clusterName
			copy.SetResourceID(0)
			clusterName = &copy
		}
		return utils.FastMarshal(clusterName)
	default:
		return nil, trace.BadParameter("unrecognized cluster name version %T", clusterName)
	}
}
