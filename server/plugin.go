package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
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
	trigger                   = "spoiler"
	contextPropSpoiler        = "spoiler"
	contextPropDescription    = "description"
	customPostType            = "custom_spoiler"
	customPostPropSpoiler     = "CustomSpoilerRawMessage"
	customPostPropDescription = "CustomSpoilerDescription"
)

// OnActivate register the plugin command
func (p *Plugin) OnActivate() error {
	return p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Hide a message that contains a spoiler",
		DisplayName:      "Spoiler",
		AutoComplete:     true,
		AutoCompleteDesc: "Hide a spoiler message",
		AutoCompleteHint: getHintMessage(trigger),
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
	description, spoiler, parseErr := parseCommandLine(args.Command, trigger)
	if parseErr != nil {
		return nil, appError("command parse error", parseErr)
	}
	// A slash command can not return a post with a custom type
	// so the spoiler post is created manually and the command
	// response is to do nothing
	_, err := p.API.CreatePost(p.getSpoilerPost(args.UserId, args.ChannelId, args.RootId, spoiler, description))
	if err != nil {
		return nil, err
	}

	return &model.CommandResponse{}, nil
}

func parseCommandLine(commandLine, trigger string) (description, spoiler string, err error) {
	reg := regexp.MustCompile("^\\s*(?P<description>(\"([^\\s\"]+\\s*)+\")+\\s+)?(?P<spoiler>(\"(\\s*[^\\s\"]+\\s*)+\")|([^\\s\"]+\\s*)+)\\s*$")
	matchIndexes := reg.FindStringSubmatch(strings.Replace(commandLine, "/"+trigger, "", 1))
	if matchIndexes == nil {
		return "", "", fmt.Errorf("could not read the command, try one of the following syntax: /%s %s", trigger, getHintMessage(trigger))
	}
	results := make(map[string]string)
	for i, name := range reg.SubexpNames() {
		results[name] = matchIndexes[i]
	}
	return strings.Trim(strings.TrimSpace(results["description"]), "\""), strings.Trim(strings.TrimSpace(results["spoiler"]), "\""), nil
}

func getHintMessage(trigger string) string {
	return "The Titanic sinks at the end. or /" + trigger + " \"You won't believe it but in this movie\" \"The Titanic sinks at the end.\""
}

func (p *Plugin) getSpoilerPost(userID, channelID, rootID, spoiler, description string) *model.Post {
	return &model.Post{
		UserId:    userID,
		ChannelId: channelID,
		Type:      customPostType,
		RootId:    rootID,
		// The webapp plugin will use the RawMessage for the custom display
		Props: map[string]interface{}{
			customPostPropSpoiler:     spoiler,
			customPostPropDescription: description,
			"attachments":             p.getPostAttachments(spoiler, description),
		},
	}
}

func (p *Plugin) getPostAttachments(spoiler, description string) []*model.SlackAttachment {
	actions := []*model.PostAction{{
		Name: "Show spoiler",
		Type: model.POST_ACTION_TYPE_BUTTON,
		Integration: &model.PostActionIntegration{
			URL:     fmt.Sprintf("/plugins/%s/show", manifest.Id),
			Context: model.StringInterface{contextPropSpoiler: spoiler, contextPropDescription: description},
		},
	},
	}

	return []*model.SlackAttachment{{
		Text:    description,
		Actions: actions,
	}}
}

// Show spoiler content as an ephemeral message
func (p *Plugin) showEphemeral(w http.ResponseWriter, r *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil || request.Context == nil {
		p.API.LogWarn("Could not parse context from request: ")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	spoiler, ok := p.parseRequestValue(w, request, contextPropSpoiler)
	if !ok {
		return
	}
	description, ok := p.parseRequestValue(w, request, contextPropDescription)
	if !ok {
		return
	}
	message := spoiler
	if description != "" {
		message = description + "\n" + spoiler
	}
	response := &model.PostActionIntegrationResponse{
		EphemeralText: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response.ToJson())
}

func (p *Plugin) parseRequestValue(w http.ResponseWriter, request *model.PostActionIntegrationRequest, valueKey string) (string, bool) {
	valueObj, ok := request.Context[valueKey]
	if !ok {
		p.notifyHandlerError("Missing "+valueKey+" from action request context", request)
		w.WriteHeader(http.StatusBadRequest)
		return "", false
	}
	valueStr, ok := valueObj.(string)
	if !ok {
		p.notifyHandlerError("Value of "+valueKey+" should be a String", request)
		w.WriteHeader(http.StatusBadRequest)
		return "", false
	}

	return valueStr, true
}

// Informs the user of an error that occurred in a button handler (no direct response possible so it uses ephemeral messages), and also logs it
func (p *Plugin) notifyHandlerError(message string, request *model.PostActionIntegrationRequest) {
	p.API.SendEphemeralPost(request.UserId, &model.Post{
		Message:   fmt.Sprintf("*%s: %s*", manifest.Name, message),
		ChannelId: request.ChannelId,
		Props: map[string]interface{}{
			"sent_by_plugin": true,
		},
	})
	p.API.LogWarn(message)
}

// appError generates a normalized error for this plugin
func appError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError(manifest.Name, message, nil, errorMessage, http.StatusBadRequest)
}
