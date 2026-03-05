package skills

import (
	"context"
	"fmt"

	"ai_agent/internal/agent"
)

// WeatherSkill returns mock weather data for a given city.
type WeatherSkill struct{}

func (w *WeatherSkill) Name() string { return "weather" }

func (w *WeatherSkill) Declaration() *agent.FunctionDeclaration {
	return &agent.FunctionDeclaration{
		Name:        "weather",
		Description: "Retorna as condições climáticas atuais para uma cidade.",
		ParametersSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]any{
					"type":        "string",
					"description": "Nome da cidade",
				},
			},
			"required": []string{"location"},
		},
		ResponseSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location":    map[string]any{"type": "string"},
				"temperature": map[string]any{"type": "string"},
				"condition":   map[string]any{"type": "string"},
			},
		},
		FunctionCall: func(ctx context.Context, args map[string]any) (map[string]any, error) {
			loc, ok := args["location"].(string)
			if !ok || loc == "" {
				return nil, fmt.Errorf("argumento location é obrigatório")
			}
			return map[string]any{
				"location":    loc,
				"temperature": "22°C",
				"condition":   "Ensolarado",
			}, nil
		},
	}
}

var _ Skill = (*WeatherSkill)(nil)
