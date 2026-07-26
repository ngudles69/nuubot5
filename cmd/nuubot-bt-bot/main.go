package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"time"

	"nuubot/internal/btbot"
	"nuubot/internal/report"
	"nuubot/internal/toolkit/logging"
)

const program = "nuubot-bt-bot"

type performanceProfile struct {
	prefix                string
	cpuFile               *os.File
	traceFile             *os.File
	previousMutexFraction int
	cpuStarted            bool
	traceStarted          bool
	started               bool
	stopped               bool
}

// Section 1 - Program Flow

func main() {
	var started = time.Now()

	// open server.log
	var log, err = logging.Open(logging.ServerLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to open log file:", err)
		os.Exit(1)
	}

	// parse input
	var sweepID, botID uint64
	var profilePrefix string
	sweepID, botID, profilePrefix, err = parseInput(os.Args[1:])
	if err != nil {
		log.Error(fmt.Sprintf("parseInput() failed: %v", err))
		os.Exit(1)
	}

	// open bot_<identity>.log
	var botLog *logging.Logger
	botLog, err = logging.OpenBotLog(sweepID, botID)
	if err != nil {
		log.Error(fmt.Sprintf("logging.OpenBotLog() failed: %v", err))
		os.Exit(1)
	}
	log = botLog

	// start performance profile
	var profile = performanceProfile{prefix: profilePrefix}
	err = profile.Start()
	if err != nil {
		log.Error(fmt.Sprintf("performanceProfile.Start() failed: %v", err))
		os.Exit(1)
	}

	// create btbot
	var runner btbot.BtBot

	// initialize btbot
	err = runner.Init(context.Background(), log, sweepID, botID)
	if err != nil {
		err = errors.Join(err, profile.Stop())
		log.Error(fmt.Sprintf("btbot.Init() failed: %v", err))
		os.Exit(1)
	}

	// start btbot
	err = runner.Start()
	if err != nil {
		err = errors.Join(err, runner.Stop(), profile.Stop())
		log.Error(fmt.Sprintf("btbot.Start() failed: %v", err))
		os.Exit(1)
	}

	// loop btbot
	err = runner.Loop()

	// stop btbot
	var stopErr = runner.Stop()
	if err != nil {
		err = errors.Join(err, stopErr, profile.Stop())
		log.Error(fmt.Sprintf("btbot.Loop() failed: %v", err))
		os.Exit(1)
	}
	if stopErr != nil {
		stopErr = errors.Join(stopErr, profile.Stop())
		log.Error(fmt.Sprintf("btbot.Stop() failed: %v", stopErr))
		os.Exit(1)
	}

	// get result
	var result btbot.Result
	result, err = runner.Result()
	if err != nil {
		err = errors.Join(err, profile.Stop())
		log.Error(fmt.Sprintf("btbot.Result() failed: %v", err))
		os.Exit(1)
	}

	// write run report
	err = report.WriteRunJSON(os.Stdout, result.Report)
	if err != nil {
		err = errors.Join(err, profile.Stop())
		log.Error(fmt.Sprintf("report.WriteRunJSON() failed: %v", err))
		os.Exit(1)
	}

	// stop performance profile
	err = profile.Stop()
	if err != nil {
		log.Error(fmt.Sprintf("performanceProfile.Stop() failed: %v", err))
		os.Exit(1)
	}

	// log result
	log.Info(fmt.Sprintf("btbot completed successfully in %s", time.Since(started)))
}

// Section 2 - Domain Helpers

func parseInput(args []string) (uint64, uint64, string, error) {
	if len(args) != 2 && len(args) != 4 {
		return 0, 0, "", fmt.Errorf("usage: %s <sweep_id> <bot_id> [-pp profile_prefix]", program)
	}

	// parse sweep id
	var sweepID, err = positiveID(args[0])
	if err != nil {
		return 0, 0, "", fmt.Errorf("parse sweep id: %w", err)
	}

	// parse bot id
	var botID uint64
	botID, err = positiveID(args[1])
	if err != nil {
		return 0, 0, "", fmt.Errorf("parse bot id: %w", err)
	}

	// parse profile prefix
	var profilePrefix string
	if len(args) == 4 {
		if args[2] != "-pp" {
			return 0, 0, "", fmt.Errorf("invalid performance profile flag: %s", args[2])
		}
		profilePrefix = args[3]
		if profilePrefix == "" {
			return 0, 0, "", fmt.Errorf("profile prefix is empty")
		}
	}
	return sweepID, botID, profilePrefix, nil
}

func positiveID(value string) (uint64, error) {
	var id, err = strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid positive id: %s", value)
	}
	return id, nil
}

// Section 3 - Generic Helpers

func (p *performanceProfile) Start() error {
	if p.prefix == "" {
		return nil
	}
	if p.started || p.stopped {
		return fmt.Errorf("performance profile cannot start from current state")
	}

	var err error
	p.cpuFile, err = os.Create(p.prefix + ".cpu.pprof")
	if err != nil {
		return fmt.Errorf("create CPU profile: %w", err)
	}
	p.traceFile, err = os.Create(p.prefix + ".trace")
	if err != nil {
		return errors.Join(fmt.Errorf("create runtime trace: %w", err), p.cpuFile.Close())
	}

	p.previousMutexFraction = runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	err = pprof.StartCPUProfile(p.cpuFile)
	if err != nil {
		runtime.SetBlockProfileRate(0)
		runtime.SetMutexProfileFraction(p.previousMutexFraction)
		return errors.Join(fmt.Errorf("start CPU profile: %w", err), p.traceFile.Close(), p.cpuFile.Close())
	}
	p.cpuStarted = true
	err = trace.Start(p.traceFile)
	if err != nil {
		pprof.StopCPUProfile()
		p.cpuStarted = false
		runtime.SetBlockProfileRate(0)
		runtime.SetMutexProfileFraction(p.previousMutexFraction)
		return errors.Join(fmt.Errorf("start runtime trace: %w", err), p.traceFile.Close(), p.cpuFile.Close())
	}
	p.traceStarted = true
	p.started = true
	return nil
}

func (p *performanceProfile) Stop() error {
	if p.prefix == "" || p.stopped {
		return nil
	}
	if !p.started {
		return fmt.Errorf("performance profile is not started")
	}
	p.stopped = true

	if p.traceStarted {
		trace.Stop()
		p.traceStarted = false
	}
	if p.cpuStarted {
		pprof.StopCPUProfile()
		p.cpuStarted = false
	}
	var err = errors.Join(p.traceFile.Close(), p.cpuFile.Close())
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(p.previousMutexFraction)
	runtime.GC()
	for _, name := range []string{"heap", "allocs", "block", "mutex"} {
		var writeErr = writeRuntimeProfile(p.prefix, name)
		err = errors.Join(err, writeErr)
	}
	return err
}

func writeRuntimeProfile(prefix string, name string) error {
	var profile = pprof.Lookup(name)
	if profile == nil {
		return fmt.Errorf("runtime profile %q is unavailable", name)
	}
	var file, err = os.Create(prefix + "." + name + ".pprof")
	if err != nil {
		return fmt.Errorf("create %s profile: %w", name, err)
	}
	var writeErr = profile.WriteTo(file, 0)
	var closeErr = file.Close()
	if writeErr != nil {
		return errors.Join(fmt.Errorf("write %s profile: %w", name, writeErr), closeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close %s profile: %w", name, closeErr)
	}
	return nil
}
