package controlplan

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	messages      []openai.ChatCompletionMessageParamUnion
	client        openai.Client
	acceptMessage string
)

func PlannerInit() error {
	// apiKey := os.Getenv("OPENAI_API_KEY")

	client = openai.NewClient(
		option.WithBaseURL("http://localhost:11434/v1"),
		option.WithAPIKey("ollama"),
	)

	data, err := os.ReadFile("prompts/planner.md")
	if err != nil {
		return err
	}

	messages = []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(string(data)),
	}

	data, err = os.ReadFile("prompts/formatter.md")
	if err != nil {
		return err
	}

	acceptMessage = string(data)

	return nil
}

func Planner() {
	reader := bufio.NewReader(os.Stdin)
	isComplete := false
	for {
		fmt.Print("\nYou: ")
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)

		if userInput == "exit" {
			break
		}

		if strings.ToLower(userInput) == "accept" {
			last := messages[len(messages)-1]
			assistantMsg := last.OfAssistant

			content := assistantMsg.Content.OfString.Value
			messages = []openai.ChatCompletionMessageParamUnion{}
			messages = append(messages, openai.SystemMessage(acceptMessage))
			messages = append(messages, openai.UserMessage(content))
			isComplete = true
		} else {
			messages = append(messages,
				openai.UserMessage(userInput),
			)
		}

		resp, err := client.Chat.Completions.New(
			context.Background(),
			openai.ChatCompletionNewParams{
				Model:    "qwen3:1.7b",
				Messages: messages,
			},
		)
		if err != nil {
			panic(err)
		}

		assistantReply := resp.Choices[0].Message.Content

		fmt.Println("\nAssistant:", assistantReply)

		if isComplete {
			// write the yaml output to the test.yaml file
			err = writeConfig(assistantReply)
			if err != nil {
				panic(err)
			}
			return
		}

		messages = append(messages,
			openai.AssistantMessage(assistantReply),
		)
	}
}

func writeConfig(yamlText string) error {
	err := os.WriteFile("test.yaml", []byte(yamlText), 0644)
	if err != nil {
		return err
	}
	return nil
}
