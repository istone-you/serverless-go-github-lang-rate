package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Repository struct {
	Name  string `json:"name"`
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
}

type Language struct {
	LanguageName string `json:"name"`
	Bytes        int    `json:"bytes"`
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	username := os.Getenv("GITHUB_USERNAME")
	// apiToken := os.Getenv("GITHUB_TOKEN")

	url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	// req.Header.Add("Authorization", "token "+apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	defer resp.Body.Close()

	var repositories []Repository
	err = json.NewDecoder(resp.Body).Decode(&repositories)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	languagesToExclude := map[string]bool{
		"HTML":       true,
		"CSS":        true,
		"SCSS":       true,
		"Dockerfile": true,
		"Shell":      true,
		"Makefile":   true,
		"Ruby":       true,
		"Jinja":      true,
		"Smarty":     true,
		// Add other languages you want to exclude here
	}

	languageStats := make(map[string]int)

	for _, repo := range repositories {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/languages", repo.Owner.Login, repo.Name)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		// req.Header.Add("Authorization", "token "+apiToken)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var languages map[string]int
		err = json.NewDecoder(resp.Body).Decode(&languages)
		if err != nil {
			continue
		}

		for lang, bytes := range languages {
			if !languagesToExclude[lang] {
				languageStats[lang] += bytes
			}
		}
	}

	totalBytes := 0
	for _, bytes := range languageStats {
		totalBytes += bytes
	}

	languageUsage := make(map[string]float64)
	for lang, bytes := range languageStats {
		percentage := float64(bytes) / float64(totalBytes) * 100
		languageUsage[lang] = percentage
	}

	outputJSON, err := json.Marshal(languageUsage)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(outputJSON),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
