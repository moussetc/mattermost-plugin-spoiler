package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommand(t *testing.T) {
	api := &plugintest.API{}

	manifest = &model.Manifest{
		Id: "testId",
	}

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

func TestExecuteCommandParseError(t *testing.T) {
	p := Plugin{}

	command := model.CommandArgs{
		Command:   "/spoiler \"\"\"",
		UserId:    "userid",
		ChannelId: "channelid",
	}
	response, err := p.ExecuteCommand(&plugin.Context{}, &command)
	assert.NotNil(t, err)
	assert.Nil(t, response)
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

func TestServeHTTP(t *testing.T) {
	spoilerMode := "kjqshdlkjhfk"
	spoiler := "hahahahaha"
	description := "hohohoho"
	for name, test := range map[string]struct {
		RequestURL         string
		RequestBody        string
		ExpectedStatusCode int
		ExpectedHeader     http.Header
		ExpectedbodyString string
		ShouldLogError     bool
		ShouldNotifyError  bool
	}{
		"Show spoiler request with description": {
			RequestURL:         "/show",
			RequestBody:        `{"Context":{"spoiler":"` + spoiler + `","description":"` + description + `"}}`,
			ExpectedStatusCode: http.StatusOK,
			ExpectedHeader:     http.Header{"Content-Type": []string{"application/json"}},
			ExpectedbodyString: `{"update":null,"ephemeral_text":"` + description + "\\n" + spoiler + `","skip_slack_parsing":false}`,
			ShouldLogError:     false,
			ShouldNotifyError:  false,
		},
		"Show spoiler request without description": {
			RequestURL:         "/show",
			RequestBody:        `{"Context":{"spoiler":"` + spoiler + `","description":""}}`,
			ExpectedStatusCode: http.StatusOK,
			ExpectedHeader:     http.Header{"Content-Type": []string{"application/json"}},
			ExpectedbodyString: `{"update":null,"ephemeral_text":"` + spoiler + `","skip_slack_parsing":false}`,
			ShouldLogError:     false,
			ShouldNotifyError:  false,
		},
		"Show invalid spoiler request (missing context)": {
			RequestURL:         "/show",
			RequestBody:        "",
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedHeader:     http.Header{},
			ExpectedbodyString: "",
			ShouldLogError:     true,
			ShouldNotifyError:  false,
		},
		"Show invalid spoiler request (missing context property)": {
			RequestURL:         "/show",
			RequestBody:        `{"Context":{"description":""}}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedHeader:     http.Header{},
			ExpectedbodyString: "",
			ShouldLogError:     true,
			ShouldNotifyError:  true,
		},
		"Show invalid spoiler request (bad context property value)": {
			RequestURL:         "/show",
			RequestBody:        `{"Context":{"spoiler":"` + spoiler + `","description":null}}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedHeader:     http.Header{},
			ExpectedbodyString: "",
			ShouldLogError:     true,
			ShouldNotifyError:  true,
		},
		"Config request": {
			RequestURL:         "/config",
			RequestBody:        "",
			ExpectedStatusCode: http.StatusOK,
			ExpectedHeader:     http.Header{"Content-Type": []string{"application/json"}},
			ExpectedbodyString: `{"spoilerMode":"` + spoilerMode + `"}`,
			ShouldLogError:     false,
			ShouldNotifyError:  false,
		},
		"InvalidRequestURL": {
			RequestURL:         "/not_found",
			RequestBody:        "",
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedHeader:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}, "X-Content-Type-Options": []string{"nosniff"}},
			ExpectedbodyString: "404 page not found\n",
			ShouldLogError:     false,
			ShouldNotifyError:  false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			api := &plugintest.API{}
			api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Return(nil)
			api.On("LogWarn", mock.AnythingOfType("string")).Return(nil)
			api.On("LogWarn", mock.AnythingOfType("string"), mock.AnythingOfType("*model.AppError")).Return(nil)

			plugin := &Plugin{}
			plugin.SetAPI(api)
			config := &Configuration{SpoilerMode: spoilerMode}
			plugin.setConfiguration(config)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", test.RequestURL, strings.NewReader(test.RequestBody))
			plugin.ServeHTTP(nil, w, r)

			result := w.Result()
			require.NotNil(t, result)

			bodyBytes, err := ioutil.ReadAll(result.Body)
			require.Nil(t, err)
			bodyString := string(bodyBytes)

			assert.Equal(test.ExpectedStatusCode, result.StatusCode)
			assert.Equal(test.ExpectedbodyString, bodyString)
			assert.Equal(test.ExpectedHeader, result.Header)
			if test.ShouldNotifyError {
				api.AssertCalled(t, "SendEphemeralPost", mock.Anything, mock.Anything)
			}
			if test.ShouldLogError {
				api.AssertCalled(t, "LogWarn", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestParseCommandeLine(t *testing.T) {
	testCases := []struct {
		command             string
		expectedError       bool
		expectedSpoiler     string
		expectedDescription string
	}{
		{command: "", expectedError: true, expectedSpoiler: "", expectedDescription: ""},
		{command: "\"k1 k2 k3", expectedError: true, expectedSpoiler: "", expectedDescription: ""},
		{command: "k1 k2 k3\"", expectedError: true, expectedSpoiler: "", expectedDescription: ""},
		{command: "\"k1 k2 k3\" m1 m2 m3", expectedError: false, expectedSpoiler: "m1 m2 m3", expectedDescription: "k1 k2 k3"},
		{command: "\"k1 k2 k3\" \"m1 m2 m3", expectedError: true, expectedSpoiler: "", expectedDescription: ""},
		{command: "\"k1 k2 k3\" m1 m2 m3\"", expectedError: true, expectedSpoiler: "", expectedDescription: ""},
		{command: "\"\" \"m1 m2 m3\"", expectedError: true, expectedSpoiler: "", expectedDescription: ""},
		{command: "unique", expectedError: false, expectedSpoiler: "unique", expectedDescription: ""},
		{command: "k1 k2", expectedError: false, expectedSpoiler: "k1 k2", expectedDescription: ""},
		{command: "\"k1 k2 k3\"", expectedError: false, expectedSpoiler: "k1 k2 k3", expectedDescription: ""},
		{command: "unique \"m1 m2 m3\"", expectedError: true, expectedSpoiler: "", expectedDescription: ""},
		{command: "\"k1 k2 k3\" \"m1 m2 m3\"", expectedError: false, expectedSpoiler: "m1 m2 m3", expectedDescription: "k1 k2 k3"},
		{command: "\"We\nlike\nnew\nlines\" \"yes\nwe\ndo\"", expectedError: false, expectedSpoiler: "yes\nwe\ndo", expectedDescription: "We\nlike\nnew\nlines"},
		{command: "\"Unicode supporté\\? ça c'est fort\" \"héhéhé !\"", expectedError: false, expectedSpoiler: "héhéhé !", expectedDescription: "Unicode supporté\\? ça c'est fort"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.command, func(t *testing.T) {
			description, spoiler, err := parseCommandLine(testCase.command, trigger)
			if testCase.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, testCase.expectedSpoiler, spoiler)
			assert.Equal(t, testCase.expectedDescription, description)
		})
	}
}
