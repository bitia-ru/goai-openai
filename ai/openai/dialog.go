package openai

import (
	"github.com/bitia-ru/goai/ai"
	goOpenai "github.com/sashabaranov/go-openai"
)

type Dialog struct {
	messages []goOpenai.ChatCompletionMessage
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
