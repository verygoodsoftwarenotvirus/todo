package tracing

import (
	"context"
	"runtime"
	"strings"

	"go.opencensus.io/trace"
)

const (
	// We need the frame at index 3, since we never want runtime.Callers or getFunctionCaller or StartSpan itself.
	runtimeFrameBuffer = 3
	counterBuffer      = 2

	this = "gitlab.com/verygoodsoftwarenotvirus/todo/"
)

// getFunctionCaller is inspired by/copied from https://stackoverflow.com/a/35213181
func getFunctionCaller() string {
	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, runtimeFrameBuffer+counterBuffer)
	n := runtime.Callers(0, programCounters)
	frame := runtime.Frame{Function: "unknown"}

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])

		for more, frameIndex := true, 0; more && frameIndex <= runtimeFrameBuffer; frameIndex++ {
			if frameIndex == runtimeFrameBuffer {
				frame, more = frames.Next()
			} else {
				_, more = frames.Next()
			}
		}
	}

	return frame.Function
}

// StartSpan starts a span.
func StartSpan(ctx context.Context) (context.Context, *trace.Span) {
	funcName := strings.TrimPrefix(getFunctionCaller(), this)

	return trace.StartSpan(ctx, funcName)
}
