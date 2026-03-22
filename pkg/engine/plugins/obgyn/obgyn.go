package obgyn

import (
	"clinlang/pkg/engine"
	"strings"
)

type ObGynPlugin struct{}

func (p *ObGynPlugin) GetName() string { return "obgyn" }

type ObGynData struct {
	LMP  string `json:"lmp,omitempty"`
	EDD  string `json:"edd,omitempty"`
	GPAL string `json:"gpal,omitempty"`
	FHS  string `json:"fhs,omitempty"`
	CTX  string `json:"ctx,omitempty"`
}

func (p *ObGynPlugin) InitData() interface{} {
	return &ObGynData{}
}

func (p *ObGynPlugin) GetCommands() map[string]engine.ParserFunc {
	return map[string]engine.ParserFunc{
		"lmp": func(tokens []string, c *engine.Case) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.LMP = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"edd": func(tokens []string, c *engine.Case) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.EDD = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"gpal": func(tokens []string, c *engine.Case) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.GPAL = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"fhs": func(tokens []string, c *engine.Case) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.FHS = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"ctx": func(tokens []string, c *engine.Case) {
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
