package output

import (
	"fmt"
	"sort"
	"strings"
)

// Registry manages all available output plugins
type Registry struct {
	plugins map[string]OutputPlugin
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]OutputPlugin),
	}
}

// Register adds a plugin to the registry
func (r *Registry) Register(plugin OutputPlugin) error {
	id := plugin.ID()
	if id == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	if _, exists := r.plugins[id]; exists {
		return fmt.Errorf("plugin with ID '%s' already registered", id)
	}

	r.plugins[id] = plugin
	return nil
}

// Get retrieves a plugin by its ID
func (r *Registry) Get(id string) (OutputPlugin, error) {
	plugin, exists := r.plugins[id]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", id)
	}
	return plugin, nil
}

// List returns all registered plugins sorted by ID
func (r *Registry) List() []OutputPlugin {
	plugins := make([]OutputPlugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	// Sort by ID for consistent ordering
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].ID() < plugins[j].ID()
	})

	return plugins
}

// GetIDs returns a list of all registered plugin IDs
func (r *Registry) GetIDs() []string {
	ids := make([]string, 0, len(r.plugins))
	for id := range r.plugins {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Exists checks if a plugin with the given ID is registered
func (r *Registry) Exists(id string) bool {
	_, exists := r.plugins[id]
	return exists
}

// FormatList returns a formatted string listing all plugins for CLI help
func (r *Registry) FormatList() string {
	if len(r.plugins) == 0 {
		return "No output plugins registered"
	}

	var builder strings.Builder
	builder.WriteString("Available output formats:\n")

	plugins := r.List()
	for _, plugin := range plugins {
		builder.WriteString(fmt.Sprintf("  %-12s %s (*.%s)\n",
			plugin.ID(),
			plugin.Description(),
			plugin.FileExtension()))
	}

	return builder.String()
}

// GetDefaultPlugin returns a reasonable default plugin (first alphabetically)
func (r *Registry) GetDefaultPlugin() (OutputPlugin, error) {
	if len(r.plugins) == 0 {
		return nil, fmt.Errorf("no plugins registered")
	}

	plugins := r.List()
	return plugins[0], nil
}

// Global registry instance
var globalRegistry = NewRegistry()

// Register registers a plugin with the global registry
func Register(plugin OutputPlugin) error {
	return globalRegistry.Register(plugin)
}

// Get retrieves a plugin from the global registry
func Get(id string) (OutputPlugin, error) {
	return globalRegistry.Get(id)
}

// List returns all plugins from the global registry
func List() []OutputPlugin {
	return globalRegistry.List()
}

// GetIDs returns all plugin IDs from the global registry
func GetIDs() []string {
	return globalRegistry.GetIDs()
}

// Exists checks if a plugin exists in the global registry
func Exists(id string) bool {
	return globalRegistry.Exists(id)
}

// FormatList returns a formatted list from the global registry
func FormatList() string {
	return globalRegistry.FormatList()
}

// GetDefaultPlugin returns the default plugin from the global registry
func GetDefaultPlugin() (OutputPlugin, error) {
	return globalRegistry.GetDefaultPlugin()
}

// GetGlobalRegistry returns the global registry instance
func GetGlobalRegistry() *Registry {
	return globalRegistry
}