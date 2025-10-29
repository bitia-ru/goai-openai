package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bitia-ru/goai-openai/ai/openai"
	"github.com/bitia-ru/goai/ai"
)

func main() {
	ctx := context.Background()

	token := os.Getenv("OPENAI_TOKEN")

	if token == "" {
		fmt.Println("OPENAI_TOKEN not set")
		return
	}

	c := openai.NewClient(ctx, token)

	d := c.NewDialog()

	_, err := c.Query("Calculate exression and answer with a single word (number): 4 * (3 + 10)", d)

	if err != nil {
		fmt.Printf("Error querying OpenAI: %v\n", err)
		return
	}

	messages := d.GetMessages()

	if len(messages) == 0 {
		fmt.Println("No messages returned")
		return
	}

	if messages[1].Content() != "52" {
		fmt.Printf("Unexpected response: %v (expected 52)\n", messages[1].Content())
	}

	err = d.SetModelSize(ai.ModelL)

	if err != nil {
		fmt.Printf("Error setting model size: %v", err)
		return
	}

	_, err = c.Query(
		"What is your AI model name? Reply with a single word. Specify the right version (3, 4, 4o, 4o-mini).",
		d,
	)

	if err != nil {
		fmt.Printf("Error querying OpenAI: %v\n", err)
		return
	}

	messages = d.GetMessages()

	if len(messages) != 4 {
		fmt.Printf("Unexpected number of messages: %v (expected 4)", len(messages))
		return
	}

	if messages[3].Content() != "GPT-4" {
		fmt.Printf("Unexpected response: %v (expected GPT-4)", messages[3].Content())
		return
	}

	fmt.Println("Smoke test passed")
}
