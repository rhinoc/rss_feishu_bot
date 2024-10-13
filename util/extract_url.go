package util

import "regexp"

var urlRegex = regexp.MustCompile(`https?:\/\/[^\s]+`)

func ExtractUrlList(text string) []string {
	return urlRegex.FindAllString(text, -1)
}

func ExtractUrl(text string) string {
	return urlRegex.FindString(text)
}
