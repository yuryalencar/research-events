package observability

// See docs/backend-libraries.md for what each library below does and why it
// was chosen (otel, sdktrace, sentry, sentryotel).
import (
	"fmt"

	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/yuryalencar/research-events/internal/config"
)

// TracesSampleRate returns the fraction of traces to sample for the given
// environment. Development samples everything (1.0) so every request is
// visible while developing; any other environment (production, staging,
// unknown/unset) samples conservatively at 0.2 to limit the volume of data
// sent to Sentry. Sampling must never be 0 — even in production we want a
// steady trickle of traces for diagnosing issues.
func TracesSampleRate(env string) float64 {
	if env == "development" {
		return 1.0
	}
	return 0.2
}

// InitTracerProvider builds the OTel TracerProvider for this service.
//
// A *sdktrace.TracerProvider is the SDK's factory for tracers/spans — it owns
// the sampler and the span processors that decide which spans are kept and
// where finished spans are sent.
//
// When cfg.SentryDSN is empty we still return a real provider (so otelhttp/
// GORM instrumentation work normally and spans are created), but with no
// span processor attached — spans are sampled and then dropped, and
// sentry.Init is never called.
//
// When cfg.SentryDSN is set, sentry.Init configures the Sentry client used
// for both error reporting and tracing, the sentryotel span processor
// forwards finished spans to Sentry, and the sentryotel propagator is
// registered globally so trace context survives across HTTP calls. A
// malformed DSN makes sentry.Init fail, and that error is returned to the
// caller (same as a bad DATABASE_URL — main.go should fatal-exit on it).
func InitTracerProvider(cfg config.Config) (*sdktrace.TracerProvider, error) {
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(TracesSampleRate(cfg.Env)))

	if cfg.SentryDSN == "" {
		return sdktrace.NewTracerProvider(sdktrace.WithSampler(sampler)), nil
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:           cfg.SentryDSN,
		Environment:   cfg.Env,
		Release:       cfg.AppVersion,
		EnableTracing: true,
		// TracesSampleRate must be 1.0, NOT TracesSampleRate(cfg.Env). The OTel
		// sampler (sdktrace.ParentBased(sdktrace.TraceIDRatioBased(...)) above)
		// is the single source of truth for which spans get created at all —
		// only sampled-in spans ever reach the sentryotel span processor. If
		// sentry-go's own TracesSampleRate were also < 1.0, it would apply a
		// SECOND, independent random sampling pass on top of that and drop
		// transactions the OTel sampler already decided to keep.
		TracesSampleRate: 1.0,
	}); err != nil {
		return nil, fmt.Errorf("observability: init sentry: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()),
	)

	otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())

	return tp, nil
}
