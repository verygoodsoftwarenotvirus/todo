package tracing

import (
	"context"
	"runtime"
	"strings"

	"go.opencensus.io/trace"
)

const (
	counterBuffer = 2

	this = "gitlab.com/verygoodsoftwarenotvirus/todo/"
)

// getFunctionCaller is inspired by/copied from https://stackoverflow.com/a/35213181
func getFunctionCaller() runtime.Frame {
	// We need the frame at index 3, since we never want runtime.Callers or getFunctionCaller or StartSpan itself
	targetFrameIndex := 3

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+counterBuffer)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])

		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			if frameIndex == targetFrameIndex {
				var frameCandidate runtime.Frame
				frameCandidate, more = frames.Next()
				frame = frameCandidate
			} else {
				_, more = frames.Next()
			}
		}
	}

	return frame
}

// StartSpan starts a span.
func StartSpan(ctx context.Context) (context.Context, *trace.Span) {
	funcName := strings.TrimPrefix(getFunctionCaller().Function, this)

	return trace.StartSpan(ctx, funcName)
}
