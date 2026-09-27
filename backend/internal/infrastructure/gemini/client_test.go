package gemini

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"learning/internal/ai"
)

func TestClientParsesStructuredInsight(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || !strings.Contains(request.URL.Path, ":generateContent") {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"answer\":\"La variación requiere revisión.\",\"explanation\":\"El consumo supera la referencia y coincide con una alerta.\",\"suggestedQuestions\":[{\"question\":\"¿Qué reviso primero?\",\"answer\":\"El medidor.\"},{\"question\":\"¿Hay un evento asociado?\",\"answer\":\"Sí.\"},{\"question\":\"¿Debo escalarlo?\",\"answer\":\"Confirma la carga.\"}]}"}]}}]}`))
	}))
	defer server.Close()

	client := NewClient("test-key", "test-model")
	client.baseURL = server.URL + "/"
	insight, err := client.Analyze(context.Background(), ai.AnalysisContext{ReadingCount: 1, MeterCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	if insight.Answer == "" || len(insight.SuggestedQuestions) != 3 {
		t.Fatalf("unexpected insight: %#v", insight)
	}
}
