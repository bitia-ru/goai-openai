package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bitia-ru/goai-openai/ai/openai"
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

	d.AppendUserMessageWithImage(
		"Describe, what do you see on the image",
		"https://upload.wikimedia.org/wikipedia/commons/thumb/b/b4/Zero_People_band.jpg/2560px-Zero_People_band.jpg",
	)

	_, err := c.RequestCompletion(d)

	if err != nil {
		fmt.Printf("Error querying OpenAI: %v\n", err)
		return
	}

	message := d.GetLastMessage()

	fmt.Println(message.Content())
}
