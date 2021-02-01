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

// Package combine combines multiple chrome://tracing trace files into a
// whole-system view trace file.
package combine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/google/chrometracing/traceinternal"
)

func loadTrace(path string) ([]traceinternal.ViewerEvent, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	b = bytes.TrimSpace(b)

	// For parsing the JSON with the stricter encoding/json package, remove the
	// trailing comma and complete the array, if needed:
	b = bytes.TrimSuffix(b, []byte{','})
	if !bytes.HasSuffix(b, []byte{']'}) {
		b = append(b, ']')
	}

	var events []traceinternal.ViewerEvent
	if err := json.Unmarshal(b, &events); err != nil {
		return nil, err
	}
	return events, nil
}

type traceFiles struct {
	eventsForFn map[string][]traceinternal.ViewerEvent
	fnForPID    map[int]string
}

func (t *traceFiles) replacePIDWithEvents(root []traceinternal.ViewerEvent) ([]traceinternal.ViewerEvent, error) {
	replaced := make([]traceinternal.ViewerEvent, 0, len(root))
	for _, ev := range root {
		replaced = append(replaced, ev)
		if !strings.HasPrefix(ev.Name, "pid:") ||
			ev.Phase != "B" {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimPrefix(ev.Name, "pid:"))
		if err != nil {
			return nil, fmt.Errorf("BUG: pid: prefix followed by invalid pid: %q", ev.Name)
		}

		insertEvents := t.eventsForFn[t.fnForPID[pid]]
		for idx, iev := range insertEvents {
			iev.Time = ev.Time + iev.Time
			insertEvents[idx] = iev
		}
		insertEvents, err = t.replacePIDWithEvents(insertEvents)
		if err != nil {
			return nil, err
		}
		replaced = append(replaced, insertEvents...)
	}
	return replaced, nil
}

// Traces reads the specified chrome://tracing trace files and combines them
// into a whole-system view trace file.
func Traces(w io.Writer, filepaths []string) error {
	var (
		fnForPID    = make(map[int]string)
		eventsForFn = make(map[string][]traceinternal.ViewerEvent)

		highestTimestamp     float64
		highestTimestampFile string
	)
	for _, fn := range filepaths {
		events, err := loadTrace(fn)
		if err != nil {
			return err
		}
		eventsForFn[fn] = events
		for _, ev := range events {
			if ev.Time > highestTimestamp {
				highestTimestamp = ev.Time
				highestTimestampFile = fn
			}
		}

		parts := strings.Split(fn, ".")
		if parts[len(parts)-1] != "trace" {
			continue
		}
		pid, err := strconv.Atoi(parts[len(parts)-2])
		if err != nil {
			return err
		}
		fnForPID[pid] = fn
	}
	t := &traceFiles{
		eventsForFn: eventsForFn,
		fnForPID:    fnForPID,
	}
	root, err := t.replacePIDWithEvents(eventsForFn[highestTimestampFile])
	if err != nil {
		return err
	}
	b, err := json.Marshal(&root)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}
