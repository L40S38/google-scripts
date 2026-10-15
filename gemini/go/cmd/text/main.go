// Command text calls the Gemini API (Interactions API) over REST and
// prints the generated text for a fixed prompt.
//
// REST直叩きの実装(標準ライブラリのみ、外部依存なし)。公式Go SDK
// (google.golang.org/genai)を使った例は ../video を参照。
//
// Usage:
//
//	GEMINI_API_KEY=... go run ./cmd/text
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const endpoint = "https://generativelanguage.googleapis.com/v1beta/interactions"

type requestBody struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type step struct {
	Type    string    `json:"type"`
	Content []content `json:"content"`
}

type interactionResponse struct {
	Status string `json:"status"`
	Steps  []step `json:"steps"`
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY環境変数が未設定")
	}

	reqBody, err := json.Marshal(requestBody{
		Model: "gemini-3.8-flash",
		Input: "Explain how AI works in a few words",
	})
	if err != nil {
		log.Fatalf("encoding request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		log.Fatalf("building request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("calling Gemini API: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("reading response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Gemini API error (status %d): %s", resp.StatusCode, respBody)
	}

	var result interactionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Fatalf("decoding response: %v", err)
	}

	for _, s := range result.Steps {
		if s.Type != "model_output" {
			continue
		}
		for _, c := range s.Content {
			if c.Type == "text" {
				fmt.Println(c.Text)
			}
		}
	}
}
