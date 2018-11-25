package main

import (
	"github.com/mattermost/mattermost-server/model"

	"testing"

	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	api := &plugintest.API{}

	var post *model.Post
	api.On("RegisterCommand", mock.Anything).Return(nil)
	mockURL := "http://localhost"
	api.On("GetConfig", mock.Anything).Return(&model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: &mockURL,
		},
	})
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
	response, err := p.ExecuteCommand(&plugin.Context{}, &command)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "", post.Message)
	assert.Equal(t, spoilerText, post.Props["CustomSpoilerRawMessage"])
	assert.NotNil(t, post.Props)
	attachments := post.Props["attachments"].([]*model.SlackAttachment)
	assert.NotNil(t, attachments)
	assert.Len(t, attachments, 1)
	actions := attachments[0].Actions
	assert.NotNil(t, actions)
	assert.Len(t, actions, 1)
	assert.Equal(t, "Show spoiler", actions[0].Name)
	assert.NotNil(t, actions[0].Integration)
	assert.NotNil(t, actions[0].Integration.Context)
	assert.Equal(t, spoilerText, actions[0].Integration.Context["spoiler"])
	assert.Equal(t, "custom_spoiler", post.Type)
	assert.Empty(t, response.ResponseType)
}

func TestExecuteCommandErrorOnPost(t *testing.T) {
	api := &plugintest.API{}

	api.On("RegisterCommand", mock.Anything).Return(nil)
	mockURL := "http://localhost"
	api.On("GetConfig", mock.Anything).Return(&model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: &mockURL,
		},
	})

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
