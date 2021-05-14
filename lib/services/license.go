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

package services

import (
	"encoding/json"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/lib/utils"
)

// UnmarshalLicense unmarshals the License resource from JSON.
func UnmarshalLicense(bytes []byte) (License, error) {
	if len(bytes) == 0 {
		return nil, trace.BadParameter("missing resource data")
	}

	var h ResourceHeader
	if err := json.Unmarshal(bytes, &h); err != nil {
		return nil, trace.Wrap(err)
	}

	switch h.Version {
	case V3:
		var license LicenseV3
		if err := utils.FastUnmarshal(bytes, &license); err != nil {
			return nil, trace.Wrap(err)
		}
		if err := license.CheckAndSetDefaults(); err != nil {
			return nil, trace.Wrap(err)
		}
		return &license, nil
	default:
		return nil, trace.BadParameter(
			"License resource version %q is not supported", h.Version)
	}
}

// MarshalLicense marshals the License resource to JSON.
func MarshalLicense(license License, opts ...MarshalOption) ([]byte, error) {
	if err := license.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	cfg, err := CollectOptions(opts)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	switch license := license.(type) {
	case *LicenseV3:
		if !cfg.PreserveResourceID {
			// avoid modifying the original object
			// to prevent unexpected data races
			copy := *license
			copy.SetResourceID(0)
			license = &copy
		}
		return utils.FastMarshal(license)
	default:
		return nil, trace.BadParameter("unrecognized license version %T", license)
	}
}
