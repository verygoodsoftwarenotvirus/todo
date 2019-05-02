package opencensus

// lovingly borrowed from https://github.com/opencensus-integrations/caddy/blob/c8498719b7c1c2a3c707355be2395a35f03e434e/caddy/caddymain/exporters.go#L54-L110

import (
	"context"
	"runtime"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	pauseTimeDistribution = view.Distribution(
		0, 1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
	)

	memoryDistribution = view.Distribution(0,
		1<<10, 10*1<<10, 100*1<<10,
		1<<20, 10*1<<20, 100*1<<20,
		1<<30, 10*1<<30, 100*1<<30,
		1<<40, 10*1<<40, 100*1<<40,
		1<<50, 10*1<<50, 100*1<<50,
		1<<60, 10*1<<60, 1<<64-1,
	)

	// MetricAggregationMeasurement keeps track of how much time we spend collecting metrics
	MetricAggregationMeasurement = stats.Int64("metrics_aggregation_time", "cumulative time in nanoseconds spent aggregating metrics", stats.UnitDimensionless)
	// MetricAggregationMeasurementView is the corresponding view for the above metric
	MetricAggregationMeasurementView = &view.View{
		Name:        "metrics_aggregation_time",
		Measure:     MetricAggregationMeasurement,
		Description: "cumulative time in nanoseconds spent aggregating metrics",
		Aggregation: view.Count(),
	}

	// RuntimeTotalAllocMeasurement captures the runtime memstats TotalAlloc field
	RuntimeTotalAllocMeasurement = stats.Int64("total_alloc", "cumulative bytes allocated for heap objects", stats.UnitDimensionless)
	// RuntimeTotalAllocView is the corresponding view for the above field
	RuntimeTotalAllocView = &view.View{
		Name:        "total_alloc",
		Measure:     RuntimeTotalAllocMeasurement,
		Description: "cumulative bytes allocated for heap objects",
		Aggregation: view.Count(),
	}

	// RuntimeSysMeasurement captures the runtime memstats Sys field
	RuntimeSysMeasurement = stats.Int64("sys", "total bytes of memory obtained from the OS", stats.UnitDimensionless)
	// RuntimeSysView is the corresponding view for the above field
	RuntimeSysView = &view.View{
		Name:        "sys",
		Measure:     RuntimeSysMeasurement,
		Description: "total bytes of memory obtained from the OS",
		Aggregation: view.Count(),
	}

	// RuntimeLookupsMeasurement captures the runtime memstats Lookups field
	RuntimeLookupsMeasurement = stats.Int64("lookups", "the number of pointer lookups performed by the runtime", stats.UnitDimensionless)
	// RuntimeLookupsView is the corresponding view for the above field
	RuntimeLookupsView = &view.View{
		Name:        "lookups",
		Measure:     RuntimeLookupsMeasurement,
		Description: "the number of pointer lookups performed by the runtime",
		Aggregation: view.Count(),
	}

	// RuntimeMallocsMeasurement captures the runtime memstats Mallocs field
	RuntimeMallocsMeasurement = stats.Int64("mallocs", "the cumulative count of heap objects allocated (the number of live objects is mallocs - frees)", stats.UnitDimensionless)
	// RuntimeNallocsView is the corresponding view for the above field
	RuntimeNallocsView = &view.View{
		Name:        "mallocs",
		Measure:     RuntimeMallocsMeasurement,
		Description: "the cumulative count of heap objects allocated (the number of live objects is mallocs - frees)",
		Aggregation: view.Count(),
	}

	// RuntimeFreesMeasurement captures the runtime memstats Frees field
	RuntimeFreesMeasurement = stats.Int64("frees", "cumulative count of heap objects freed (the number of live objects is mallocs - frees)", stats.UnitDimensionless)
	// RuntimeFreesView is the corresponding view for the above field
	RuntimeFreesView = &view.View{
		Name:        "frees",
		Measure:     RuntimeFreesMeasurement,
		Description: "cumulative count of heap objects freed (the number of live objects is mallocs - frees)",
		Aggregation: view.Count(),
	}

	// RuntimeHeapAllocMeasurement captures the runtime memstats HeapAlloc field
	RuntimeHeapAllocMeasurement = stats.Int64("heap_alloc", "bytes of allocated heap objects", stats.UnitDimensionless)
	// RuntimeHeapAllocView is the corresponding view for the above field
	RuntimeHeapAllocView = &view.View{
		Name:        "heap_alloc",
		Measure:     RuntimeHeapAllocMeasurement,
		Description: "bytes of allocated heap objects",
		Aggregation: view.Count(),
	}

	// RuntimeHeapSysMeasurement captures the runtime memstats HeapSys field
	RuntimeHeapSysMeasurement = stats.Int64("heap_sys", "bytes of heap memory obtained from the OS", stats.UnitDimensionless)
	// RuntimeHeapSysView is the corresponding view for the above field
	RuntimeHeapSysView = &view.View{
		Name:        "heap_sys",
		Measure:     RuntimeHeapSysMeasurement,
		Description: "bytes of heap memory obtained from the OS",
		Aggregation: view.Count(),
	}

	// RuntimeHeapIdleMeasurement captures the runtime memstats HeapIdle field
	RuntimeHeapIdleMeasurement = stats.Int64("heap_idle", "bytes in idle (unused) spans", stats.UnitDimensionless)
	// RuntimeHeapIdleView is the corresponding view for the above field
	RuntimeHeapIdleView = &view.View{
		Name:        "heap_idle",
		Measure:     RuntimeHeapIdleMeasurement,
		Description: "bytes in idle (unused) spans",
		Aggregation: view.Count(),
	}

	// RuntimeHeapInuseMeasurement captures the runtime memstats HeapInuse field
	RuntimeHeapInuseMeasurement = stats.Int64("heap_inuse", "bytes in in-use spans", stats.UnitDimensionless)
	// RuntimeHeapInuseView is the corresponding view for the above field
	RuntimeHeapInuseView = &view.View{
		Name:        "heap_inuse",
		Measure:     RuntimeHeapInuseMeasurement,
		Description: "bytes in in-use spans",
		Aggregation: view.Count(),
	}

	// RuntimeHeapReleasedMeasurement captures the runtime memstats HeapReleased field
	RuntimeHeapReleasedMeasurement = stats.Int64("heap_released", "bytes of physical memory returned to the OS", stats.UnitDimensionless)
	// RuntimeHeapReleasedView is the corresponding view for the above field
	RuntimeHeapReleasedView = &view.View{
		Name:        "heap_released",
		Measure:     RuntimeHeapReleasedMeasurement,
		Description: "bytes of physical memory returned to the OS",
		Aggregation: view.Count(),
	}

	// RuntimeHeapObjectsMeasurement captures the runtime memstats HeapObjects field
	RuntimeHeapObjectsMeasurement = stats.Int64("heap_objects", "the number of allocated heap objects.", stats.UnitDimensionless)
	// RuntimeHeapObjectsView is the corresponding view for the above field
	RuntimeHeapObjectsView = &view.View{
		Name:        "heap_objects",
		Measure:     RuntimeHeapObjectsMeasurement,
		Description: "the number of allocated heap objects.",
		Aggregation: view.Count(),
	}

	// RuntimeStackInuseMeasurement captures the runtime memstats StackInuse field
	RuntimeStackInuseMeasurement = stats.Int64("stack_inuse", "bytes in stack spans.", stats.UnitDimensionless)
	// RuntimeStackInuseView is the corresponding view for the above field
	RuntimeStackInuseView = &view.View{
		Name:        "stack_inuse",
		Measure:     RuntimeStackInuseMeasurement,
		Description: "bytes in stack spans.",
		Aggregation: view.Count(),
	}

	// RuntimeStackSysMeasurement captures the runtime memstats StackSys field
	RuntimeStackSysMeasurement = stats.Int64("stack_sys", "bytes of stack memory obtained from the OS.", stats.UnitDimensionless)
	// RuntimeStackSysView is the corresponding view for the above field
	RuntimeStackSysView = &view.View{
		Name:        "stack_sys",
		Measure:     RuntimeStackSysMeasurement,
		Description: "bytes of stack memory obtained from the OS.",
		Aggregation: view.Count(),
	}

	// RuntimeMSpanInuseMeasurement captures the runtime memstats mSpanInuse field
	RuntimeMSpanInuseMeasurement = stats.Int64("mspan_inuse", "bytes of allocated mspan structures.", stats.UnitDimensionless)
	// RuntimemSpanInuseView is the corresponding view for the above field
	RuntimemSpanInuseView = &view.View{
		Name:        "mspan_inuse",
		Measure:     RuntimeMSpanInuseMeasurement,
		Description: "bytes of allocated mspan structures.",
		Aggregation: view.Count(),
	}

	// RuntimeMSpanSysMeasurement captures the runtime memstats mSpanSys field
	RuntimeMSpanSysMeasurement = stats.Int64("mspan_sys", "bytes of memory obtained from the OS for mspan structures.", stats.UnitDimensionless)
	// RuntimemSpanSysView is the corresponding view for the above field
	RuntimemSpanSysView = &view.View{
		Name:        "mspan_sys",
		Measure:     RuntimeMSpanSysMeasurement,
		Description: "bytes of memory obtained from the OS for mspan structures.",
		Aggregation: view.Count(),
	}

	// RuntimeMCacheInuseMeasurement captures the runtime memstats MCacheInuse field
	RuntimeMCacheInuseMeasurement = stats.Int64("mcache_inuse", "bytes of allocated mcache structures.", stats.UnitDimensionless)
	// RuntimeMCacheInuseView is the corresponding view for the above field
	RuntimeMCacheInuseView = &view.View{
		Name:        "mcache_inuse",
		Measure:     RuntimeMCacheInuseMeasurement,
		Description: "bytes of allocated mcache structures.",
		Aggregation: view.Count(),
	}

	// RuntimeMCacheSysMeasurement captures the runtime memstats MCacheSys field
	RuntimeMCacheSysMeasurement = stats.Int64("mcache_sys", "bytes of memory obtained from the OS for mcache structures.", stats.UnitDimensionless)
	// RuntimeMCacheSysView is the corresponding view for the above field
	RuntimeMCacheSysView = &view.View{
		Name:        "mcache_sys",
		Measure:     RuntimeMCacheSysMeasurement,
		Description: "bytes of memory obtained from the OS for mcache structures.",
		Aggregation: view.Count(),
	}

	// RuntimeBuckHashSysMeasurement captures the runtime memstats BuckHashSys field
	RuntimeBuckHashSysMeasurement = stats.Int64("buck_hash_sys", "bytes of memory in profiling bucket hash tables.", stats.UnitDimensionless)
	// RuntimeBuckHashSysView is the corresponding view for the above field
	RuntimeBuckHashSysView = &view.View{
		Name:        "buck_hash_sys",
		Measure:     RuntimeBuckHashSysMeasurement,
		Description: "bytes of memory in profiling bucket hash tables.",
		Aggregation: view.Count(),
	}

	// RuntimeGCSysMeasurement captures the runtime memstats GCSys field
	RuntimeGCSysMeasurement = stats.Int64("gc_sys", "bytes of memory in garbage collection metadata.", stats.UnitDimensionless)
	// RuntimeGCSysView is the corresponding view for the above field
	RuntimeGCSysView = &view.View{
		Name:        "gc_sys",
		Measure:     RuntimeGCSysMeasurement,
		Description: "bytes of memory in garbage collection metadata.",
		Aggregation: view.Count(),
	}

	// RuntimeOtherSysMeasurement captures the runtime memstats OtherSys field
	RuntimeOtherSysMeasurement = stats.Int64("other_sys", "bytes of memory in miscellaneous off-heap runtime allocations.", stats.UnitDimensionless)
	// RuntimeOtherSysView is the corresponding view for the above field
	RuntimeOtherSysView = &view.View{
		Name:        "other_sys",
		Measure:     RuntimeOtherSysMeasurement,
		Description: "bytes of memory in miscellaneous off-heap runtime allocations.",
		Aggregation: view.Count(),
	}

	// RuntimeNextGCMeasurement captures the runtime memstats NextGC field
	RuntimeNextGCMeasurement = stats.Int64("next_gc", "the target heap size of the next GC cycle.", stats.UnitDimensionless)
	// RuntimeNextGCView is the corresponding view for the above field
	RuntimeNextGCView = &view.View{
		Name:        "next_gc",
		Measure:     RuntimeNextGCMeasurement,
		Description: "the target heap size of the next GC cycle.",
		Aggregation: view.Count(),
	}

	// RuntimePauseTotalNsMeasurement captures the runtime memstats PauseTotalNs field
	RuntimePauseTotalNsMeasurement = stats.Int64("pause_total_ns", "the cumulative nanoseconds in GC", stats.UnitDimensionless)
	// RuntimePauseTotalNsView is the corresponding view for the above field
	RuntimePauseTotalNsView = &view.View{
		Name:        "pause_total_ns",
		Measure:     RuntimePauseTotalNsMeasurement,
		Description: "the cumulative nanoseconds in GC",
		Aggregation: view.Count(),
	}

	// RuntimePauseNsMeasurement captures the runtime memstats PauseNs field
	RuntimePauseNsMeasurement = stats.Int64("pause_ns", "a circular buffer of recent GC stop-the-world pause times in nanoseconds (the most recent pause is at PauseNs[(NumGC+255)%256])", stats.UnitDimensionless)
	// RuntimePauseNsView is the corresponding view for the above field
	RuntimePauseNsView = &view.View{
		Name:        "pause_ns",
		Measure:     RuntimePauseNsMeasurement,
		Description: "a circular buffer of recent GC stop-the-world pause times in nanoseconds (the most recent pause is at PauseNs[(NumGC+255)%256])",
		Aggregation: pauseTimeDistribution,
	}

	// RuntimePauseEndMeasurement captures the runtime memstats PauseEnd field
	RuntimePauseEndMeasurement = stats.Int64("pause_end", "a circular buffer of recent GC pause end times, as nanoseconds since 1970 (the UNIX epoch).", stats.UnitDimensionless)
	// RuntimePauseEndView is the corresponding view for the above field
	RuntimePauseEndView = &view.View{
		Name:        "pause_end",
		Measure:     RuntimePauseEndMeasurement,
		Description: "a circular buffer of recent GC pause end times, as nanoseconds since 1970 (the UNIX epoch).",
		Aggregation: view.Count(),
	}

	// RuntimeNumGCMeasurement captures the runtime memstats NumGC field
	RuntimeNumGCMeasurement = stats.Int64("num_gc", "the number of completed GC cycles.", stats.UnitDimensionless)
	// RuntimeNumGCView is the corresponding view for the above field
	RuntimeNumGCView = &view.View{
		Name:        "num_gc",
		Measure:     RuntimeNumGCMeasurement,
		Description: "the number of completed GC cycles.",
		Aggregation: view.Count(),
	}

	// RuntimeNumForcedGCMeasurement captures the runtime memstats NumForcedGC field
	RuntimeNumForcedGCMeasurement = stats.Int64("num_forced_gc", "the number of GC cycles that were forced by the application calling the GC function.", stats.UnitDimensionless)
	// RuntimeNumForcedGCView is the corresponding view for the above field
	RuntimeNumForcedGCView = &view.View{
		Name:        "num_forced_gc",
		Measure:     RuntimeNumForcedGCMeasurement,
		Description: "the number of GC cycles that were forced by the application calling the GC function.",
		Aggregation: view.Count(),
	}

	// RuntimeGCCPUFractionMeasurement captures the runtime memstats GCCPUFraction field
	RuntimeGCCPUFractionMeasurement = stats.Float64("gc_cpu_fraction", "the fraction of this program's available CPU time used by the GC since the program started.", stats.UnitDimensionless)
	// RuntimeGCCPUFractionView is the corresponding view for the above field
	RuntimeGCCPUFractionView = &view.View{
		Name:        "gc_cpu_fraction",
		Measure:     RuntimeGCCPUFractionMeasurement,
		Description: "the fraction of this program's available CPU time used by the GC since the program started.",
		Aggregation: view.Count(),
	}

	// DefaultRuntimeViews represents the pre-configured views
	DefaultRuntimeViews = []*view.View{
		RuntimeTotalAllocView,
		RuntimeSysView,
		RuntimeLookupsView,
		RuntimeNallocsView,
		RuntimeFreesView,
		RuntimeHeapAllocView,
		RuntimeHeapSysView,
		RuntimeHeapIdleView,
		RuntimeHeapInuseView,
		RuntimeHeapReleasedView,
		RuntimeHeapObjectsView,
		RuntimeStackInuseView,
		RuntimeStackSysView,
		RuntimemSpanInuseView,
		RuntimemSpanSysView,
		RuntimeMCacheInuseView,
		RuntimeMCacheSysView,
		RuntimeBuckHashSysView,
		RuntimeGCSysView,
		RuntimeOtherSysView,
		RuntimeNextGCView,
		RuntimePauseTotalNsView,
		RuntimePauseNsView,
		RuntimePauseEndView,
		RuntimeNumGCView,
		RuntimeNumForcedGCView,
		RuntimeGCCPUFractionView,
		MetricAggregationMeasurementView,
	}
)

