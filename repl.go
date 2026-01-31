package main

import "strings"

func cleanInput(text string) []string {
	// implementation goes here
	// trim spaces and split the text into words
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	if text == "" {
		return []string{}
	}
	words := strings.Fields(text)
	return words
}