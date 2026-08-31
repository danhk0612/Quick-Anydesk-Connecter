package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadOrCreateConfig(path string) (string, string, bool, string, error) {
	p, lang, imageEnabled, model, err := readConfig(path)
	if err == nil {
		if lang != "en" {
			lang = "ko"
		}
		return p, lang, imageEnabled, model, nil
	}

	if !os.IsNotExist(err) {
		return "", "ko", false, defaultOpenRouterModel, err
	}

	language = "ko"
	for {
		p, err = askPassword()
		if err != nil {
			return "", "ko", false, defaultOpenRouterModel, err
		}
		p = normalizeAnyDeskPasswordInput(p)
		if err := validateAnyDeskPasswordInput(p); err != nil {
			showError(err.Error())
			continue
		}
		break
	}

	if err := saveConfig(path, p, "ko", false, defaultOpenRouterModel); err != nil {
		return "", "ko", false, defaultOpenRouterModel, err
	}
	return p, "ko", false, defaultOpenRouterModel, nil
}

func readConfig(path string) (string, string, bool, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "ko", false, defaultOpenRouterModel, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	section := ""
	p := ""
	lang := "ko"
	imageEnabled := false
	model := defaultOpenRouterModel

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section = strings.ToLower(strings.TrimSpace(trimmed[1 : len(trimmed)-1]))
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:eq]))
		value := line[eq+1:]

		switch section {
		case "anydesk":
			if key == "password" {
				p = normalizeAnyDeskPasswordInput(value)
			}
		case "general":
			switch key {
			case "language":
				if strings.EqualFold(strings.TrimSpace(value), "en") {
					lang = "en"
				}
			case "image_analysis":
				imageEnabled = strings.EqualFold(strings.TrimSpace(value), "true") || strings.TrimSpace(value) == "1"
			}
		case "openrouter":
			if key == "model" && strings.TrimSpace(value) != "" {
				model = strings.TrimSpace(value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "ko", false, defaultOpenRouterModel, fmt.Errorf(currentMessages().errConfigRead, err)
	}
	if p == "" {
		return "", "ko", false, defaultOpenRouterModel, fmt.Errorf("%s", currentMessages().errPasswordMissing)
	}
	if err := validateAnyDeskPasswordInput(p); err != nil {
		return "", "ko", false, defaultOpenRouterModel, err
	}
	return p, lang, imageEnabled, model, nil
}

func saveConfig(path, p, lang string, imageEnabled bool, model string) error {
	p = normalizeAnyDeskPasswordInput(p)
	if err := validateAnyDeskPasswordInput(p); err != nil {
		return err
	}
	if lang != "en" {
		lang = "ko"
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultOpenRouterModel
	}
	imageValue := "false"
	if imageEnabled {
		imageValue = "true"
	}

	content := "[anydesk]\r\npassword=" + p + "\r\n\r\n[general]\r\nlanguage=" + lang + "\r\nimage_analysis=" + imageValue + "\r\n\r\n[openrouter]\r\nmodel=" + model + "\r\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf(currentMessages().errConfigSave, err)
	}
	return nil
}

func persistConfig() error {
	cleanPassword := normalizeAnyDeskPasswordInput(password)
	if err := validateAnyDeskPasswordInput(cleanPassword); err != nil {
		return err
	}
	if err := saveConfig(configPath, cleanPassword, language, imageAnalysisEnabled, openRouterModel); err != nil {
		return err
	}
	password = cleanPassword
	return nil
}

func setLanguage(lang string) {
	if lang != "en" {
		lang = "ko"
	}
	if language == lang {
		return
	}
	old := language
	language = lang
	if err := persistConfig(); err != nil {
		language = old
		showError(err.Error())
		return
	}
	refreshTrayAppearance()
}

func setImageAnalysisEnabled(enabled bool) error {
	old := imageAnalysisEnabled
	imageAnalysisEnabled = enabled
	if err := persistConfig(); err != nil {
		imageAnalysisEnabled = old
		return err
	}
	refreshTrayAppearance()
	return nil
}
