package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestServeHTTP(t *testing.T) {
	spoilerMode := "kjqshdlkjhfk"
	spoiler := "hahahahaha"
	for name, test := range map[string]struct {
		RequestURL         string
		RequestBody        string
		ExpectedStatusCode int
		ExpectedHeader     http.Header
		ExpectedbodyString string
	}{
		"Show spoiler request": {
			RequestURL:         "/show",
			RequestBody:        `{"Context":{"spoiler":"` + spoiler + `"}}`,
			ExpectedStatusCode: http.StatusOK,
			ExpectedHeader:     http.Header{"Content-Type": []string{"application/json"}},
			ExpectedbodyString: `{"update":null,"ephemeral_text":"` + spoiler + `","skip_slack_parsing":false}`,
		},
		"Show invalid spoiler request": {
			RequestURL:         "/show",
			RequestBody:        "",
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedHeader:     http.Header{},
			ExpectedbodyString: "",
		},
		"Config request": {
			RequestURL:         "/config",
			RequestBody:        "",
			ExpectedStatusCode: http.StatusOK,
			ExpectedHeader:     http.Header{"Content-Type": []string{"application/json"}},
			ExpectedbodyString: `{"spoilerMode":"` + spoilerMode + `"}`,
		},
		"InvalidRequestURL": {
			RequestURL:         "/not_found",
			RequestBody:        "",
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedHeader:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}, "X-Content-Type-Options": []string{"nosniff"}},
			ExpectedbodyString: "404 page not found\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			plugin := &Plugin{}
			config := &Configuration{SpoilerMode: spoilerMode}
			plugin.setConfiguration(config)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", test.RequestURL, strings.NewReader(test.RequestBody))
			plugin.ServeHTTP(nil, w, r)

			result := w.Result()
			require.NotNil(t, result)

			bodyBytes, err := io.ReadAll(result.Body)
			require.Nil(t, err)
			bodyString := string(bodyBytes)

			assert.Equal(test.ExpectedStatusCode, result.StatusCode)
			assert.Equal(test.ExpectedbodyString, bodyString)
			assert.Equal(test.ExpectedHeader, result.Header)
		})
	}
}

func TestParseCommandLine(t *testing.T) {
	testCases := []struct {
		command         string
		expectedError   bool
		expectedSpoiler string
		expectedCaption string
	}{
		// True errors (not a lot, we are very lenient)
		{command: "", expectedError: true, expectedSpoiler: "", expectedCaption: ""},
		// Typos that will cause the whole message to be considered as a spoiler
		{command: "\"missing closing quote", expectedError: false, expectedSpoiler: "missing closing quote", expectedCaption: ""},
		{command: "missing opening quote\" spoiler", expectedError: false, expectedSpoiler: "missing opening quote\" spoiler", expectedCaption: ""},
		{command: "\"caption without spoiler\"", expectedError: false, expectedSpoiler: "caption without spoiler", expectedCaption: ""},
		// Fully correct commands
		{command: "spoiler without caption", expectedError: false, expectedSpoiler: "spoiler without caption", expectedCaption: ""},
		{command: "\"caption\" spoiler", expectedError: false, expectedSpoiler: "spoiler", expectedCaption: "caption"},
		{command: "\"caption\" \"spoiler\"", expectedError: false, expectedSpoiler: "spoiler", expectedCaption: "caption"},
		{command: "\"caption\" spoiler", expectedError: false, expectedSpoiler: "spoiler", expectedCaption: "caption"},
		{command: "\"caption with \\\"escaped quote\\\"!\" spoiler", expectedError: false, expectedSpoiler: "spoiler", expectedCaption: "caption with \\\"escaped quote\\\"!"},
		{command: "\"We\nlike\nnew\nlines\" yes\nwe\ndo", expectedError: false, expectedSpoiler: "yes\nwe\ndo", expectedCaption: "We\nlike\nnew\nlines"},
		{command: "\"Unicode supporté? 👍\" héhéhé 👍!", expectedError: false, expectedSpoiler: "héhéhé 👍!", expectedCaption: "Unicode supporté? 👍"},
		{command: "\"Caption inside (link)[https://test.com]?\" Spoiler inside (link)[https://test.com]?", expectedError: false, expectedSpoiler: "Spoiler inside (link)[https://test.com]?", expectedCaption: "Caption inside (link)[https://test.com]?"},
	}
	for _, testCase := range testCases {
		keywords, caption, err := parseCommandLine("/spoiler " + testCase.command)

		if testCase.expectedError {
			assert.NotNil(t, err, "Testing: "+testCase.command)
		} else {
			assert.Nil(t, err, "Testing: "+testCase.command)
		}
		assert.Equal(t, testCase.expectedSpoiler, keywords, "Testing: "+testCase.command)
		assert.Equal(t, testCase.expectedCaption, caption, "Testing: "+testCase.command)
	}
}
