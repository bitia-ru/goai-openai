package openai

import (
	"context"
	"encoding/json"
	"github.com/bitia-ru/goai/ai"
	goOpenai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type client struct {
	ctx              context.Context
	aiClient         *goOpenai.Client
	aiTools          []goOpenai.Tool
	tools            []ai.Tool
	maxQueryMessages int
}

func NewClient(ctx context.Context, token string) ai.Client {
	return &client{
		ctx:              ctx,
		aiClient:         goOpenai.NewClient(token),
		aiTools:          []goOpenai.Tool{},
		tools:            []ai.Tool{},
		maxQueryMessages: 15,
	}
}

func (c *client) NewDialog() ai.Dialog {
	return &Dialog{}
}

func (c *client) Query(query string, dialog ai.Dialog) error {
	aiDialog := dialog.(*Dialog)

	aiDialog.messages = append(aiDialog.messages, goOpenai.ChatCompletionMessage{
		Role:    goOpenai.ChatMessageRoleUser,
		Content: query,
	})

	finished := false

	for !finished {
		resp, err := c.aiClient.CreateChatCompletion(
			c.ctx,
			goOpenai.ChatCompletionRequest{
				Model:    goOpenai.GPT4o,
				Tools:    c.aiTools,
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
				for _, tool := range c.tools {
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
		} else {

		}
	}

	return nil
}

func (c *client) SetTools(tools []ai.Tool) {
	c.tools = tools

	for _, tool := range tools {
		parameters := jsonschema.Definition{
			Type:       jsonschema.Object,
			Properties: make(map[string]jsonschema.Definition),
			Required:   []string{},
			Items:      nil,
		}

		for _, parameter := range tool.Parameters {
			parameters.Properties[parameter.Name] = jsonschema.Definition{
				Type:        aiParameterTypeToOpenaiToolParameterType(parameter.Type),
				Description: parameter.Description,
			}
		}

		c.aiTools = append(c.aiTools, goOpenai.Tool{
			Type: goOpenai.ToolTypeFunction,
			Function: &goOpenai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Strict:      false,
				Parameters:  parameters,
			},
		})
	}
}

func ExecuteTool(tool ai.Tool, toolCall goOpenai.ToolCall) (string, error) {
	parameters := make(map[string]interface{})
	err := json.Unmarshal([]byte(toolCall.Function.Arguments), &parameters)

	if err != nil {
		return "", err
	}

	return tool.Function(parameters)
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
