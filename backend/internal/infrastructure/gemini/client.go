package gemini

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"learning/internal/ai"
	"learning/internal/anomaly"
	"learning/internal/event"
	"learning/internal/reading"
)

//go:embed prompts/analysis.txt
var analysisPrompt string

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

func (client *Client) GenerateConsumptionInsight(ctx context.Context, input ai.ConsumptionInsightInput) (ai.ConsumptionInsightResult, error) {
	prompt, err := buildOptimizationPrompt(input)
	if err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("build optimization prompt: %w", err)
	}
	requestBody := geminiRequest{
		Contents: []content{{Role: "user", Parts: []part{{Text: prompt}}}},
		GenerationConfig: generationConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   optimizationResponseSchema(),
		},
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("encode optimization request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+client.model+":generateContent?key="+client.apiKey, bytes.NewReader(body))
	if err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("create optimization request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("call Gemini optimization: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("read optimization response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("Gemini returned status %d", response.StatusCode)
	}
	var result geminiResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("decode optimization response: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("Gemini returned no optimization content")
	}
	text := strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(strings.TrimSpace(text), "```")
	var insight ai.ConsumptionInsightResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &insight); err != nil {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("decode structured optimization insight: %w", err)
	}
	if strings.TrimSpace(insight.Summary) == "" || len(insight.Recommendations) == 0 {
		return ai.ConsumptionInsightResult{}, fmt.Errorf("Gemini returned incomplete optimization insight")
	}
	if insight.GeneratedAt.IsZero() {
		insight.GeneratedAt = time.Now().UTC()
	}
	return insight, nil
}

func NewClient(apiKey, model string) *Client {
	if strings.TrimSpace(apiKey) == "" {
		return nil
	}
	if strings.TrimSpace(model) == "" {
		model = "gemini-2.0-flash"
	}
	return &Client{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 45 * time.Second},
		baseURL:    "https://generativelanguage.googleapis.com/v1beta/models/",
	}
}

