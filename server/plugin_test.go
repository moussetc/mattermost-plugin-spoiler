package main

import (
	"github.com/mattermost/mattermost-server/model"

	"testing"

	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
)

// TODO : uncomment when OnConfigurationChange is fixed
/*
func TestOnConfigurationChange(t *testing.T) {

	configuration := Configuration{}

	api := &plugintest.API{}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(func(dest interface{}) error {
		*dest.(*Configuration) = configuration
		return nil
	})

	api.On("RegisterCommand", mock.Anything).Return(nil)

	p := Plugin{}
	p.SetAPI(api)

	assert.NoError(t, p.OnConfigurationChange())

	assert.NotNil(t, p.getConfiguration())
	assert.Equal(t, configuration, *p.getConfiguration())
}

func TestOnConfigurationChangeError(t *testing.T) {

	configuration := Configuration{}

	api := &plugintest.API{}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(&model.AppError{Message: "argh"})

	api.On("RegisterCommand", mock.Anything).Return(nil)

	p := Plugin{}
	p.SetAPI(api)

	assert.Error(t, p.OnConfigurationChange())
	assert.NotNil(t, p.getConfiguration())
	assert.Equal(t, configuration, *p.getConfiguration())
}
*/

func TestExecuteCommand(t *testing.T) {
	api := &plugintest.API{}

	var post *model.Post
	api.On("RegisterCommand", mock.Anything).Return(nil)
	api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, nil).Run(func(args mock.Arguments) {
		post = args.Get(0).(*model.Post)
	})

	p := Plugin{}

	p.SetAPI(api)

	assert.Nil(t, p.OnActivate())

	spoilerText := "Luke I am your father"
	command := model.CommandArgs{
		Command:   "/spoiler " + spoilerText,
		UserId:    "userid",
		ChannelId: "channelid",
	}
	formattedSpoilerText := "**[SPOILER]**\n\n"+spoilerText+"\n**[/SPOILER]**"
	response, err := p.ExecuteCommand(&plugin.Context{}, &command)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, formattedSpoilerText, post.Message)
	assert.Equal(t, spoilerText, post.Props["CustomSpoilerRawMessage"])
	assert.Equal(t, "custom_spoiler", post.Type)
	assert.Empty(t, response.ResponseType)
}

func TestExecuteCommandErrorOnPost(t *testing.T) {
	api := &plugintest.API{}

	api.On("RegisterCommand", mock.Anything).Return(nil)

	errCreatePost := model.AppError{Message: "blablabla"}
	api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, &errCreatePost)

	p := Plugin{}
	p.SetAPI(api)

	assert.Nil(t, p.OnActivate())

	spoilerText := "Luke I am your father"
	command := model.CommandArgs{
		Command:   "/spoiler " + spoilerText,
		UserId:    "userid",
		ChannelId: "channelid",
	}

	response, err := p.ExecuteCommand(&plugin.Context{}, &command)
	assert.NotNil(t, err)
	assert.Equal(t, errCreatePost, *err)
	assert.Nil(t, response)
}
