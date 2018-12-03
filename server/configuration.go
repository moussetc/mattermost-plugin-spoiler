package main

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-server/model"
	"github.com/pkg/errors"
)

// Plugin configuration
//
// May be changed at anytime, beware concurrent calls by asynchronous hooks
type Configuration struct {
	SpoilerMode string
}

// Clone shallow copies the configuration.
func (c *Configuration) Clone() *Configuration {
	var clone = *c
	return &clone
}

// getConfiguration retrieves the active configuration under lock
//
// The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *Configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &Configuration{}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
//
// This method panics if setConfiguration is called with the existing configuration. This almost
// certainly means that the configuration was modified without being cloned and may result in
// an unsafe access.
func (p *Plugin) setConfiguration(configuration *Configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(Configuration)

	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	p.setConfiguration(configuration)

	p.emitConfigChange()

	return nil
}

// handleConfigRequest answers a HTTP request for the plugin's configuration
func (p *Plugin) handleConfigRequest(w http.ResponseWriter, r *http.Request) {
	configuration := p.getConfiguration()
	var response = struct {
		SpoilerMode string `json:"spoilerMode"`
	}{
		SpoilerMode: configuration.SpoilerMode,
	}
	responseJSON, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// emitConfigChange alerts the frontend that the configuration has changed
func (p *Plugin) emitConfigChange() {
	configuration := p.getConfiguration()
	p.API.PublishWebSocketEvent("config_change", map[string]interface{}{
		"spoilerMode": configuration.SpoilerMode,
	}, &model.WebsocketBroadcast{})
}