func (client *Client) Analyze(ctx context.Context, input ai.AnalysisContext) (ai.Insight, error) {
	prompt, err := buildPrompt(input)
	if err != nil {
		return ai.Insight{}, fmt.Errorf("build Gemini prompt: %w", err)
	}
	requestBody := geminiRequest{
		Contents: []content{{Role: "user", Parts: []part{{Text: prompt}}}},
		GenerationConfig: generationConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   responseSchema(),
		},
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return ai.Insight{}, fmt.Errorf("encode Gemini request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+client.model+":generateContent?key="+client.apiKey, bytes.NewReader(body))
	if err != nil {
		return ai.Insight{}, fmt.Errorf("create Gemini request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return ai.Insight{}, fmt.Errorf("call Gemini: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return ai.Insight{}, fmt.Errorf("read Gemini response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return ai.Insight{}, fmt.Errorf("Gemini returned status %d", response.StatusCode)
	}
	var result geminiResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return ai.Insight{}, fmt.Errorf("decode Gemini response: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return ai.Insight{}, fmt.Errorf("Gemini returned no content")
	}
	text := strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(strings.TrimSpace(text), "```")
	var insight ai.Insight
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &insight); err != nil {
		return ai.Insight{}, fmt.Errorf("decode structured Gemini insight: %w", err)
	}
	if strings.TrimSpace(insight.Answer) == "" || strings.TrimSpace(insight.Explanation) == "" || len(insight.SuggestedQuestions) != 3 {
		return ai.Insight{}, fmt.Errorf("Gemini returned an incomplete structured insight")
	}
	return insight, nil
}

type geminiRequest struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type content struct {
	Role  string `json:"role"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	ResponseMIMEType string         `json:"responseMimeType"`
	ResponseSchema   map[string]any `json:"responseSchema"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []part `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type promptEvent struct {
	MeterID     string    `json:"meter_id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
}

type meterStats struct {
	MeterID            string  `json:"meter_id"`
	ReadingCount       int     `json:"reading_count"`
	AverageConsumption float64 `json:"average_consumption_kwh"`
	MinimumConsumption float64 `json:"minimum_consumption_kwh"`
	MaximumConsumption float64 `json:"maximum_consumption_kwh"`
	AverageVoltage     float64 `json:"average_voltage_v"`
	AverageCurrent     float64 `json:"average_current_a"`
	AveragePowerFactor float64 `json:"average_power_factor"`
}

func buildPrompt(input ai.AnalysisContext) (string, error) {
	payload := struct {
		ReadingCount int               `json:"reading_count"`
		MeterCount   int               `json:"meter_count"`
		Meters       []meterStats      `json:"meters"`
		Events       []promptEvent     `json:"events"`
		Anomalies    []anomaly.Anomaly `json:"anomalies"`
	}{
		ReadingCount: input.ReadingCount,
		MeterCount:   input.MeterCount,
		Meters:       summarizeMeters(input.Readings),
		Events:       mapEvents(input.Events),
		Anomalies:    input.Anomalies,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(analysisPrompt) + " Datos: " + string(data), nil
}

func mapEvents(events []event.Event) []promptEvent {
	result := make([]promptEvent, 0, len(events))
	for _, item := range events {
		result = append(result, promptEvent{MeterID: item.MeterID, Timestamp: item.Timestamp, Type: item.Type, Description: item.Description})
	}
	return result
}

func summarizeMeters(readings []reading.Reading) []meterStats {
	groups := make(map[string][]reading.Reading)
	for _, item := range readings {
		groups[item.MeterID] = append(groups[item.MeterID], item)
	}
	result := make([]meterStats, 0, len(groups))
	for meterID, series := range groups {
		stats := meterStats{MeterID: meterID, ReadingCount: len(series), MinimumConsumption: series[0].ConsumptionKWh, MaximumConsumption: series[0].ConsumptionKWh}
		for _, item := range series {
			stats.AverageConsumption += item.ConsumptionKWh
			stats.AverageVoltage += item.VoltageV
			stats.AverageCurrent += item.CurrentA
			stats.AveragePowerFactor += item.PowerFactor
			if item.ConsumptionKWh < stats.MinimumConsumption {
				stats.MinimumConsumption = item.ConsumptionKWh
			}
			if item.ConsumptionKWh > stats.MaximumConsumption {
				stats.MaximumConsumption = item.ConsumptionKWh
			}
		}
		count := float64(len(series))
		stats.AverageConsumption /= count
		stats.AverageVoltage /= count
		stats.AverageCurrent /= count
		stats.AveragePowerFactor /= count
		result = append(result, stats)
	}
	return result
}

func responseSchema() map[string]any {
	return map[string]any{
		"type": "OBJECT",
		"properties": map[string]any{
			"answer":      map[string]any{"type": "STRING"},
			"explanation": map[string]any{"type": "STRING"},
			"suggestedQuestions": map[string]any{
				"type": "ARRAY",
				"items": map[string]any{
					"type": "OBJECT",
					"properties": map[string]any{
						"question": map[string]any{"type": "STRING"},
						"answer":   map[string]any{"type": "STRING"},
					},
					"required": []string{"question", "answer"},
				},
			},
		},
		"required": []string{"answer", "explanation", "suggestedQuestions"},
	}
}

func optimizationResponseSchema() map[string]any {
	return map[string]any{
		"type": "OBJECT",
		"properties": map[string]any{
			"summary":                map[string]any{"type": "STRING"},
			"recommendations":        map[string]any{"type": "ARRAY", "items": map[string]any{"type": "STRING"}},
			"potential_savings_kwh":  map[string]any{"type": "NUMBER"},
			"potential_savings_cost": map[string]any{"type": "NUMBER"},
			"currency":               map[string]any{"type": "STRING"},
		},
		"required": []string{"summary", "recommendations", "potential_savings_kwh", "potential_savings_cost", "currency"},
	}
}

func buildOptimizationPrompt(input ai.ConsumptionInsightInput) (string, error) {
	payload := map[string]any{
		"meter_id":                input.MeterID,
		"provider":                input.Provider,
		"region":                  input.Region,
		"from":                    input.From.Format(time.RFC3339),
		"to":                      input.To.Format(time.RFC3339),
		"current_consumption_kwh": input.CurrentConsumptionKWh,
		"baseline_kwh":            input.BaselineKWh,
		"average_power_factor":    input.AveragePowerFactor,
		"estimated_cost":          input.EstimatedCost,
		"currency":                input.Currency,
		"reading_count":           len(input.Readings),
		"readings":                input.Readings,
		"anomalies":               input.Anomalies,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Eres un analista energético. Genera una recomendación accionable para este medidor. Responde en JSON con: summary, recommendations (lista de 3 o más), potential_savings_kwh, potential_savings_cost y currency. Usa datos reales del medidor y da recomendaciones concretas y medibles. Datos: %s", string(data)), nil
}
