package doctor

import (
	"flag"
	"github.com/tideland/goas/v3/logger"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"time"
)

type Doctor struct {
	startTime          time.Time
	writes             int
	memprofFile        string
	cpuprofFile        string
	lastSignal         *time.Time
	logLevel           logger.LogLevel
	enable             bool
	showGCStats        bool
	showMemStats       bool
	showGoroutineStats bool
}

func (this *Doctor) writeGCStats() {
	if !this.showGCStats {
		return
	}
	stats := new(debug.GCStats)
	debug.ReadGCStats(stats)
	logger.Infof("===GCStats======")
	logger.Infof("gc:last_gc %v", stats.LastGC)
	logger.Infof("gc:num_gc %v", stats.NumGC)
	logger.Infof("gc:pause_total %v", stats.PauseTotal)
	logger.Infof("gc:pause %v", stats.Pause)
	logger.Infof("gc:pause_quantiles %v", stats.PauseQuantiles)
}

func (this *Doctor) writeMemStats() {
	if !this.showMemStats {
		return
	}
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	logger.Infof("===memStats=====")
	logger.Infof("mem:alloc %v - %v", stats.Alloc, "bytes allocated and still in use")
	logger.Infof("mem:total_alloc %v - %v", stats.TotalAlloc, "bytes allocated (even if freed)")
	logger.Infof("mem:sys %v - %v", stats.Sys, "bytes obtained from system (sum of XxxSys below)")
	logger.Infof("mem:lookups %v - %v", stats.Lookups, "number of pointer lookups")
	logger.Infof("mem:mallocs %v - %v", stats.Mallocs, "number of mallocs")
	logger.Infof("mem:frees %v - %v", stats.Frees, "number of frees")
	logger.Infof("================")
	logger.Infof("mem:heap_alloc %v - %v", stats.HeapAlloc, "bytes allocated and still in use")
	logger.Infof("mem:heap_sys %v - %v", stats.HeapSys, "bytes obtained from system")
	logger.Infof("mem:heap_idle %v - %v", stats.HeapIdle, "bytes in idle spans")
	logger.Infof("mem:heap_in_use %v - %v", stats.HeapInuse, "bytes in non-idle spans")
	logger.Infof("mem:heap_released %v - %v", stats.HeapReleased, "bytes in released to the OS")
	logger.Infof("mem:heap_objects %v - %v", stats.HeapObjects, "total number of allocated objects")
	logger.Infof("================")
	logger.Infof("mem:next_gc %v -%v", stats.NextGC, "next run in HeapAlloc time (bytes)")
	logger.Infof("mem:last_gc %v - %v", stats.LastGC, "last run in absolute time (ns)")
	logger.Infof("mem:pause_total_ns %v", stats.PauseTotalNs)
	logger.Infof("mem:pause_ns %v - %v", stats.PauseNs, "circular buffer of recent GC pause times, most recent at [(NumGC+255)%256]")
	logger.Infof("mem:num_gc %v", stats.NumGC)
	logger.Infof("mem:enable_gc %v", stats.EnableGC)
	logger.Infof("mem:debug_gc %v", stats.DebugGC)
}

func (this *Doctor) writeGoroutineStats() {
	if !this.showGoroutineStats {
		return
	}
	logger.Infof("===GoroutineStats")
	logger.Infof("num_goroutine: %v", runtime.NumGoroutine())
	logger.Infof("================")
}

func (this *Doctor) processSignal(s chan os.Signal) {
	logger.Debugf("setting up interrupt signal")
	for {
		<-s
		logger.Infof("\n\n\nruntime: %v", time.Now().Sub(this.startTime))
		if this.lastSignal != nil && time.Now().Sub(*this.lastSignal) < 1e9 {
			os.Exit(2)
		}

		this.Update()

		logger.Infof("hit C-c again within a second to exit")

		signalTime := time.Now()

		this.lastSignal = &signalTime
	}
}

func (this *Doctor) Update() {
	this.writeGCStats()
	this.writeGoroutineStats()
	this.writeMemStats()

	if this.memprofFile != "" {
		f, err := os.Create(this.memprofFile)
		if err == nil {
			pprof.WriteHeapProfile(f)
			logger.Infof(">> writing heap profile to %v", this.memprofFile)
		} else {
			logger.Errorf("couldn't write to %v", this.memprofFile)
		}
	}

	if this.cpuprofFile != "" {
		logger.Infof(">> writing cpu samples to %v", this.cpuprofFile)
		pprof.StopCPUProfile()
		this.startCPUProfile()
	}
}

func (this *Doctor) startCPUProfile() {
	logger.Debugf("Starting CPUProfile...")
	this.writes = this.writes + 1
	f, err := os.Create(this.cpuprofFile + "_" + strconv.Itoa(this.writes))

	if err == nil {
		pprof.StartCPUProfile(f)
	} else {
		this.cpuprofFile = ""
		logger.Errorf("Error occured %v", err)
	}
}

func StartWithFlags() *Doctor {
	doctor := new(Doctor)

	enable := flag.Bool("doctor", false, "enable doctor")
	cpu := flag.String("cpu", "prof.cprof", "write cpu profile to this file")
	mem := flag.String("mem", "prof.mprof", "write mem profile to this file")
	statsgc := flag.Bool("statsgc", true, "show GC stats")
	statsmem := flag.Bool("statsmem", true, "show mem stats")
	statsgoroutine := flag.Bool("statsgoroutine", true, "show goroutine stats")

	flag.Parse()

	doctor.enable = *enable
	doctor.cpuprofFile = *cpu
	doctor.memprofFile = *mem
	doctor.showGCStats = *statsgc
	doctor.showMemStats = *statsmem
	doctor.showGoroutineStats = *statsgoroutine

	if !doctor.enable {
		return doctor
	}
	doctor.Start()

	return doctor
}

func (this *Doctor) Start() {
	logger.SetLevel(logger.LevelInfo)
	logger.Infof("Calling a Doctor...")

	this.startTime = time.Now()

	sig := make(chan os.Signal)
	go this.processSignal(sig)
	signal.Notify(sig, os.Interrupt)
	this.startCPUProfile()
}
