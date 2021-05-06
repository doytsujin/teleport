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

package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/gravitational/teleport/api/defaults"
	"github.com/gravitational/teleport/api/utils"

	"github.com/gravitational/trace"
)

// SessionRecordingConfig defines session recording configuration. This is
// a configuration resource, never create more than one instance of it.
type SessionRecordingConfig interface {
	Resource

	// GetMode gets the session recording mode.
	GetMode() string

	// SetMode sets the session recording mode.
	SetMode(string)

	// GetProxyChecksHostKeys gets if the proxy will check host keys.
	GetProxyChecksHostKeys() string

	// SetProxyChecksHostKeys sets if the proxy will check host keys.
	SetProxyChecksHostKeys(string)

	// CheckAndSetDefaults sets and default values and then
	// verifies the constraints for SessionRecordingConfig.
	CheckAndSetDefaults() error
}

// NewSessionRecordingConfig is a convenience method to to create SessionRecordingConfigV2.
func NewSessionRecordingConfig(spec SessionRecordingConfigSpecV2) (SessionRecordingConfig, error) {
	recConfig := SessionRecordingConfigV2{
		Kind:    KindSessionRecordingConfig,
		Version: V2,
		Metadata: Metadata{
			Name:      MetaNameSessionRecordingConfig,
			Namespace: defaults.Namespace,
		},
		Spec: spec,
	}

	if err := recConfig.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	return &recConfig, nil
}

// DefaultSessionRecordingConfig returns the default session recording config.
func DefaultSessionRecordingConfig() SessionRecordingConfig {
	config, _ := NewSessionRecordingConfig(SessionRecordingConfigSpecV2{})
	return config
}

// GetVersion returns resource version.
func (c *SessionRecordingConfigV2) GetVersion() string {
	return c.Version
}

// GetName returns the name of the resource.
func (c *SessionRecordingConfigV2) GetName() string {
	return c.Metadata.Name
}

// SetName sets the name of the resource.
func (c *SessionRecordingConfigV2) SetName(e string) {
	c.Metadata.Name = e
}

// SetExpiry sets expiry time for the object.
func (c *SessionRecordingConfigV2) SetExpiry(expires time.Time) {
	c.Metadata.SetExpiry(expires)
}

// Expiry returns object expiry setting.
func (c *SessionRecordingConfigV2) Expiry() time.Time {
	return c.Metadata.Expiry()
}

// SetTTL sets Expires header using the provided clock.
// Use SetExpiry instead.
// DELETE IN 7.0.0
func (c *SessionRecordingConfigV2) SetTTL(clock Clock, ttl time.Duration) {
	c.Metadata.SetTTL(clock, ttl)
}

// GetMetadata returns object metadata.
func (c *SessionRecordingConfigV2) GetMetadata() Metadata {
	return c.Metadata
}

// GetResourceID returns resource ID.
func (c *SessionRecordingConfigV2) GetResourceID() int64 {
	return c.Metadata.ID
}

// SetResourceID sets resource ID.
func (c *SessionRecordingConfigV2) SetResourceID(id int64) {
	c.Metadata.ID = id
}

// GetKind returns resource kind.
func (c *SessionRecordingConfigV2) GetKind() string {
	return c.Kind
}

// GetSubKind returns resource subkind.
func (c *SessionRecordingConfigV2) GetSubKind() string {
	return c.SubKind
}

// SetSubKind sets resource subkind.
func (c *SessionRecordingConfigV2) SetSubKind(sk string) {
	c.SubKind = sk
}

// GetMode gets the cluster's session recording mode.
func (c *SessionRecordingConfigV2) GetMode() string {
	return c.Spec.Mode
}

// SetMode sets the cluster's session recording mode.
func (c *SessionRecordingConfigV2) SetMode(m string) {
	c.Spec.Mode = m
}

// GetProxyChecksHostKeys sets if the proxy will check host keys.
func (c *SessionRecordingConfigV2) GetProxyChecksHostKeys() string {
	return c.Spec.ProxyChecksHostKeys
}

// SetProxyChecksHostKeys sets if the proxy will check host keys.
func (c *SessionRecordingConfigV2) SetProxyChecksHostKeys(t string) {
	c.Spec.ProxyChecksHostKeys = t
}

// CheckAndSetDefaults verifies the constraints for SessionRecordingConfig.
func (c *SessionRecordingConfigV2) CheckAndSetDefaults() error {
	// Make sure we have defaults for all metadata fields.
	err := c.Metadata.CheckAndSetDefaults()
	if err != nil {
		return trace.Wrap(err)
	}

	if c.Spec.Mode == "" {
		c.Spec.Mode = RecordAtNode
	}
	if c.Spec.ProxyChecksHostKeys == "" {
		c.Spec.ProxyChecksHostKeys = HostKeyCheckYes
	}

	// Check that the session recording mode is set to a valid value.
	if !utils.SliceContainsStr(SessionRecordingModes, c.Spec.Mode) {
		return trace.BadParameter("session recording mode must be one of: %v", strings.Join(SessionRecordingModes, ","))
	}

	// Check that proxy_checks_host_keys is set to a valid value.
	if !utils.SliceContainsStr(ProxyChecksHostKeysValues, c.Spec.ProxyChecksHostKeys) {
		return trace.BadParameter("proxy_checks_host_keys must be one of: %v", strings.Join(ProxyChecksHostKeysValues, ","))
	}

	return nil
}

// String returns string representation of session recording configuration.
func (c *SessionRecordingConfigV2) String() string {
	return fmt.Sprintf("SessionRecordingConfig(Mode=%v,ProxyChecksHostKeys=%v)", c.Spec.Mode, c.Spec.ProxyChecksHostKeys)
}
