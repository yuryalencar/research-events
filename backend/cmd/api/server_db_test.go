package main

// Spec: specs/backend/observability-opentelemetry.yaml
// Rule: "every GORM query gets a child span (via the GORM OTel tracing plugin)"
// Rule: "GORM plugin registered with tracing.WithoutQueryVariables() — required
//        to avoid leaking PII (emails etc.) into span attributes"

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/yuryalencar/research-events/internal/health"
)

// tracerForTest installs tp as the global OTel tracer provider for the
// duration of the test, then restores the previous one. The GORM tracing
// plugin (registered once in TestMain against the global provider) picks up
// tp via that global delegation — see testmain_test.go for why this works.
func tracerForTest(t *testing.T, tp *sdktrace.TracerProvider) {
	t.Helper()
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
}

func TestBuildHandler_DBQueryProducesChildSpanUnderHTTPSpan(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	tracerForTest(t, tp)

	h := BuildHandler(testConfig(), tx, health.NewRegistry(), discardLogger, tp)

	// GET /api/v1/events with no filters defaults to the current year and
	// reaches eventRepo.ListEvents — a real DB query.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	spans := exporter.GetSpans()
	require.GreaterOrEqual(t, len(spans), 2, "expected at least an HTTP root span and a DB query span")

	var httpSpan, dbSpan *tracetest.SpanStub
	for i := range spans {
		s := &spans[i]
		switch {
		case s.Name == "GET /api/v1/events":
			httpSpan = s
		case s.SpanKind == trace.SpanKindClient:
			dbSpan = s
		}
	}

	require.NotNil(t, httpSpan, "expected a root HTTP span named 'GET /api/v1/events'")
	require.NotNil(t, dbSpan, "expected a DB client span")
	assert.Equal(t, httpSpan.SpanContext.SpanID(), dbSpan.Parent.SpanID(),
		"DB span's parent should be the HTTP root span")
}

func TestBuildHandler_DBSpanDoesNotLeakRawQueryParameters(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	tracerForTest(t, tp)

	h := BuildHandler(testConfig(), tx, health.NewRegistry(), discardLogger, tp)

	const secretEmail = "secret-pii-tracer@example.com"
	body, err := json.Marshal(map[string]any{
		"name":        "Cycle 6 Test Conference",
		"slug":        "cycle6-test-conference",
		"country":     "Brazil",
		"city":        "Recife",
		"latitude":    -8.0476,
		"longitude":   -34.8770,
		"start_date":  "2027-09-21",
		"end_date":    "2027-09-25",
		"website_url": "https://example.com",
		"domain":      "computer_science",
		"submitter": map[string]string{
			"name":  "PII Test Submitter",
			"email": secretEmail,
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	for _, span := range exporter.GetSpans() {
		for _, attr := range span.Attributes {
			assert.NotContains(t, attr.Value.Emit(), secretEmail,
				"span attribute %s must not contain the raw submitter email", attr.Key)
		}
	}
}
