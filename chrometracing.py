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
  in_test = os.getenv('TEST_TMPDIR')
  explicitly_enabled = os.getenv('CHROMETRACING_DIR')
  enable_tracing = in_test or explicitly_enabled
  if not enable_tracing:
    return None
  output_dir = os.getenv(
      'TEST_UNDECLARED_OUTPUTS_DIR',
      default=os.getenv('CHROMETRACING_DIR', default='/usr/local/google/tmp'))
  fn = os.path.join(
      output_dir,
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
  if not traceFile:
    return
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
    _release_tid(self.tid)


def event(name):
  tid = _tid()
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
  if not traceFile:
    return
  traceFile.flush()

# tids is a chrome://tracing thread id pool. Python does not have threads or
# thread ids (unlike e.g. Java or C++, but similar to Go), so we need to
# maintain our own identifier. The chrome://tracing file format requires a
# numeric thread id, so we just increment whenever we need a thread id, and
# reuse the ones no longer in use.
tids = []


def _tid():
  # Re-use released tids if any:
  for tid, used in enumerate(tids):
    if not used:
      tids[tid] = True
      return tid
  tid = len(tids)
  tids.append(True)
  return tid


def _release_tid(tid):
  tids[tid] = False
