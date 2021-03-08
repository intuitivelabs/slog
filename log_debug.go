// Copyright 2021 Intuitive Labs GmbH. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE.txt file in the root of the source
// tree.

//+build default debug !nodebug

package slog

func init() {
	BuildTags = append(BuildTags, "debug")
}

// DBG() in default debugging mode

// DBGon returns true if logging at LDBG level is enabled.
func (l Log) DBGon() bool {
	return l.L(LDBG)
}

// DBG logs a message at the LDBG level  with a DBG prefix
// (shorthand for Log(LDBG, "DBG: ", ...))
func (l Log) DBG(fmt string, args ...interface{}) {
	l.LLog(LDBG, 1, "DBG: ", fmt, args...)
}
