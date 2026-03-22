package engine

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed drugs.json
var drugsJSON []byte

var DrugsList []string

func init() {
	err := json.Unmarshal(drugsJSON, &DrugsList)
	if err != nil {
		// Log or handle error if needed, for now we keep an empty list
		DrugsList = []string{}
	}
}

// SearchDrugs returns a list of drug names that start with the given prefix.
func SearchDrugs(prefix string) []string {
	prefix = strings.ToLower(prefix)
	var matches []string
	
	if prefix == "" {
		return matches
	}

	for _, drug := range DrugsList {
		if strings.HasPrefix(strings.ToLower(drug), prefix) {
			matches = append(matches, drug)
		}
	}
	
	// Optional: limit results to avoid huge dropdowns
	if len(matches) > 15 {
		matches = matches[:15]
	}

	return matches
}
