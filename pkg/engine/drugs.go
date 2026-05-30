package engine

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed drugs.json
var drugsJSON []byte

var DrugsList []string

type drugSearchEntry struct {
	name      string
	nameLower string
}

var drugsSearchList []drugSearchEntry

func init() {
	err := json.Unmarshal(drugsJSON, &DrugsList)
	if err != nil {
		// Log or handle error if needed, for now we keep an empty list
		DrugsList = []string{}
	}

	drugsSearchList = make([]drugSearchEntry, len(DrugsList))
	for i, drug := range DrugsList {
		drugsSearchList[i] = drugSearchEntry{
			name:      drug,
			nameLower: strings.ToLower(drug),
		}
	}
}

// SearchDrugs returns a list of drug names matching the prefix query, prioritized by:
// 1. Prefix matches (starts with prefix)
// 2. Word-boundary matches (prefix matches start of any word inside the name)
// 3. Substring matches (prefix matches anywhere in the name)
func SearchDrugs(prefix string) []string {
	return SearchDrugsWithList(prefix, DrugsList)
}

// SearchDrugsWithList searches for drugs in a custom provided slice.
func SearchDrugsWithList(prefix string, list []string) []string {
	prefix = strings.ToLower(prefix)
	if prefix == "" {
		return []string{}
	}

	var startsWith []string
	var wordStarts []string
	var contains []string

	for _, name := range list {
		nameLower := strings.ToLower(name)
		if strings.HasPrefix(nameLower, prefix) {
			startsWith = append(startsWith, name)
		} else if strings.Contains(nameLower, " "+prefix) {
			wordStarts = append(wordStarts, name)
		} else if strings.Contains(nameLower, prefix) {
			contains = append(contains, name)
		}
	}

	// Combine results in order of priority, up to 15 items
	var matches []string
	matches = append(matches, startsWith...)
	if len(matches) < 15 {
		matches = append(matches, wordStarts...)
	}
	if len(matches) < 15 {
		matches = append(matches, contains...)
	}

	if len(matches) > 15 {
		matches = matches[:15]
	}

	return matches
}
