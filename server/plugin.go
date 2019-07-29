package main

import (
	"fmt"
	"io"
	"net/http"
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
	customPostProp = "CustomSpoilerRawMessage"
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

// ServeHTTP serve the post action to display an ephemeral spoiler
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/show":
		p.showEphemeral(w, r)
	case "/config":
		p.handleConfigRequest(w, r)
	default:
		http.NotFound(w, r)
	}
}

// ExecuteCommand post a custom-type spoiler post, the webapp part of the plugin will display it right
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	rawText := strings.TrimSpace((strings.Replace(args.Command, "/"+trigger, "", 1)))

	// A slash command can not return a post with a custom type
	// so the spoiler post is created manually and the command
	// response is to do nothing
	_, err := p.API.CreatePost(p.getSpoilerPost(args.UserId, args.ChannelId, args.RootId, rawText))
	if err != nil {
		return nil, err
	}

	return &model.CommandResponse{}, nil
}

func (p *Plugin) getSpoilerPost(userId, channelId, rootId, spoiler string) *model.Post {
	return &model.Post{
		UserId:    userId,
		ChannelId: channelId,
		Type:      customPostType,
		RootId:    rootId,
		// The webapp plugin will use the RawMessage for the custom display
		Props: map[string]interface{}{
			customPostProp: spoiler,
			"attachments":  p.getPostAttachments(*p.API.GetConfig().ServiceSettings.SiteURL, p.getConfiguration().IntegrationURL, manifest.Id, spoiler),
		},
	}
}

func (p *Plugin) getPostAttachments(siteURL, integrationURL, pluginID, spoilerText string) []*model.SlackAttachment {
	baseURL := strings.TrimSuffix(integrationURL, "/")
	if baseURL == "" {
		baseURL = siteURL
	}
	actions := []*model.PostAction{{
		Name: "Show spoiler",
		Type: model.POST_ACTION_TYPE_BUTTON,
		Integration: &model.PostActionIntegration{
			URL:     fmt.Sprintf("%s/plugins/%s/show", siteURL, pluginID),
			Context: model.StringInterface{"spoiler": spoilerText},
		},
	},
	}

	return []*model.SlackAttachment{{
		Actions: actions,
	}}
}

// Show spoiler content as an ephemeral message
func (p *Plugin) showEphemeral(w http.ResponseWriter, r *http.Request) {
	request := model.PostActionIntegrationRequesteFromJson(r.Body)
	if request == nil || request.Context == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := &model.PostActionIntegrationResponse{
		EphemeralText: request.Context["spoiler"].(string),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, response.ToJson())
}
