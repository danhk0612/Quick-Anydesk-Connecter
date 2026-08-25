package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadOrCreateConfig(path string) (string, string, error) {
	p, lang, err := readConfig(path)
	if err == nil {
		if lang != "en" {
			lang = "ko"
		}
		return p, lang, nil
	}

	if !os.IsNotExist(err) {
		return "", "ko", err
	}

	// Default language is Korean for the first-run setup.
	language = "ko"
	p, err = askPassword()
	if err != nil {
		return "", "ko", err
	}

	if err := saveConfig(path, p, "ko"); err != nil {
		return "", "ko", err
	}
	return p, "ko", nil
}

func readConfig(path string) (string, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "ko", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	section := ""
	p := ""
	lang := "ko"

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
			if key == "language" {
				v := strings.ToLower(strings.TrimSpace(value))
				if v == "en" {
					lang = "en"
				} else {
					lang = "ko"
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "ko", fmt.Errorf(currentMessages().errConfigRead, err)
	}
	if p == "" {
		return "", "ko", fmt.Errorf("%s", currentMessages().errPasswordMissing)
	}

	return p, lang, nil
}

func saveConfig(path, p, lang string) error {
	if p == "" {
		return fmt.Errorf("%s", currentMessages().errPasswordEmpty)
	}
	if lang != "en" {
		lang = "ko"
	}

	content := "[anydesk]\r\npassword=" + p + "\r\n\r\n[general]\r\nlanguage=" + lang + "\r\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf(currentMessages().errConfigSave, err)
	}
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
	if err := saveConfig(configPath, password, language); err != nil {
		language = old
		showError(err.Error())
	}
}
