package main

import (
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

// Plugin that adds a slash command to hide spoilers
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *Configuration
}

const (
	trigger        = "spoiler"
	customPostType = "custom_spoiler"
)

// OnActivate register the plugin command
func (p *Plugin) OnActivate() error {
	return p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Hide a message that contains a spoiler",
		DisplayName:      "Spoiler",
		AutoComplete:     true,
		AutoCompleteDesc: "Hide a spoiler message",
		AutoCompleteHint: "The Titanic sinks at the end.",
	})
}

// ExecuteCommand post a custom-type spoiler post, the webapp part of the plugin will display it right
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	rawText := strings.TrimSpace((strings.Replace(args.Command, "/"+trigger, "", 1)))

	// Native apps (Android, Apple...) do not support plugins yet,
	// so by default, the spoiler is displayed surrounded by spoiler tags
	text := "**[SPOILER]**\n\n" + rawText + "\n**[/SPOILER]**"

	// A slash command can not return a post with a custom type
	// so the spoiler post is created manually and the command
	// response is to do nothing
	_, err := p.API.CreatePost(&model.Post{
		UserId:    args.UserId,
		ChannelId: args.ChannelId,
		Message:   text,
		Type:      customPostType,
		// The webapp plugin will use the RawMessage and display it blurred (without tags)
		Props: map[string]interface{}{
			"CustomSpoilerRawMessage": rawText,
		},
	})
	if err != nil {
		return nil, err
	}

	return &model.CommandResponse{}, nil
}
