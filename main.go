package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	_ = godotenv.Load()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is required (set it in .env or the environment)")
	}

	summary, err := SummarizeWebsite("https://www.aljazeera.com/")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("--- summary ---")
	fmt.Println(summary)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	prompt := `You are a marketing copywriter. Using the company facts below, write the full text for a single-page A4 company brochure in Markdown.

Requirements:
- Fit a one-sided A4 layout: headline, short intro, and a few scannable sections
- Cover who the company is, what they do, who they serve, and what makes them different
- Use the Description to set tone and context; keep claims grounded in the facts provided
- Professional, clear, persuasive — no fluff, no inventing numbers or awards not in the facts
- If a field is Unknown, do not invent a value; work around it
- Output valid Markdown only (# / ## headings, paragraphs, lists as needed) — no design instructions, no wrapping code fences

Company facts:
` + summary

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("--- brochure ---")
	fmt.Println(result.Text())
}
