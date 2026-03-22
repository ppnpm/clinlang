//go:build js && wasm

package main

import (
	"encoding/json"
	"sort"
	"syscall/js"

	"clinlang/pkg/engine"
	_ "clinlang/pkg/engine/plugins/obgyn"
)

// parseClinLang is exported to JavaScript.
// It takes a single string parameter (the .cln text)
// and returns a JSON string containing the formatted SOAP note and raw JSON structure.
func parseClinLang(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return `{"error": "No input provided"}`
	}

	input := args[0].String()
	if input == "" {
		return `{"soap": "", "json": null, "abnormal_flags": []}`
	}

	// Parse using the core engine
	c := engine.ParseString(input)

	// Format output
	soapNote := engine.FormatSOAP(c)

	// Prepare result payload for the frontend
	result := map[string]interface{}{
		"soap":           soapNote,
		"json":           c,
		"abnormal_flags": c.AbnormalFlags,
		"warnings":       c.Warnings,
	}

	b, err := json.Marshal(result)
	if err != nil {
		return `{"error": "Failed to marshal JSON"}`
	}

	return string(b)
}

// searchDrugs is exported to JavaScript.
// It takes a single string parameter (the search prefix)
// and returns a JSON string containing an array of matching drugs.
func searchDrugs(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return `[]`
	}

	prefix := args[0].String()
	matches := engine.SearchDrugs(prefix)

	if len(matches) == 0 {
		return `[]`
	}

	b, err := json.Marshal(matches)
	if err != nil {
		return `[]`
	}

	return string(b)
}

// getAutocompleteCommands is exported to JavaScript.
func getAutocompleteCommands(this js.Value, args []js.Value) any {
	profile := "general"
	if len(args) > 0 {
		profile = args[0].String()
	}

	type CommandDesc struct {
		Cmd  string `json:"cmd"`
		Desc string `json:"desc"`
	}

	var cmds []CommandDesc

	// Add core commands
	core := engine.GetCoreCommandNames()
	for _, cmd := range core {
		desc := "Core Command"
		switch cmd {
		case "pt": desc = "Patient Demographics"
		case "cc": desc = "Chief Complaint"
		case "hpi": desc = "History of Presenting Illness"
		case "pmh": desc = "Past Medical History"
		case "dx": desc = "Diagnosis"
		case "ddx": desc = "Differential Diagnosis"
		case "sx": desc = "Symptoms"
		case "vitals": desc = "Vital Signs"
		case "rx": desc = "Prescription"
		case "id": desc = "Patient Identifier"
		}
		cmds = append(cmds, CommandDesc{Cmd: cmd, Desc: desc})
	}

	// Add plugin commands
	plugin := engine.GetPlugin(profile)
	if plugin != nil {
		for cmd := range plugin.GetCommands() {
			cmds = append(cmds, CommandDesc{Cmd: cmd, Desc: "Specialty Command (" + profile + ")"})
		}
	}

	// Sort alphabetically
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Cmd < cmds[j].Cmd
	})

	b, _ := json.Marshal(cmds)
	return string(b)
}

func main() {
	// A Go WebAssembly program needs to keep running in the background.
	// We create an empty channel and block on it forever.
	c := make(chan struct{})

	// Expose the functions to the global JavaScript window object
	js.Global().Set("parseClinLang", js.FuncOf(parseClinLang))
	js.Global().Set("searchDrugs", js.FuncOf(searchDrugs))
	js.Global().Set("getAutocompleteCommands", js.FuncOf(getAutocompleteCommands))

	// Block forever
	<-c
}
