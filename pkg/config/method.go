package config

// GetMCPServers returns a map of MCP servers based on the `shared` parameter.
// If `shared` is true, it returns servers that are shared across sessions.
// Otherwise, it returns servers that are not shared.
func (c Config) GetMCPServers(shared bool) MCPServers {
	// Return a copy of the MCP servers to avoid external modifications
	sharedServers := make(MCPServers)
	for k, v := range c.MCPServers {
		if shared != v.IsShared() {
			continue
		}

		sharedServers[k] = v
	}
	return sharedServers
}

// IsShared returns true if the MCP server is shared across sessions.
// A server is considered shared if the `Shared` field is nil or if it is
// explicitly set to true.
func (s MCPServer) IsShared() bool {
	return s.Shared == nil || *s.Shared
}
