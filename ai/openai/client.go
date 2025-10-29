package openai

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bitia-ru/goai/ai"
	goOpenai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type ClientConfig struct {
	HttpClient *http.Client
	BaseURL    string
}

type client struct {
	ctx              context.Context
	aiClient         *goOpenai.Client
	maxQueryMessages int
}

func NewClientWithConfig(ctx context.Context, token string, clientConfig ClientConfig) ai.Client {
	config := goOpenai.DefaultConfig(token)

	if clientConfig.HttpClient != nil {
		config.HTTPClient = clientConfig.HttpClient
	}

	if clientConfig.BaseURL != "" {
		config.BaseURL = clientConfig.BaseURL
	}

	c := &client{
		ctx:              ctx,
		aiClient:         goOpenai.NewClientWithConfig(config),
		maxQueryMessages: 1024,
	}

	return c
}

func NewClient(ctx context.Context, token string) ai.Client {
	return NewClientWithConfig(ctx, token, ClientConfig{})
}

func (c *client) NewDialog() ai.Dialog {
	d := &Dialog{
		tools:       []ai.Tool{},
		temperature: float32(1.0),
	}

	_ = d.SetModelSize(ai.ModelS)

	return d
}

func (c *client) RequestCompletion(dialog ai.Dialog) (ai.ResponseMetadata, error) {
	metadata := ai.ResponseMetadata{}

	aiDialog, ok := dialog.(*Dialog)

	if !ok {
		return metadata, fmt.Errorf("invalid dialog type")
	}

	finished := false

	for !finished {
		startTime := time.Now()

		resp, err := c.aiClient.CreateChatCompletion(
			c.ctx,
			goOpenai.ChatCompletionRequest{
				Model:       aiDialog.GetOpenAIModelName(),
				Tools:       aiDialog.GetOpenAITools(),
				Messages:    aiDialog.messages,
				Temperature: aiDialog.temperature,
			},
		)

		elapsed := time.Since(startTime)

		if err != nil {
			return metadata, err
		}

		finished = true

		metadata.PromptTokensSpent += resp.Usage.PromptTokens
		metadata.CompletionTokensSpent += resp.Usage.CompletionTokens
		metadata.TotalTokensSpent += resp.Usage.TotalTokens
		metadata.TimeSpentMs += float32(elapsed.Milliseconds())

		aiDialog.messages = append(aiDialog.messages, resp.Choices[0].Message)

		if len(resp.Choices[0].Message.ToolCalls) > 0 {
			for _, toolCall := range resp.Choices[0].Message.ToolCalls {
				for _, tool := range aiDialog.tools {
					if toolCall.Function.Name == tool.Name {
						result, err := ExecuteTool(tool, toolCall)

						if err != nil {
							aiDialog.messages = append(aiDialog.messages, goOpenai.ChatCompletionMessage{
								Role:       goOpenai.ChatMessageRoleTool,
								Content:    "{ \"error\": \"" + err.Error() + "\" }",
								Name:       tool.Name,
								ToolCallID: toolCall.ID,
							})
						} else {
							aiDialog.messages = append(aiDialog.messages, goOpenai.ChatCompletionMessage{
								Role:       goOpenai.ChatMessageRoleTool,
								Content:    result,
								Name:       tool.Name,
								ToolCallID: toolCall.ID,
							})
						}

						finished = false

						break
					}
				}
			}
		}
	}

	return metadata, nil
}

func (c *client) Query(query string, dialog ai.Dialog) (ai.ResponseMetadata, error) {
	aiDialog, ok := dialog.(*Dialog)

	if !ok {
		return ai.ResponseMetadata{}, fmt.Errorf("invalid dialog type")
	}

	aiDialog.AppendUserMessage(query)

	return c.RequestCompletion(aiDialog)
}

func aiParameterTypeToOpenaiToolParameterType(parameterType ai.DataType) jsonschema.DataType {
	switch parameterType {
	case ai.String:
		return jsonschema.String
	case ai.Integer:
		return jsonschema.Integer
	case ai.Real:
		return jsonschema.Number
	case ai.Boolean:
		return jsonschema.Boolean
	case ai.Datetime:
		return jsonschema.String
	default:
		return jsonschema.String
	}
}
