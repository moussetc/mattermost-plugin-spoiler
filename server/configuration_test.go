package main

import (
	"errors"
	"testing"

	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
)

func TestConfigurationSetterGetter(t *testing.T) {
	p := Plugin{}

	originalConfig := &Configuration{
		SpoilerMode: "fkjsdlkjdsvqkj",
	}

	assert.NotNil(t, p.getConfiguration())

	p.setConfiguration(originalConfig)

	config := p.getConfiguration()
	assert.NotNil(t, config)
	assert.Equal(t, originalConfig.SpoilerMode, config.SpoilerMode)
}

func TestOnConfigurationChange(t *testing.T) {
	api := &plugintest.API{}

	originalConfig := &Configuration{
		SpoilerMode: "fkjsdlkjdsvqkj",
	}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(nil).Run(func(args mock.Arguments) {
		config := args.Get(0).(*Configuration)
		config.SpoilerMode = originalConfig.SpoilerMode
	})

	api.On("PublishWebSocketEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	p := Plugin{}

	p.SetAPI(api)
	_ = p.OnConfigurationChange()

	config := p.getConfiguration()
	assert.NotNil(t, config)
	assert.Equal(t, originalConfig.SpoilerMode, config.SpoilerMode)

	apiError := &plugintest.API{}

	errAPI := errors.New("oops LoadPluginConfiguration failed")
	apiError.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(errAPI)

	p.SetAPI(apiError)
	err := p.OnConfigurationChange()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), errAPI.Error())
}
