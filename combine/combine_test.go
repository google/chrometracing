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

package combine

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/chrometracing/traceinternal"

	"github.com/google/go-cmp/cmp"
)

func TestLoadTrace(t *testing.T) {
	t.Run("Incomplete", func(t *testing.T) {
		const incompleteTrace = `[{"ph":"B"},`
		fn := filepath.Join(t.TempDir(), "incomplete.trace")
		if err := ioutil.WriteFile(fn, []byte(incompleteTrace), 0644); err != nil {
			t.Fatal(err)
		}
		events, err := loadTrace(fn)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(events), 1; got != want {
			t.Errorf("unexpected number of events loaded: got %d, want %d", got, want)
		}
	})

	t.Run("Complete", func(t *testing.T) {
		const completeTrace = `[{"ph":"B"}]`
		fn := filepath.Join(t.TempDir(), "complete.trace")
		if err := ioutil.WriteFile(fn, []byte(completeTrace), 0644); err != nil {
			t.Fatal(err)
		}
		events, err := loadTrace(fn)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(events), 1; got != want {
			t.Errorf("unexpected number of events loaded: got %d, want %d", got, want)
		}
	})
}

func writeEvents(t *testing.T, filename string, events []traceinternal.ViewerEvent) {
	t.Helper()

	b, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filename, b, 0644); err != nil {
		t.Fatal(err)
	}
}

const (
	begin = "B"
	end   = "E"
)

func span(name string, pid uint64, from, to float64) []traceinternal.ViewerEvent {
	return []traceinternal.ViewerEvent{
		{
			Name:  name,
			Phase: begin,
			Pid:   pid,
			Tid:   pid,
			Time:  from,
		},
		{
			Name:  name,
			Phase: end,
			Pid:   pid,
			Tid:   pid,
			Time:  to,
		},
	}

}

func TestCombine(t *testing.T) {
	startupEvents := span("startup", 1234, 0, 2000000)
	const embedStart = 2500000
	pidEvents := span("pid:5678", 1234, embedStart, 5000000)
	shutdownEvents := span("shutdown", 1234, 6000000, 7000000)

	topLevelEvents := append(append(append([]traceinternal.ViewerEvent{}, startupEvents...), pidEvents...), shutdownEvents...)
	topLevelFn := filepath.Join(t.TempDir(), "chrometracing.1234.trace")
	writeEvents(t, topLevelFn, topLevelEvents)

	childEvents := append(span("frobnicate", 5678, 0, 1000000),
		span("quxnicate", 5678, 1100000, 2000000)...)
	childEventsEmbedded := append(span("frobnicate", 5678, embedStart+0, embedStart+1000000),
		span("quxnicate", 5678, embedStart+1100000, embedStart+2000000)...)

	childFn := filepath.Join(t.TempDir(), "chrometracing.5678.trace")
	writeEvents(t, childFn, childEvents)

	var buf bytes.Buffer
	if err := Traces(&buf, []string{childFn, topLevelFn}); err != nil {
		t.Fatal(err)
	}

	var events []traceinternal.ViewerEvent
	if err := json.Unmarshal(buf.Bytes(), &events); err != nil {
		t.Fatal(err)
	}

	wantEvents := append(append([]traceinternal.ViewerEvent{
		startupEvents[0],
		startupEvents[1],

		pidEvents[0],
	}, childEventsEmbedded...),
		pidEvents[1],

		shutdownEvents[0],
		shutdownEvents[1])
	if diff := cmp.Diff(wantEvents, events); diff != "" {
		t.Errorf("unexpected combination: diff (-want +got):\n%s", diff)
	}
}
