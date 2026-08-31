package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const anyDeskPasswordMinLength = 8

func normalizeAnyDeskPasswordInput(value string) string {
	replacer := strings.NewReplacer("\r", "", "\n", "", "\t", "")
	return replacer.Replace(value)
}

func validateAnyDeskPasswordInput(value string) error {
	if value == "" {
		return fmt.Errorf("%s", currentMessages().emptyPassword)
	}
	if utf8.RuneCountInString(value) < anyDeskPasswordMinLength {
		if language == "en" {
			return fmt.Errorf("AnyDesk unattended-access passwords must be at least %d characters.", anyDeskPasswordMinLength)
		}
		return fmt.Errorf("AnyDesk 무인 접속 암호는 최소 %d자 이상이어야 합니다.", anyDeskPasswordMinLength)
	}
	return nil
}

func normalizeOpenRouterKeyInput(value string) string {
	return strings.TrimSpace(value)
}
