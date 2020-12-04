// this will be needed to get a tracer
import opentelemetry, { Tracer } from '@opentelemetry/api';
// and a Jaeger-specific exporter
import { JaegerExporter } from '@opentelemetry/exporter-jaeger';
// and an exporter with span processor
import { SimpleSpanProcessor } from '@opentelemetry/tracing';
// tracer provider for web
import { WebTracerProvider } from '@opentelemetry/web';
// Create a provider for activating and tracking spans
const tracerProvider = new WebTracerProvider();

let tracer: Tracer;

export default (): Tracer => {
  if (tracer) {
    return tracer;
  }

  let exporter = new JaegerExporter({
    serviceName: 'todo-web-frontend',
    host: 'localhost',
    port: 6831,
  });

  tracerProvider.addSpanProcessor(new SimpleSpanProcessor(exporter));
  tracerProvider.register();
  tracer = opentelemetry.trace.getTracer('todo-web-client');

  return tracer;
};
