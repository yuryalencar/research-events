package observability_test

// Spec: specs/backend/observability-opentelemetry.yaml
// Rule: "Sampling: ParentBased(TraceIDRatioBased(rate)), rate = 1.0 when cfg.Env ==
//        'development', else 0.2. Never 0."
// Rule: "SENTRY_DSN env var, optional. Empty -> no-op tracer, sentry.Init never called."

import (
	"reflect"
	"testing"

	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/yuryalencar/research-events/internal/config"
	"github.com/yuryalencar/research-events/internal/observability"
)

func TestTracesSampleRate_DevelopmentReturns1_0(t *testing.T) {
	rate := observability.TracesSampleRate("development")

	assert.Equal(t, 1.0, rate)
}

func TestTracesSampleRate_ProductionReturns0_2(t *testing.T) {
	rate := observability.TracesSampleRate("production")

	assert.Equal(t, 0.2, rate)
}

func TestTracesSampleRate_UnknownEnvReturns0_2(t *testing.T) {
	// Any non-"development" value (including typos/unset) must fall back to the
	// conservative production rate — sampling must never default to 0.
	rate := observability.TracesSampleRate("staging")

	assert.Equal(t, 0.2, rate)
}

func TestInitTracerProvider_EmptyDSN_ReturnsProviderWithoutCallingSentryInit(t *testing.T) {
	cfg := config.Config{Env: "development", SentryDSN: ""}

	tp, err := observability.InitTracerProvider(cfg)

	require.NoError(t, err)
	require.NotNil(t, tp)

	// sentry.Init() sets a client on the current hub. If InitTracerProvider
	// skipped Sentry entirely (empty DSN), the hub must still have no client.
	assert.Nil(t, sentry.CurrentHub().Client())
}

func TestInitTracerProvider_ValidDSN_InitializesSentryAndRegistersPropagator(t *testing.T) {
	// Syntactically valid but fake DSN — no network call is made by sentry.Init,
	// it only parses and validates the DSN format.
	cfg := config.Config{Env: "production", SentryDSN: "https://abc123@o0.ingest.de.sentry.io/1"}

	tp, err := observability.InitTracerProvider(cfg)

	require.NoError(t, err)
	require.NotNil(t, tp)

	// sentry.Init() binds a client to the current hub.
	assert.NotNil(t, sentry.CurrentHub().Client())

	// The global propagator must be the sentryotel propagator so trace context
	// (trace IDs, baggage) flows correctly between Sentry and OTel.
	wantType := reflect.TypeOf(sentryotel.NewSentryPropagator())
	gotType := reflect.TypeOf(otel.GetTextMapPropagator())
	assert.Equal(t, wantType, gotType)

	// TracesSampleRate must be non-zero. The sentryotel span processor starts
	// a *Sentry* transaction for every OTel root span with an "undefined"
	// explicit sampling decision (no incoming sentry-trace header), so
	// sentry-go falls back to ClientOptions.TracesSampleRate to decide
	// whether to keep it. If this is left at its zero value (0.0), sentry-go
	// drops every transaction itself — even though the OTel sampler already
	// decided to sample it — and nothing is ever sent to Sentry.
	assert.Equal(t, 1.0, sentry.CurrentHub().Client().Options().TracesSampleRate)
}

func TestInitTracerProvider_MalformedDSN_ReturnsError(t *testing.T) {
	cfg := config.Config{Env: "production", SentryDSN: "not-a-valid-dsn"}

	_, err := observability.InitTracerProvider(cfg)

	require.Error(t, err)
}
