// Copyright 2021 Intuitive Labs GmbH. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE.txt file in the root of the source
// tree.

// Package slog provides a very simple, fast level-based log.

package slog

import (
	"fmt"
)

// default package level log

var defaultLog Log

func init() {
	Init(&defaultLog, LMAX, 0, LStdOut)
}

func L(l LogLevel) bool {
	return defaultLog.L(l)
}

func DBGon() bool {
	return defaultLog.DBGon()
}

func INFOon() bool {
	return defaultLog.INFOon()
}

func NOTICEon() bool {
	return defaultLog.NOTICEon()
}

func WARNon() bool {
	return defaultLog.WARNon()
}

func ERRon() bool {
	return defaultLog.ERRon()
}

func CRITon() bool {
	return defaultLog.CRITon()
}

func DBG(f string, args ...interface{}) {
	defaultLog.LLog(LDBG, 1, "DBG: ", f, args...)
}

func INFO(f string, args ...interface{}) {
	defaultLog.LLog(LINFO, 1, "INFO: ", f, args...)
}

func NOTICE(f string, args ...interface{}) {
	defaultLog.LLog(LNOTICE, 1, "NOTICE: ", f, args...)
}

func WARN(f string, args ...interface{}) {
	defaultLog.LLog(LWARN, 1, "WARN: ", f, args...)
}

func ERR(f string, args ...interface{}) {
	defaultLog.LLog(LERR, 1, "ERROR: ", f, args...)
}

func CRIT(f string, args ...interface{}) {
	defaultLog.LLog(LCRIT, 1, "CRIT: ", f, args...)
}

func BUG(f string, args ...interface{}) {
	defaultLog.LLog(LBUG, 1, "BUG: ", f, args...)
}

func PANIC(f string, args ...interface{}) {
	s := fmt.Sprintf(f, args...)
	defaultLog.LLog(LBUG, 1, "BUG: PANIC: ", "%s", s)
	panic(s)
}

func DefaultLogSetLevel(lev LogLevel) {
	SetLevel(&defaultLog, lev)
}
func DefaultLogSetOutput(out LogOutput) {
	SetOutput(&defaultLog, out)
}

func DefaultLogSetOptions(opt LogOptions) {
	SetOptions(&defaultLog, opt)
}

func DefaultLogInit(lev LogLevel, opt LogOptions, logger LogOutput) {
	Init(&defaultLog, lev, opt, logger)
}
