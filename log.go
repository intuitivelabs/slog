// Copyright 2021 Intuitive Labs GmbH. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE.txt file in the root of the source
// tree.

// Package slog provides a very simple, fast level-based log.

package slog

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"time"
)

const Version = "0.1.1"

var BuildTags []string

const (
	logMaskLevel   = 0xff
	logMaskOpt     = 0xff00
	logMaskLogger  = 0xff0000
	logShiftOpt    = 8
	logShiftLogger = 16
)

// log output types
const (
	LDefaultOut LogOutput = iota
	LStdOut
	LStdErr
	LDisabledOut LogOutput = 9
)

// log levels
const (
	LMIN    LogLevel = -3
	LBUG    LogLevel = -3
	LCRIT   LogLevel = -2
	LERR    LogLevel = -1
	LWARN   LogLevel = 0
	LNOTICE LogLevel = 1
	LINFO   LogLevel = 2
	LDBG    LogLevel = 3
	LMAX    LogLevel = 3
)

// level names, keep in sync with the above Lxx constants
var levNames = [...]string{
	"BUG",
	"CRIT",
	"ERR",
	"WARN",
	"NOTICE",
	"INFO",
	"DBG",
}

// LevelName returns the log level name as string.
// On error it returns the empty string.
func LevelName(l LogLevel) string {
	if l < LMIN || l > LMAX {
		return ""
	}
	return levNames[l-LMIN]
}

// log options
const (
	LOptNone    LogOptions = 0
	LlocInfoS   LogOptions = 1 << 0 // short filenames
	LlocInfoL   LogOptions = 1 << 1 // long version (complete filenames)
	LtimeStamp  LogOptions = 1 << 2
	LbackTraceS LogOptions = 1 << 3 // short
	LbackTraceL LogOptions = 1 << 4 // long
)

var optNames = [...]string{
	"none",
	"location_short",
	"location_long",
	"timestamp",
	"backtrace_short",
	"backtrace_long",
}

// LogOptString returns the log options as string (seprated by "|").
// It returns the string representation of the options and a boolean
// specifying if all the options are known (true) or if an error was
// encountered (false)
func LogOptString(o LogOptions) (string, bool) {
	s := ""
	for i := 0; o != 0; i++ {
		if o&1 != 0 {
			if (i + 1) >= len(optNames) {
				return s, false
			}
			if len(s) == 0 {
				s = optNames[i+1]
			} else {
				s = s + "|" + optNames[i+1]
			}
		}
		o >>= 1
	}
	return s, true
}

// Log is a simple logger, keeping all its configuration inside
// an uint64.
// Format: LogOutput | LogOptions |  LogLevel (3 bytes used, 5 free)
type Log uint64

// LogLevel is the type used for the logging level.
type LogLevel int8

// LogOptions is the type used for the log options.
type LogOptions uint8

// LogOutput is the type used for the log output type.
type LogOutput uint8

// New returns a new log, intialised with the give options
func New(lev LogLevel, opt LogOptions, logger LogOutput) Log {
	var l Log
	Init(&l, lev, opt, logger)
	return l
}

// Init initialises a log with a given maximum log level, options and output
// type.
func Init(l *Log, lev LogLevel, opt LogOptions, logger LogOutput) {
	*l = (Log(logger) << logShiftLogger) |
		(Log(opt) << logShiftOpt) |
		(Log(uint8(lev)))
}

// SetLevel changes the log level.
func SetLevel(l *Log, lev LogLevel) {
	*l = (*l & ^Log(logMaskLevel)) | Log(uint8(lev))
}

// SetLevel changes the output type.
func SetOutput(l *Log, out LogOutput) {
	*l = (*l & ^Log(logMaskLogger)) | (Log(out) << logShiftLogger)
}

// SetOptions changes the log options.
func SetOptions(l *Log, opt LogOptions) {
	*l = (*l & ^Log(logMaskOpt)) | (Log(opt) << logShiftOpt)
}

// GetLevel returns the current log level.
func (l Log) GetLevel() LogLevel {
	return LogLevel(l & logMaskLevel)
}

// L returns true if the specified level is enabled for logging.
// Can be used to quickly check if logging is enabled, before calling
// a logging function (thus avoiding unneeded parameter evaluation).
func (l Log) L(lev LogLevel) bool {
	return l.GetLevel() >= lev
}

// GetOpts returns the current log options.
func (l Log) GetOpt() LogOptions {
	return LogOptions((l & logMaskOpt) >> logShiftOpt)
}

// GetLogger returns the current log output type.
func (l Log) GetLogger() LogOutput {
	return LogOutput((l & logMaskLogger) >> logShiftLogger)
}

// Log() logs directly without adding any log level prefix.
func (l Log) Log(lev LogLevel, f string, args ...interface{}) {
	l.LLog(lev, 1, "", f, args...)
}

