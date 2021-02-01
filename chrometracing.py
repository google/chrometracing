# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# -*- coding: utf-8 -*-
"""Writes per-process Chrome trace_event files, for chrome://tracing."""

import json
import os
import os.path
import sys
import time


def _open_trace_file():
  """Opens the per-process trace file."""
  fn = os.path.join(
      os.getenv('TEST_UNDECLARED_OUTPUTS_DIR', default='/usr/local/google/tmp'),
      'ctrace.%s.%d.trace' % (os.path.basename(sys.argv[0]), os.getpid()))
  f = open(fn, mode='w')
  # We only ever open a JSON array. Ending the array is optional as per
  # go/trace_event so that not cleanly finished traces can still be read.
  f.write('[')
  return f


traceFile = _open_trace_file()
tracePid = os.getpid()
traceStart = time.time()


def microseconds_since_trace_start():
  return (time.time() - traceStart) * 1000000


def write_event(ev):
  traceFile.write(json.dumps(ev))
  traceFile.write(',\n')


write_event({
    'name': 'process_name',
    'ph': 'M',  # metadata event
    'pid': tracePid,
    'tid': tracePid,
    'args': {
        'name': ' '.join(sys.argv),
    },
})


class PendingEvent(object):
  """Pending trace event (not yet completed)."""

  def __init__(self, name, tid):
    self.name = name
    self.tid = tid

  def done(self):
    write_event({
        'name': self.name,
        'ph': 'E',  # Phase: End
        'pid': tracePid,
        'tid': self.tid,
        'ts': microseconds_since_trace_start(),
    })


def event(name, tid=tracePid):
  write_event({
      'name': name,
      'ph': 'B',  # Phase: Begin
      'pid': tracePid,
      'tid': tid,
      'ts': microseconds_since_trace_start(),
  })
  return PendingEvent(name, tid)


def flush():
  """Flushes the trace file to disk."""
  traceFile.flush()
