package main

import "strings"

func normalizeRepoName(repository string) string {
	return strings.ToLower(strings.TrimSpace(repository))
}
