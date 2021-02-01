// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Tool combinetraces combines multiple chrome://tracing trace files into a
// whole-system view trace file.
//
// The tool identifies the top-level trace by looking at which trace file covers
// the longest time span. Then, it recursively replaces trace spans matching
// pid:<pid> (e.g. pid:1234) with the contents of the trace file for pid 1234
// (based on the file name).
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/chrometracing/combine"
)

func combineTraces() error {
	if flag.NArg() < 1 {
		return fmt.Errorf("syntax: %s <tracefile> [<tracefile>...]", filepath.Base(os.Args[0]))
	}
	return combine.Traces(os.Stdout, flag.Args())
}

func main() {
	flag.Parse()
	if err := combineTraces(); err != nil {
		log.Fatal(err)
	}
}
