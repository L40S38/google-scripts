// Command video uploads a local video file to the Gemini Files API and asks
// the model a batch of questions about its content, to see how much detail
// Gemini can read out of a video (要約・登場人物・タイムスタンプ・音の種類)。
//
// 公式Go SDK (google.golang.org/genai) を使用。事前に `go mod tidy` が必要。
//
// Usage:
//
//	GEMINI_API_KEY=... go run ./cmd/video [path/to/video.mp4]
//
// 引数を省略した場合は gemini/README.md で取得する
// BigBuckBunny_320x180.mp4 (CC BY 3.0, Blender Foundation) を使う想定。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/genai"
	"google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
)

const defaultVideoPath = "BigBuckBunny_320x180.mp4"

const prompt = `この動画を分析し、指定されたJSON Schemaに従って日本語で出力してください。
- summary: 3文程度の要約
- characters: 登場するキャラクターと、それぞれの特徴
- key_moments: 印象的な出来事が起きる場面を3つ。timestampはおおよその分:秒形式
- sounds: 動画中に聞こえる音(BGM・効果音・鳴き声など)の種類`

// videoAnalysisSchema: 検証(同じ入力に対する出力の比較)をしやすくするため、
// 自由文ではなくJSON Schemaで出力形式を固定する。
var videoAnalysisSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"summary": map[string]any{
			"type":        "string",
			"description": "動画の3文程度の要約",
		},
		"characters": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":        map[string]any{"type": "string"},
					"description": map[string]any{"type": "string"},
				},
				"required": []string{"name", "description"},
			},
		},
		"key_moments": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"timestamp":   map[string]any{"type": "string", "description": "分:秒 形式。例: 2:15"},
					"description": map[string]any{"type": "string"},
				},
				"required": []string{"timestamp", "description"},
			},
		},
		"sounds": map[string]any{
			"type":  "array",
			"items": map[string]any{"type": "string"},
		},
	},
	"required": []string{"summary", "characters", "key_moments", "sounds"},
}

func main() {
	videoPath := defaultVideoPath
	if len(os.Args) > 1 {
		videoPath = os.Args[1]
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s をアップロード中...\n", videoPath)
	myfile, err := client.Files.UploadFromPath(ctx, videoPath, &genai.UploadFileConfig{
		MIMEType: "video/mp4",
	})
	if err != nil {
		log.Fatalf("uploading video: %v", err)
	}

	for myfile.State != genai.FileStateActive {
		fmt.Println("処理中...")
		time.Sleep(5 * time.Second)
		myfile, err = client.Files.Get(ctx, myfile.Name, nil)
		if err != nil {
			log.Fatalf("checking file status: %v", err)
		}
	}

	contents := []interactions.Content{
		interactions.NewContent(interactions.VideoContent{
			URI:      genai.Ptr(myfile.URI),
			MimeType: interactions.VideoContentMimeType(myfile.MIMEType).ToPointer(),
		}),
		interactions.NewContent(interactions.TextContent{
			Text: prompt,
		}),
	}

	input := interactions.NewInteractionsInput(contents)
	responseFormat := interactions.NewCreateModelInteractionResponseFormat(
		interactions.NewResponseFormat(interactions.TextResponseFormat{
			MimeType: interactions.TextResponseFormatMimeTypeApplicationJSON.ToPointer(),
			Schema:   videoAnalysisSchema,
		}),
	)

	res, err := client.Interactions.Create(ctx, operations.CreateInteractionRequest{
		Body: operations.NewCreateInteractionRequestBody(interactions.CreateModelInteraction{
			Model: interactions.Model("gemini-3.8-flash"),
			Input: &input,
			GenerationConfig: &interactions.GenerationConfig{
				// 検証用に固定(0.0)。ただしSDK上はDeprecated注記あり(将来の置き換えに注意)。
				Temperature: genai.Ptr(float32(0.0)),
			},
			ResponseFormat: &responseFormat,
		}),
	})
	if err != nil {
		log.Fatalf("calling Gemini API: %v", err)
	}
	if res.Interaction == nil {
		log.Fatal("empty response from Gemini API")
	}

	fmt.Printf("status: %s\n", res.Interaction.Status)
	for _, e := range res.Interaction.Errors {
		fmt.Printf("error: %s\n", stringOrEmpty(e.Message))
	}

	var outputText string
	for _, s := range res.Interaction.Steps {
		if s.Type != interactions.StepTypeModelOutput || s.ModelOutputStep == nil {
			continue
		}
		if s.ModelOutputStep.Error != nil {
			fmt.Printf("model_output error: %s\n", stringOrEmpty(s.ModelOutputStep.Error.Message))
		}
		for _, c := range s.ModelOutputStep.Content {
			if c.Type == interactions.ContentTypeText && c.TextContent != nil {
				outputText += c.TextContent.Text
			}
		}
	}

	// OutputTextはSDKが付与する便利フィールド(Stepsから計算される)。
	// Stepsから見つからなかった場合のフォールバックとして使う。
	if outputText == "" && res.Interaction.OutputText != nil {
		outputText = *res.Interaction.OutputText
	}
	if outputText == "" {
		fmt.Println("(モデルからテキスト出力がありませんでした。statusやerrorを確認してください)")
		return
	}

	// response_formatでJSONを指定しているので、見やすく整形して表示する。
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(outputText), "", "  "); err != nil {
		fmt.Println(outputText)
		return
	}
	fmt.Println(pretty.String())
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
