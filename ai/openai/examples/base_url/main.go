package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bitia-ru/goai-openai/ai/openai"
)

func main() {
	ctx := context.Background()

	token := os.Getenv("OPENAI_TOKEN")

	if token == "" {
		fmt.Println("OPENAI_TOKEN not set")
		return
	}

	c := openai.NewClientWithConfig(ctx, token, openai.ClientConfig{
		BaseURL: "http://localhost:11434/v1",
	})

	d := c.NewDialog()

	err := d.SetModelName("mistral:7b")

	if err != nil {
		fmt.Printf("Error setting model size: %v", err)
		return
	}

	err = d.SetTemperature(0.2)

	if err != nil {
		fmt.Printf("Error setting temperature: %v", err)
		return
	}

	d.AppendSystemMessage("You need to calculate a mathematical expression.")
	d.AppendSystemMessage("Answer format: number without any additional symbol. Example:42")

	_, err = c.Query("4 * (3 + 10)", d)

	if err != nil {
		fmt.Printf("Error querying OpenAI: %v\n", err)
		return
	}

	answer := d.GetLastMessage()

	if strings.TrimSpace(answer.Content()) != "52" {
		fmt.Printf("Unexpected response: %v (expected 52)\n", answer.Content())
		return
	}

	fmt.Println("Smoke test passed")
}