// RegisterViews registers default runtime views
func RegisterViews() error {
	for _, v := range DefaultRuntimeViews {
		if err := view.Register(v); err != nil {
			return err
		}
	}
	return nil
}

// RecordStats records runtime statistics at the provided interval.
func RecordStats(interval time.Duration) (fnStop func()) {
	var (
		closeOnce sync.Once
		ctx       = context.Background()
		ticker    = time.NewTicker(interval)
		done      = make(chan struct{})
	)

	ms := &runtime.MemStats{}
	go func() {
		for {
			select {
			case <-ticker.C:
				startTime := time.Now()
				runtime.ReadMemStats(ms)
				stats.Record(
					ctx,
					RuntimeTotalAllocMeasurement.M(int64(ms.TotalAlloc)),
					RuntimeSysMeasurement.M(int64(ms.Sys)),
					RuntimeLookupsMeasurement.M(int64(ms.Lookups)),
					RuntimeMallocsMeasurement.M(int64(ms.Mallocs)),
					RuntimeFreesMeasurement.M(int64(ms.Frees)),
					RuntimeHeapAllocMeasurement.M(int64(ms.HeapAlloc)),
					RuntimeHeapSysMeasurement.M(int64(ms.HeapSys)),
					RuntimeHeapIdleMeasurement.M(int64(ms.HeapIdle)),
					RuntimeHeapInuseMeasurement.M(int64(ms.HeapInuse)),
					RuntimeHeapReleasedMeasurement.M(int64(ms.HeapReleased)),
					RuntimeHeapObjectsMeasurement.M(int64(ms.HeapObjects)),
					RuntimeStackInuseMeasurement.M(int64(ms.StackInuse)),
					RuntimeStackSysMeasurement.M(int64(ms.StackSys)),
					RuntimeMSpanInuseMeasurement.M(int64(ms.MSpanInuse)),
					RuntimeMSpanSysMeasurement.M(int64(ms.MSpanSys)),
					RuntimeMCacheInuseMeasurement.M(int64(ms.MCacheInuse)),
					RuntimeMCacheSysMeasurement.M(int64(ms.MCacheSys)),
					RuntimeBuckHashSysMeasurement.M(int64(ms.BuckHashSys)),
					RuntimeGCSysMeasurement.M(int64(ms.GCSys)),
					RuntimeOtherSysMeasurement.M(int64(ms.OtherSys)),
					RuntimeNextGCMeasurement.M(int64(ms.NextGC)),
					RuntimePauseTotalNsMeasurement.M(int64(ms.PauseTotalNs)),
					RuntimePauseNsMeasurement.M(int64(ms.PauseNs[(ms.NumGC+255)%256])),
					RuntimePauseEndMeasurement.M(int64(ms.PauseEnd[(ms.NumGC+255)%256])),
					RuntimeNumGCMeasurement.M(int64(ms.NumGC)),
					RuntimeNumForcedGCMeasurement.M(int64(ms.NumForcedGC)),
					RuntimeGCCPUFractionMeasurement.M(ms.GCCPUFraction),
				)
				stats.Record(ctx, MetricAggregationMeasurement.M(time.Since(startTime).Nanoseconds()))
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		closeOnce.Do(func() {
			close(done)
		})
	}
}
