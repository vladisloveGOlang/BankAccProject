package federation

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

func NewSpan(ctx context.Context, name string) trace.Span {
	tracer, ok := ctx.Value("tracer").(trace.Tracer)
	if !ok {
		logrus.Warn("tracer not found in context")
		return nil
	}

	if tracer != nil {
		_, span := tracer.Start(ctx, name)
		return span
	}
	logrus.Warn("tracer is nil")

	return nil
}

func Span(span trace.Span) func() {
	return func() {
		if span != nil {
			span.End()
		}
	}
}