// LogMux() will write the message both to the passed io.Writer and to the log.
// The message is written to the io.Writer independent of the log level and
// only if wcond is true. The message in the log will have added prefixes
// depending on the log options.
func (l Log) LogMux(w io.Writer, wcond bool,
	lev LogLevel, f string, args ...interface{}) {
	l.LLog(lev, 1, "", f, args...)
	if wcond && w != ioutil.Discard {
		//_, file, line, _ := runtime.Caller(1)
		//fmt.Fprintf(w, "MUXW:"+file+":"+strconv.Itoa(line)+":"+f, args...)
		fmt.Fprintf(w, f, args...)
	}
}

// return the last n elements of the path (separated by `\`)
//  the 0th element is the whole path
func getTrailingPath(path string, n int) string {
	var hits int
	if len(path) == 0 {
		return path
	}
	for i := len(path) - 1; i > 0; i-- {
		if path[i] == '/' {
			hits++
			if hits == n {
				path = path[i+1:] // get only the file
				break
			}
		}
	}
	return path
}

// LLog logs a message with a given prefix at a given level.
// It is supposed to be used mostly internally. The callersSkip parameter
// is used to find the caller for which to print the file/line information
// or where the backtrace should start.
func (l Log) LLog(lev LogLevel, callersSkip int, prefix string,
	f string, args ...interface{}) {
	if lev > l.GetLevel() {
		return
	}
	if l.GetOpt()&LtimeStamp != 0 {
		prefix = time.Now().Format("2006/01/02T15:04:05.99") + ":" + prefix
	}
	if l.GetOpt()&(LlocInfoS|LlocInfoL) != 0 {
		// add location info
		if _, file, line, ok := runtime.Caller(callersSkip + 1); ok {
			if l.GetOpt()&LlocInfoS != 0 {
				file = getTrailingPath(file, 2)
			}
			prefix += file + ":" + strconv.Itoa(line) + ": "
		}
	}
	if l.GetOpt()&(LbackTraceS|LbackTraceL) != 0 {
		var pc [10]uintptr
		n := runtime.Callers(callersSkip+2, pc[:])
		if n > 0 {
			frames := runtime.CallersFrames(pc[:n])
			prefix += "["
			for {
				frame, cont := frames.Next()
				function := frame.Function
				if l.GetOpt()&LbackTraceS != 0 {
					function = getTrailingPath(frame.Function, 1)
				}
				file := getTrailingPath(frame.File, 1)
				prefix += function + "(" + file + ":" +
					strconv.Itoa(frame.Line) + ")"
				if !cont {
					break
				}
				prefix += ":"
			}
			prefix += "]: "
		}
	}

	switch l.GetLogger() {
	default:
		fallthrough
	case LStdErr:
		fmt.Fprintf(os.Stderr, prefix+f, args...)
	case LStdOut:
		fmt.Fprintf(os.Stdout, prefix+f, args...)
	case LDisabledOut:
	}
}

// INFOon returns true if logging at LINFO level is enabled.
func (l Log) INFOon() bool {
	return l.L(LINFO)
}

// NOTICEon returns true if logging at LNOTICE level is enabled.
func (l Log) NOTICEon() bool {
	return l.L(LNOTICE)
}

// WARNon returns true if logging at LWARN level is enabled.
func (l Log) WARNon() bool {
	return l.L(LWARN)
}

// ERRon returns true if logging at LERR level is enabled.
func (l Log) ERRon() bool {
	return l.L(LERR)
}

// CRITon returns true if logging at LCRIT level is enabled.
func (l Log) CRITon() bool {
	return l.L(LCRIT)
}

// BUGon returns true if logging at LBUG level is enabled.
func (l Log) BUGon() bool {
	return l.L(LBUG)
}

// INFO logs a message at the LINFO level  with an INFO prefix
// (shorthand for Log(LINFO, "INFO: ", ...))
func (l Log) INFO(f string, args ...interface{}) {
	l.LLog(LINFO, 1, "INFO: ", f, args...)
}

// WARN logs a message at the LWARN level  with a WARN prefix
// (shorthand for Log(LWARN, "WARN: ", ...))
func (l Log) WARN(f string, args ...interface{}) {
	l.LLog(LWARN, 1, "WARNING: ", f, args...)
}

// ERR logs a message at the LERR level  with a ERROR prefix
// (shorthand for Log(LERR, "ERROR: ", ...))
func (l Log) ERR(f string, args ...interface{}) {
	l.LLog(LERR, 1, "ERROR: ", f, args...)
}

// BUG logs a message at the BUG level  with a BUG prefix.
func (l Log) BUG(f string, args ...interface{}) {
	l.LLog(LBUG, 1, "BUG: ", f, args...)
}

// PANIC logs a message at the BUG level and then calls panic()
func (l Log) PANIC(f string, args ...interface{}) {
	s := fmt.Sprintf(f, args...)
	l.LLog(LBUG, 1, "BUG: PANIC: ", "%s", s)
	panic(s)
}
