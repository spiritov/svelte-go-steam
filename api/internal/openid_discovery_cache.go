package internal

import "github.com/yohcop/openid-go"

// NoOpDiscoveryCache implements the DiscoveryCache interface and doesn't cache anything.
type NoOpDiscoveryCache struct{}

// Put is a no op.
func (n *NoOpDiscoveryCache) Put(id string, info openid.DiscoveredInfo) {}

// Get always returns nil.
func (n *NoOpDiscoveryCache) Get(id string) openid.DiscoveredInfo {
	return nil
}
