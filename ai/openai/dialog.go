package openai

import (
	"encoding/json"
	"github.com/bitia-ru/goai/ai"
	goOpenai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type Dialog struct {
	modelType string
	tools     []ai.Tool
	messages  []goOpenai.ChatCompletionMessage
}

type Message struct {
	*goOpenai.ChatCompletionMessage
}

func (d *Dialog) AppendUserMessage(message string) {
	d.messages = append(d.messages, goOpenai.ChatCompletionMessage{
		Role:    goOpenai.ChatMessageRoleUser,
		Content: message,
	})
}

func (d *Dialog) AppendSystemMessage(message string) {
	d.messages = append(d.messages, goOpenai.ChatCompletionMessage{
		Role:    goOpenai.ChatMessageRoleSystem,
		Content: message,
	})
}

func (d *Dialog) GetMessages() []ai.Message {
	var messages []ai.Message
	for _, message := range d.messages {
		if len(message.ToolCalls) > 0 {
			continue
		}

		messages = append(messages, Message{&message})
	}
	return messages
}

func (d *Dialog) GetLastMessage() ai.Message {
	if len(d.messages) == 0 {
		return nil
	}

	return Message{&d.messages[len(d.messages)-1]}
}

func (m Message) Content() string {
	return m.ChatCompletionMessage.Content
}

func (m Message) Type() ai.MessageType {
	switch m.ChatCompletionMessage.Role {
	case goOpenai.ChatMessageRoleUser:
		return ai.MessageTypeUser
	case goOpenai.ChatMessageRoleAssistant:
		return ai.MessageTypeAi
	case goOpenai.ChatMessageRoleTool:
		return ai.MessageTypeTool
	}

	return ai.MessageTypeUndefined
}

func (d *Dialog) SetModelSize(size ai.ModelSize) error {
	switch size {
	case ai.ModelS:
		d.modelType = goOpenai.GPT4oMini
	default:
		d.modelType = goOpenai.GPT4o
	}

	return nil
}

func (d *Dialog) GetOpenAIModelName() string {
	return d.modelType
}

func (d *Dialog) SetTools(tools []ai.Tool) error {
	// TODO: check tools content
	d.tools = tools

	return nil
}

func (d *Dialog) Duplicate() ai.Dialog {
	res := Dialog{
		modelType: d.modelType,
		tools:     d.tools,
	}

	copy(res.messages, d.messages)

	return &res
}

func (d *Dialog) GetOpenAITools() []goOpenai.Tool {
	var aiTools []goOpenai.Tool

	for _, tool := range d.tools {
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

		aiTools = append(aiTools, goOpenai.Tool{
			Type: goOpenai.ToolTypeFunction,
			Function: &goOpenai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Strict:      false,
				Parameters:  parameters,
			},
		})
	}

	return aiTools
}

func ExecuteTool(tool ai.Tool, toolCall goOpenai.ToolCall) (string, error) {
	parameters := make(map[string]interface{})
	err := json.Unmarshal([]byte(toolCall.Function.Arguments), &parameters)

	if err != nil {
		return "", err
	}

	return tool.Function(parameters)
}
