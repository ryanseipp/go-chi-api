package otel

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Service interface {
	Shutdown(context.Context) error
}

type service struct {
	shutdownFuncs []func(context.Context) error
}

func New(ctx context.Context, serviceName string, version string) (serv Service, err error) {
	var shutdownFuncs []func(context.Context) error
	s := &service{shutdownFuncs: shutdownFuncs}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, s.Shutdown(ctx))
	}

	resource, err := initResource(ctx, serviceName, version)
	if err != nil {
		handleErr(err)
		return
	}

	tp, err := initTracerProvider(resource)
	if err != nil {
		handleErr(err)
		return
	}

	s.shutdownFuncs = append(s.shutdownFuncs, tp.Shutdown)

	mp, err := initMeterProvider(resource)
	if err != nil {
		handleErr(err)
		return
	}

	s.shutdownFuncs = append(s.shutdownFuncs, mp.Shutdown)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	serv = s
	return
}

func initResource(ctx context.Context, serviceName string, version string) (res *resource.Resource, err error) {
	resourceDetected, err := resource.New(
		ctx,
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return
	}

	return resource.Merge(
		resourceDetected,
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
		),
	)
}

func initTracerProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	return tp, nil
}

func initMeterProvider(res *resource.Resource) (*metric.MeterProvider, error) {
	exporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter)),
	)

	return mp, nil
}

func (s *service) Shutdown(ctx context.Context) (err error) {
	for _, fn := range s.shutdownFuncs {
		err = errors.Join(err, fn(ctx))
	}
	s.shutdownFuncs = nil
	return err
}
