package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadOrCreateConfig(path string) (string, string, bool, error) {
	p, lang, imageEnabled, err := readConfig(path)
	if err == nil {
		if lang != "en" {
			lang = "ko"
		}
		return p, lang, imageEnabled, nil
	}

	if !os.IsNotExist(err) {
		return "", "ko", false, err
	}

	language = "ko"
	p, err = askPassword()
	if err != nil {
		return "", "ko", false, err
	}

	if err := saveConfig(path, p, "ko", false); err != nil {
		return "", "ko", false, err
	}
	return p, "ko", false, nil
}

func readConfig(path string) (string, string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "ko", false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	section := ""
	p := ""
	lang := "ko"
	imageEnabled := false

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
				p = value
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
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "ko", false, fmt.Errorf(currentMessages().errConfigRead, err)
	}
	if p == "" {
		return "", "ko", false, fmt.Errorf("%s", currentMessages().errPasswordMissing)
	}
	return p, lang, imageEnabled, nil
}

func saveConfig(path, p, lang string, imageEnabled bool) error {
	if p == "" {
		return fmt.Errorf("%s", currentMessages().errPasswordEmpty)
	}
	if lang != "en" {
		lang = "ko"
	}
	imageValue := "false"
	if imageEnabled {
		imageValue = "true"
	}

	content := "[anydesk]\r\npassword=" + p + "\r\n\r\n[general]\r\nlanguage=" + lang + "\r\nimage_analysis=" + imageValue + "\r\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf(currentMessages().errConfigSave, err)
	}
	return nil
}

func persistConfig() error {
	return saveConfig(configPath, password, language, imageAnalysisEnabled)
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
	}
}

func setImageAnalysisEnabled(enabled bool) error {
	old := imageAnalysisEnabled
	imageAnalysisEnabled = enabled
	if err := persistConfig(); err != nil {
		imageAnalysisEnabled = old
		return err
	}
	return nil
}
