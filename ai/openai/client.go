package openai

import (
	"context"
	"fmt"
	"github.com/bitia-ru/goai/ai"
	goOpenai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"net/http"
)

type ClientConfig struct {
	HttpClient *http.Client
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
		tools: []ai.Tool{},
	}

	_ = d.SetModelSize(ai.ModelS)

	return d
}

func (c *client) RequestCompletion(dialog ai.Dialog) error {
	aiDialog, ok := dialog.(*Dialog)

	if !ok {
		return fmt.Errorf("invalid dialog type")
	}

	finished := false

	for !finished {
		resp, err := c.aiClient.CreateChatCompletion(
			c.ctx,
			goOpenai.ChatCompletionRequest{
				Model:    aiDialog.GetOpenAIModelName(),
				Tools:    aiDialog.GetOpenAITools(),
				Messages: aiDialog.messages,
			},
		)

		if err != nil {
			return err
		}

		finished = true

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

	return nil
}

func (c *client) Query(query string, dialog ai.Dialog) error {
	aiDialog, ok := dialog.(*Dialog)

	if !ok {
		return fmt.Errorf("invalid dialog type")
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
