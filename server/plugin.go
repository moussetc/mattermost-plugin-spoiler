package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	manifest "github.com/moussetc/mattermost-plugin-spoiler"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
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
		AutoCompleteDesc: "Hide a spoiler message.\nCareful, file attachments will not be spoilered, use inline images instead!",
		AutoCompleteHint: getHintMessage(),
	})
}

// ServeHTTP serve the post action to display an ephemeral spoiler
func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
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
func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	spoiler, caption, parseErr := parseCommandLine(args.Command)
	if parseErr != nil {
		return nil, parseErr
	}
	// A slash command can not return a post with a custom type
	// so the spoiler post is created manually and the command
	// response is to do nothing
	_, err := p.API.CreatePost(p.getSpoilerPost(args.UserId, args.ChannelId, args.RootId, spoiler, caption))
	if err != nil {
		return nil, err
	}

	return &model.CommandResponse{}, nil
}

func parseCommandLine(commandLine string) (spoiler, caption string, err *model.AppError) {
	reg := regexp.MustCompile(`(?sm)^(\s*\"(?P<caption>(\s*(\\\"|[^\s\"])+\s*)+)\")?\s*\"?(?P<spoiler>.+)\"?$`)
	matchIndexes := reg.FindStringSubmatch(strings.TrimSpace(strings.Replace(commandLine, "/spoiler", "", 1)))
	if matchIndexes == nil {
		return "", "", &model.AppError{Message: fmt.Sprintf("could not read the command, try one of the following syntax: /spoiler %s", getHintMessage())}
	}
	results := make(map[string]string)
	for i, name := range reg.SubexpNames() {
		results[name] = matchIndexes[i]
	}
	return strings.Trim(strings.TrimSpace(results["spoiler"]), "\""), strings.Trim(strings.TrimSpace(results["caption"]), "\""), nil
}

func getHintMessage() string {
	return "my spoiler content, or /spoiler \"Visible spoiler caption\" Hidden spoiler content..."
}

func (p *Plugin) getSpoilerPost(userID, channelID, rootID, spoiler, caption string) *model.Post {
	return &model.Post{
		UserId:    userID,
		ChannelId: channelID,
		Type:      customPostType,
		RootId:    rootID,
		Message:   caption,
		// The webapp plugin will use the RawMessage for the custom display
		Props: map[string]interface{}{
			customPostProp: spoiler,
			"attachments":  p.getPostAttachments(spoiler),
		},
	}
}

func (p *Plugin) getPostAttachments(spoilerText string) []*model.SlackAttachment {
	actions := []*model.PostAction{{
		Name: "Show spoiler",
		Type: model.PostActionTypeButton,
		Integration: &model.PostActionIntegration{
			URL:     fmt.Sprintf("/plugins/%s/show", manifest.Manifest.Id),
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
	body, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var request model.PostActionIntegrationRequest
	jsonErr := json.Unmarshal(body, &request)
	if jsonErr != nil || request.Context == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := &model.PostActionIntegrationResponse{
		EphemeralText: request.Context["spoiler"].(string),
	}
	w.Header().Set("Content-Type", "application/json")
	json, jsonErr := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	if jsonErr == nil {
		_, _ = w.Write(json)
	}
}
