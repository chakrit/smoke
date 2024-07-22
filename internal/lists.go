package internal

import "strings"

func Whitelist[T any](arr []T, whitelist []string, f func(T) string) []T {
	output := make([]T, 0, len(arr))
NextItem:
	for _, item := range arr {
		for _, white := range whitelist {
			if strings.Contains(f(item), white) {
				output = append(output, item)
				continue NextItem
			}
		}
	}
	return output
}

func Blacklist[T any](arr []T, blacklist []string, f func(T) string) []T {
	output := make([]T, 0, len(arr))
NextItem:
	for _, item := range arr {
		for _, black := range blacklist {
			if strings.Contains(f(item), black) {
				continue NextItem
			}
		}
		output = append(output, item)
	}
	return output
}
