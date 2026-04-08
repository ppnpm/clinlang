// Package autocomplete provides desktop UI-specific suggestion logic.
// It wraps pkg/engine functions — it is NOT a clone of the engine.
package autocomplete

import (
	"sort"
	"strings"

	"clinlang/pkg/engine"
)

// Suggestion is a single autocomplete result for the frontend.
type Suggestion struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// autocompleteDB holds static suggestion lists by command keyword.
var autocompleteDB = map[string][]Suggestion{
	"sx": {
		{Label: "chest_pain", Value: "chest_pain", Description: "Chest pain", Category: "sx"},
		{Label: "sob", Value: "sob", Description: "Shortness of breath", Category: "sx"},
		{Label: "fever", Value: "fever", Description: "Fever", Category: "sx"},
		{Label: "cough", Value: "cough", Description: "Cough", Category: "sx"},
		{Label: "nausea", Value: "nausea", Description: "Nausea", Category: "sx"},
		{Label: "vomiting", Value: "vomiting", Description: "Vomiting", Category: "sx"},
		{Label: "headache", Value: "headache", Description: "Headache", Category: "sx"},
		{Label: "abd_pain", Value: "abd_pain", Description: "Abdominal pain", Category: "sx"},
		{Label: "diarrhea", Value: "diarrhea", Description: "Diarrhea", Category: "sx"},
		{Label: "dysuria", Value: "dysuria", Description: "Painful urination", Category: "sx"},
		{Label: "diaphoresis", Value: "diaphoresis", Description: "Excessive sweating", Category: "sx"},
		{Label: "palpitations", Value: "palpitations", Description: "Heart palpitations", Category: "sx"},
		{Label: "syncope", Value: "syncope", Description: "Fainting/syncope", Category: "sx"},
		{Label: "edema", Value: "edema", Description: "Peripheral edema", Category: "sx"},
	},
	"ix": {
		{Label: "hb", Value: "hb", Description: "Hemoglobin", Category: "ix"},
		{Label: "wbc", Value: "wbc", Description: "White blood cell count", Category: "ix"},
		{Label: "plt", Value: "plt", Description: "Platelet count", Category: "ix"},
		{Label: "na", Value: "na", Description: "Sodium", Category: "ix"},
		{Label: "k", Value: "k", Description: "Potassium", Category: "ix"},
		{Label: "cr", Value: "cr", Description: "Creatinine", Category: "ix"},
		{Label: "cxr:", Value: "cxr:", Description: "Chest X-ray", Category: "ix"},
		{Label: "ecg:", Value: "ecg:", Description: "ECG", Category: "ix"},
		{Label: "usg:", Value: "usg:", Description: "Ultrasound", Category: "ix"},
		{Label: "hiv-", Value: "hiv-", Description: "HIV serology negative", Category: "ix"},
		{Label: "hbv-", Value: "hbv-", Description: "Hepatitis B negative", Category: "ix"},
	},
	"dx": {
		{Label: "Community Acquired Pneumonia", Value: "Community Acquired Pneumonia", Description: "CAP", Category: "dx"},
		{Label: "Acute Coronary Syndrome", Value: "Acute Coronary Syndrome", Description: "ACS", Category: "dx"},
		{Label: "Anterior STEMI", Value: "Anterior STEMI", Description: "ST elevation MI - anterior", Category: "dx"},
		{Label: "Diabetic Ketoacidosis", Value: "Diabetic Ketoacidosis", Description: "DKA", Category: "dx"},
		{Label: "Sepsis", Value: "Sepsis", Description: "Infection + organ dysfunction", Category: "dx"},
	},
	"pmh": {
		{Label: "dm2", Value: "dm2", Description: "Type 2 Diabetes Mellitus", Category: "pmh"},
		{Label: "htn", Value: "htn", Description: "Hypertension", Category: "pmh"},
		{Label: "ihd", Value: "ihd", Description: "Ischaemic Heart Disease", Category: "pmh"},
	},
}

// GetSuggestions returns autocomplete suggestions for a given command and partial query.
// For "rx" it delegates to engine.SearchDrugs.
func GetSuggestions(command, query string) []Suggestion {
	command = strings.ToLower(command)
	query = strings.ToLower(strings.TrimSpace(query))

	if command == "rx" {
		matches := engine.SearchDrugs(query)
		results := make([]Suggestion, 0, len(matches))
		for _, m := range matches {
			results = append(results, Suggestion{
				Label:       m,
				Value:       m,
				Description: "Drug",
				Category:    "rx",
			})
		}
		return results
	}

	list, ok := autocompleteDB[command]
	if !ok {
		return []Suggestion{}
	}

	if query == "" {
		if len(list) > 8 {
			return list[:8]
		}
		return list
	}

	var matched []Suggestion
	for _, s := range list {
		if strings.HasPrefix(strings.ToLower(s.Label), query) ||
			strings.HasPrefix(strings.ToLower(s.Value), query) {
			matched = append(matched, s)
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return len(matched[i].Label) < len(matched[j].Label)
	})

	if len(matched) > 8 {
		return matched[:8]
	}
	return matched
}
