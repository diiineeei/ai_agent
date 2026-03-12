// cmd/mcp-weather/main.go
//
// Servidor MCP de previsão do tempo — transporte HTTP.
// Endpoint: POST http://localhost:PORT/mcp
//
// Uso:
//
//	go run ./cmd/mcp-weather
//	MCP_PORT=3001 go run ./cmd/mcp-weather
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// ---------------------------------------------------------------------------
// JSON-RPC 2.0 types
// ---------------------------------------------------------------------------

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	errParse          = -32700
	errMethodNotFound = -32601
	errInvalidParams  = -32602
)

// ---------------------------------------------------------------------------
// MCP protocol types
// ---------------------------------------------------------------------------

type InitializeResult struct {
	ProtocolVersion string     `json:"protocolVersion"`
	Capabilities    struct{}   `json:"capabilities"`
	ServerInfo      ServerInfo `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolCallResult struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ---------------------------------------------------------------------------
// Weather tool — Open-Meteo (gratuito, sem API key)
// ---------------------------------------------------------------------------

type geoResult struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}

type weatherResult struct {
	Current struct {
		Temperature   float64 `json:"temperature_2m"`
		Humidity      int     `json:"relative_humidity_2m"`
		WindSpeed     float64 `json:"wind_speed_10m"`
		WeatherCode   int     `json:"weather_code"`
		Precipitation float64 `json:"precipitation"`
	} `json:"current"`
}

func wmoDescription(code int) string {
	switch {
	case code == 0:
		return "☀️ Céu limpo"
	case code <= 2:
		return "⛅ Parcialmente nublado"
	case code == 3:
		return "☁️ Nublado"
	case code <= 48:
		return "🌫️ Neblina"
	case code <= 55:
		return "🌦️ Garoa"
	case code <= 65:
		return "🌧️ Chuva"
	case code <= 75:
		return "❄️ Neve"
	case code <= 82:
		return "🌨️ Pancadas de chuva"
	case code <= 84:
		return "🌩️ Granizo"
	case code <= 99:
		return "⛈️ Tempestade"
	default:
		return "❓ Desconhecido"
	}
}

func fetchJSON(rawURL string, dest any) error {
	resp, err := http.Get(rawURL) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dest)
}

func getWeather(location string) (string, error) {
	geoURL := "https://geocoding-api.open-meteo.com/v1/search?" +
		url.Values{
			"name":     {location},
			"count":    {"1"},
			"language": {"pt"},
			"format":   {"json"},
		}.Encode()

	var geo geoResult
	if err := fetchJSON(geoURL, &geo); err != nil {
		return "", fmt.Errorf("geocoding: %w", err)
	}
	if len(geo.Results) == 0 {
		return "", fmt.Errorf("cidade %q não encontrada", location)
	}

	place := geo.Results[0]

	weatherURL := "https://api.open-meteo.com/v1/forecast?" +
		url.Values{
			"latitude":        {fmt.Sprintf("%f", place.Latitude)},
			"longitude":       {fmt.Sprintf("%f", place.Longitude)},
			"current":         {"temperature_2m,relative_humidity_2m,wind_speed_10m,weather_code,precipitation"},
			"wind_speed_unit": {"kmh"},
			"timezone":        {"auto"},
		}.Encode()

	var w weatherResult
	if err := fetchJSON(weatherURL, &w); err != nil {
		return "", fmt.Errorf("clima: %w", err)
	}

	c := w.Current
	return fmt.Sprintf(
		"Previsão do tempo para %s, %s:\n• Temperatura: %.1f°C\n• Condição: %s\n• Umidade: %d%%\n• Vento: %.0f km/h\n• Precipitação: %.1f mm",
		place.Name, place.Country,
		c.Temperature,
		wmoDescription(c.WeatherCode),
		c.Humidity,
		c.WindSpeed,
		c.Precipitation,
	), nil
}

// ---------------------------------------------------------------------------
// MCP tools registry
// ---------------------------------------------------------------------------

var tools = []Tool{
	{
		Name:        "get_weather",
		Description: "Retorna as condições climáticas atuais para uma cidade.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]any{
					"type":        "string",
					"description": "Nome da cidade (ex.: São Paulo, Rio de Janeiro)",
				},
			},
			"required": []string{"location"},
		},
	},
}

func callTool(name string, args map[string]any) ToolCallResult {
	switch name {
	case "get_weather":
		location, _ := args["location"].(string)
		if location == "" {
			return ToolCallResult{
				Content: []TextContent{{Type: "text", Text: "Erro: campo 'location' é obrigatório"}},
				IsError: true,
			}
		}
		text, err := getWeather(location)
		if err != nil {
			return ToolCallResult{
				Content: []TextContent{{Type: "text", Text: "Erro ao buscar clima: " + err.Error()}},
				IsError: true,
			}
		}
		return ToolCallResult{Content: []TextContent{{Type: "text", Text: text}}}

	default:
		return ToolCallResult{
			Content: []TextContent{{Type: "text", Text: "Ferramenta desconhecida: " + name}},
			IsError: true,
		}
	}
}

// ---------------------------------------------------------------------------
// JSON-RPC dispatcher
// ---------------------------------------------------------------------------

func handle(req *Request) *Response {
	resp := &Response{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case "initialize":
		resp.Result = InitializeResult{
			ProtocolVersion: "2024-11-05",
			ServerInfo:      ServerInfo{Name: "mcp-weather", Version: "1.0.0"},
		}

	case "notifications/initialized":
		return nil // notificação — sem resposta

	case "ping":
		resp.Result = map[string]any{}

	case "tools/list":
		resp.Result = ToolsListResult{Tools: tools}

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &RPCError{Code: errInvalidParams, Message: "params inválidos: " + err.Error()}
			return resp
		}
		resp.Result = callTool(params.Name, params.Arguments)

	default:
		resp.Error = &RPCError{Code: errMethodNotFound, Message: "método não encontrado: " + req.Method}
	}

	return resp
}

// ---------------------------------------------------------------------------
// HTTP handler
// ---------------------------------------------------------------------------

func mcpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		json.NewEncoder(w).Encode(Response{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: errParse, Message: "JSON inválido: " + err.Error()},
		})
		return
	}

	resp := handle(&req)
	if resp == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "3001"
	}

	http.HandleFunc("/mcp", mcpHandler)

	log.Printf("mcp-weather: ouvindo em http://localhost:%s/mcp", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("servidor: %v", err)
	}
}
