package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func generateCommit(questionText string) string {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(questionText))
	if err != nil {
		log.Fatal(err)
	}

	// printResponse(resp)
	
	return string(resp.Candidates[0].Content.Parts[0].(genai.Text)) 
	
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
				fmt.Printf("Type %T \n", part)
			}
		}
	}
	fmt.Println("---")
}
