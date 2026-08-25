package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	openRouterKeyURL = "https://openrouter.ai/api/v1/key"
	openRouterModel  = "google/gemini-2.5-flash-lite"
)

func validateOpenRouterKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("%s", currentMessages().errOpenRouterKeyEmpty)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openRouterKeyURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("User-Agent", "Quick-Anydesk-Connect/1.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return errOpenRouterTimeout
		}
		return fmt.Errorf("%w: %v", errOpenRouterNetwork, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return errOpenRouterBadRequest
	case http.StatusNotFound:
		return errOpenRouterModelUnavailable
	case http.StatusUnauthorized:
		return errOpenRouterUnauthorized
	case http.StatusForbidden:
		return errOpenRouterForbidden
	case http.StatusTooManyRequests:
		return errOpenRouterRateLimit
	default:
		if resp.StatusCode >= 500 {
			return errOpenRouterServer
		}
		return fmt.Errorf("OpenRouter HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

func configureOpenRouter() bool {
	key, err := askOpenRouterKey()
	if err != nil {
		return false
	}
	if err := validateOpenRouterKey(key); err != nil {
		showOpenRouterError(err)
		return false
	}
	if err := saveOpenRouterAPIKey(key); err != nil {
		showError(err.Error())
		return false
	}
	messageBox(mainWindow, currentMessages().openRouterSaved, "Quick Anydesk Connect", MB_OK|MB_TOPMOST|MB_SETFOREGROUND)
	return true
}

func ensureOpenRouterReady() bool {
	key, err := loadOpenRouterAPIKey()
	if err != nil {
		showError(err.Error())
		return false
	}
	if key != "" {
		if err := validateOpenRouterKey(key); err == nil {
			return true
		}
	}
	return configureOpenRouter()
}

func showOpenRouterError(err error) {
	m := currentMessages()
	text := err.Error()
	switch err {
	case errOpenRouterUnauthorized:
		text = m.errOpenRouterUnauthorized
	case errOpenRouterForbidden:
		text = m.errOpenRouterForbidden
	case errOpenRouterRateLimit:
		text = m.errOpenRouterRateLimit
	case errOpenRouterServer:
		text = m.errOpenRouterServer
	case errOpenRouterTimeout:
		text = m.errOpenRouterTimeout
	case errOpenRouterBadRequest:
		text = m.errOpenRouterBadRequest
	case errOpenRouterModelUnavailable:
		text = m.errOpenRouterModelUnavailable
	default:
		if strings.Contains(err.Error(), errOpenRouterNetwork.Error()) {
			text = m.errOpenRouterNetwork
		}
	}
	showError(text)
}

type openRouterChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Code    interface{} `json:"code"`
		Message string      `json:"message"`
	} `json:"error,omitempty"`
}

var digitSequencePattern = regexp.MustCompile(`\d+`)

func analyzeImageForAnyDesk(pngData []byte) (string, error) {
	key, err := loadOpenRouterAPIKey()
	if err != nil {
		return "", err
	}
	if key == "" {
		return "", fmt.Errorf("%s", currentMessages().errOpenRouterKeyEmpty)
	}
	prompt := "Inspect this screenshot for an AnyDesk address. An AnyDesk address must be exactly 9 or 10 digits. Return only the single numeric address. If none exists, return NONE. If multiple possible 9 or 10 digit addresses exist, return MULTIPLE. Do not return any other text."
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData)
	payload := map[string]interface{}{
		"model":       openRouterModel,
		"temperature": 0,
		"max_tokens":  32,
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": []interface{}{
			map[string]interface{}{"type": "text", "text": prompt},
			map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": dataURL}},
		}}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Quick-Anydesk-Connect/1.1")
	req.Header.Set("X-OpenRouter-Title", "Quick Anydesk Connect")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", errOpenRouterTimeout
		}
		return "", fmt.Errorf("%w: %v", errOpenRouterNetwork, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusPaymentRequired:
		return "", errOpenRouterPayment
	case http.StatusBadRequest:
		return "", errOpenRouterBadRequest
	case http.StatusNotFound:
		return "", errOpenRouterModelUnavailable
	case http.StatusUnauthorized:
		return "", errOpenRouterUnauthorized
	case http.StatusForbidden:
		return "", errOpenRouterForbidden
	case http.StatusTooManyRequests:
		return "", errOpenRouterRateLimit
	default:
		if resp.StatusCode >= 500 {
			return "", errOpenRouterServer
		}
		return "", fmt.Errorf("OpenRouter HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed openRouterChatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", errOpenRouterInvalidResponse
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("OpenRouter: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", errOpenRouterInvalidResponse
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	upper := strings.ToUpper(content)
	if upper == "NONE" {
		return "", errAnyDeskNotFound
	}
	if upper == "MULTIPLE" {
		return "", errMultipleAnyDeskIDs
	}
	seqs := digitSequencePattern.FindAllString(content, -1)
	uniq := map[string]bool{}
	for _, seq := range seqs {
		if isAnyDeskID(seq) {
			uniq[seq] = true
		}
	}
	if len(uniq) == 0 {
		return "", errOpenRouterInvalidResponse
	}
	if len(uniq) > 1 {
		return "", errMultipleAnyDeskIDs
	}
	for id := range uniq {
		return id, nil
	}
	return "", errOpenRouterInvalidResponse
}

func showImageAnalysisError(err error) {
	m := currentMessages()
	text := err.Error()
	switch {
	case errors.Is(err, errAnyDeskNotFound):
		text = m.anyDeskNotFoundInImage
	case errors.Is(err, errMultipleAnyDeskIDs):
		text = m.multipleAnyDeskIDs
	case errors.Is(err, errOpenRouterPayment):
		text = m.errOpenRouterPayment
	case errors.Is(err, errOpenRouterUnauthorized):
		text = m.errOpenRouterUnauthorized
	case errors.Is(err, errOpenRouterForbidden):
		text = m.errOpenRouterForbidden
	case errors.Is(err, errOpenRouterRateLimit):
		text = m.errOpenRouterRateLimit
	case errors.Is(err, errOpenRouterServer):
		text = m.errOpenRouterServer
	case errors.Is(err, errOpenRouterTimeout):
		text = m.errImageAnalysisTimeout
	case errors.Is(err, errOpenRouterInvalidResponse):
		text = m.errOpenRouterInvalidResponse
	case errors.Is(err, errOpenRouterBadRequest):
		text = m.errOpenRouterBadRequest
	case errors.Is(err, errOpenRouterModelUnavailable):
		text = m.errOpenRouterModelUnavailable
	default:
		if strings.Contains(err.Error(), errOpenRouterNetwork.Error()) {
			text = m.errOpenRouterNetwork
		}
	}
	showError(text)
}
