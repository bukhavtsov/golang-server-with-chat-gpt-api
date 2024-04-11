package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bukhavtsov/dictionary-tutorial/pkg/domain"
	"github.com/labstack/echo/v4"
	"io"
	"log/slog"
	"net/http"
)

type translatorServer struct {
	logger        slog.Logger
	chatGptApiKey string
	chatGptApiUrl string
}

func NewTranslatorServer(logger slog.Logger, chatGptApiKey, chatGptApiUrl string) *translatorServer {
	return &translatorServer{
		logger:        logger,
		chatGptApiKey: chatGptApiKey,
		chatGptApiUrl: chatGptApiUrl,
	}
}

func (t translatorServer) Translate(c echo.Context) error {
	lexicalItem := c.Param("lexicalItem")
	if lexicalItem == "" {
		return c.String(http.StatusBadRequest, "lexical item wasn't provided")
	}

	translationResponse, err := t.translateByChatGPT(lexicalItem)
	if err != nil {
		t.logger.Error("Failed to translate", slog.Any("err", err))
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, translationResponse)
}

func (t translatorServer) translateByChatGPT(lexicalItem string) (*domain.TranslationResponse, error) {
	prompt := "Translate the lexical item, provide response in the following json format: lexicalItem(string), meaning (string), example (string). lexical item to translate:" + lexicalItem

	requestBody := fmt.Sprintf(`{"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "%s"}]}`, prompt)

	req, err := http.NewRequest("POST", t.chatGptApiUrl, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.chatGptApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make req to chat gpt api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var chatGPTResponse domain.ChatGPTResponse
	err = json.Unmarshal(body, &chatGPTResponse)
	if err != nil {
		return nil, err
	}
	if len(chatGPTResponse.Choices) != 1 {
		return nil, fmt.Errorf("expected only one choice, but recieved: %d", len(chatGPTResponse.Choices))
	}

	var translationResp domain.TranslationResponse
	err = json.Unmarshal([]byte(chatGPTResponse.Choices[0].Message.Content), &translationResp)
	if err != nil {
		return nil, err
	}
	return &translationResp, nil
}
