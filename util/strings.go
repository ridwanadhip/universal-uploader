package util

import (
	"encoding/json"
	"unicode"
)

func FindWordsByPrefix(text string, prefix rune) (res []string) {
	res = []string{}

	runes := []rune(text)
	start := -1
	for i := 0; i < len(runes); i++ {
		if start == -1 && runes[i] == prefix {
			start = i
		} else if start > -1 {
			var found string
			if unicode.IsSpace(runes[i]) {
				found = string(runes[start:i])
				start = -1
			} else if runes[i] == prefix {
				found = string(runes[start:i])
				start = i
			} else if i+1 == len(runes) {
				found = string(runes[start : i+1])
			}

			if found != "" && found != string(prefix) {
				res = append(res, found)
			}
		}
	}

	return res
}

func FindSurroundedWords(text string, character rune) (res []string) {
	res = []string{}

	runes := []rune(text)
	start := -1
	for i := 0; i < len(runes); i++ {
		match := runes[i] == character
		if match && start == -1 {
			start = i
		} else if match && start > -1 {
			found := string(runes[start : i+1])
			res = append(res, found)
			start = -1
		}
	}

	return res
}

func RemoveToken(text string) string {
	length := len(text)
	if length < 2 {
		return text
	}

	return text[1 : length-1]
}

func Jsonify(anything any) string {
	marshalled, _ := json.Marshal(anything)
	return string(marshalled)
}
