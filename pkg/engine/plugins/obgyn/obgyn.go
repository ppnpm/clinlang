package obgyn

import (
	"clinlang/pkg/engine"
	"strings"
)

type ObGynPlugin struct{}

func (p *ObGynPlugin) GetName() string { return "obgyn" }

func (p *ObGynPlugin) GetDescription() string {
	return "OB/GYN Specialty — obstetric and gynaecologic fields"
}

func (p *ObGynPlugin) GetCommandSummary() map[string]string {
	return map[string]string{
		"lmp":  "Last menstrual period date",
		"edd":  "Estimated date of delivery",
		"gpal": "Gravida, Para, Abortus, Living",
		"fhs":  "Fetal heart sounds",
		"ctx":  "Contractions",
	}
}

type ObGynData struct {
	LMP  string `json:"lmp,omitempty"`
	EDD  string `json:"edd,omitempty"`
	GPAL string `json:"gpal,omitempty"`
	FHS  string `json:"fhs,omitempty"`
	CTX  string `json:"ctx,omitempty"`
}

func (p *ObGynPlugin) InitData() any {
	return &ObGynData{}
}

func (p *ObGynPlugin) GetCommands() map[string]engine.ParserFunc {
	return map[string]engine.ParserFunc{
		"lmp": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.LMP = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"edd": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.EDD = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"gpal": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.GPAL = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"fhs": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.FHS = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"ctx": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.CTX = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
	}
}

func init() {
	engine.RegisterPlugin(&ObGynPlugin{})
}