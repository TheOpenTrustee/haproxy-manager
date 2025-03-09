package util

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

// FeatureFlags holds the state of feature flags.
type FeatureFlags struct {
	flags map[string]bool
	mu    sync.RWMutex
}

// New creates a new FeatureFlags instance.
func NewFeatureFlags() *FeatureFlags {
	return &FeatureFlags{
		flags: map[string]bool{
			"dashboard":         true,
			"hosts":             false,
			"access_lists":      false,
			"ssl_certs":         false,
			"users":             false,
			"audit_logs":        false,
			"settings":          false,
			"streams":           false,
			"404_hosts":         false,
			"redirection_hosts": false,
		},
	}
}

// Get all feature flags
func (f *FeatureFlags) GetAll() map[string]bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.flags
}

// Set sets the value of a feature flag.
func (f *FeatureFlags) Set(flag string, enabled bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flags[flag] = enabled
}

// IsEnabled checks if a feature flag is enabled.
func (f *FeatureFlags) IsEnabled(flag string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.flags[flag]
}

// Toggle toggles the state of a feature flag.
func (f *FeatureFlags) Toggle(flag string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flags[flag] = !f.flags[flag]
}

// LoadFlagsFromFile loads feature flags from a JSON or YAML file.
func (f *FeatureFlags) LoadFlagsFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var data map[string]bool
	if err := json.NewDecoder(file).Decode(&data); err == nil {
		for flag, enabled := range data {
			f.Set(flag, enabled)
		}
		return nil
	}

	// If JSON decoding fails, try YAML
	file.Seek(0, 0) // Reset file pointer for YAML
	if err := yaml.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("failed to parse feature flag file: %v", err)
	}

	for flag, enabled := range data {
		f.Set(flag, enabled)
	}
	return nil
}

// WatchFile watches the feature flag file for changes and reloads it.
func (f *FeatureFlags) WatchFile(filePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	if err := watcher.Add(filePath); err != nil {
		return fmt.Errorf("failed to watch file: %v", err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("Feature flag file changed, reloading...")
				if err := f.LoadFlagsFromFile(filePath); err != nil {
					fmt.Printf("Error reloading file: %v\n", err)
				}
			}
		case err := <-watcher.Errors:
			fmt.Printf("Error watching file: %v\n", err)
		}
	}
}
