// Copyright 2021 Intuitive Labs GmbH. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE.txt file in the root of the source
// tree.

//+build nodebug

package slog

func init() {
	BuildTags = append(BuildTags, "nodebug")
}

// DBG() in nodebug mode (empty)

// DBGon returns true if logging at LDBG level is enabled.
func (l Log) DBGon() bool {
	return false
}

// DBG logs a message at the LDBG level  with a DBG prefix
// (shorthand for Log(LDBG, "DBG: ", ...))
func (l Log) DBG(f string, args ...interface{}) {
}
