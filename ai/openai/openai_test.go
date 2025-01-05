package openai

import (
	"context"
	"github.com/bitia-ru/goai/ai"
	"os"
	"testing"
)

func TestQuery(t *testing.T) {
	ctx := context.Background()

	token := os.Getenv("OPENAI_TOKEN")

	if token == "" {
		t.Errorf("OPENAI_TOKEN not set")
		return
	}

	c := NewClient(ctx, token)

	d := c.NewDialog()

	err := c.Query("Calculate exression and answer with a single word (number): 4 * (3 + 10)", d)

	if err != nil {
		t.Errorf("Error querying OpenAI: %v", err)
		return
	}

	messages := d.GetMessages()

	if len(messages) == 0 {
		t.Errorf("No messages returned")
		return
	}

	if messages[1].Content() != "52" {
		t.Errorf("Unexpected response: %v (expected 52)", messages[1].Content())
	}

	err = d.SetModelSize(ai.ModelL)

	if err != nil {
		t.Errorf("Error setting model size: %v", err)
		return
	}

	err = c.Query("What is your AI model name? Reply with a single word. Specify the right version (3, 4, 4o, 4o-mini).", d)

	if err != nil {
		t.Errorf("Error querying OpenAI: %v", err)
		return
	}

	messages = d.GetMessages()

	if len(messages) != 4 {
		t.Errorf("Unexpected number of messages: %v (expected 4)", len(messages))
		return
	}

	if messages[3].Content() != "GPT-4" {
		t.Errorf("Unexpected response: %v (expected GPT-4)", messages[3].Content())
	}
}
