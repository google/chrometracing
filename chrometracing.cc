// Copyright 2021 Google LLC
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

#include "chrometracing.h"

#include <stdio.h>
#include <unistd.h>

#include "base/googleinit.h"
#include "base/logging_extensions.h"
#include "file/base/path.h"
#include "third_party/absl/base/log_severity.h"
#include "third_party/absl/strings/str_replace.h"
#include "third_party/absl/strings/substitute.h"
#include "third_party/absl/time/clock.h"
#include "util/task/status.h"

namespace chrometracing {

namespace internal {
static FILE* log_file = nullptr;
static pid_t my_pid = 0;
static int64 start_nanos = 0;

std::string JSONEscape(const std::string& s) {
  return absl::StrReplaceAll(s, {
                                    {"\\", "\\\\"},
                                    {"\"", "\\\""},
                                    {"\n", "\\n"},
                                });
}

std::string RenderEvent(TraceEvent e) {
  std::vector<std::string> parts;
  parts.push_back(absl::Substitute(
      R"json({name: "$0", "ph": "$1", "pid": $2, "tid": $3, )json",
      JSONEscape(e.name), static_cast<char>(e.phase), e.pid, e.tid));
  if (e.time) {
    parts.push_back(absl::Substitute(R"json("time": $0, )json", e.time));
  }
  if (e.process_name) {
    parts.push_back(absl::Substitute(R"json("args": {"name": "$0"}, )json",
                                     JSONEscape(*e.process_name)));
  }
  parts.push_back("},\n");
  return absl::StrJoin(parts, "");
}

void WriteEvent(TraceEvent e) {
  if (log_file) {
    const std::string s = RenderEvent(e);
    fwrite(s.data(), sizeof(char), s.size(), log_file);
  }
}

std::string GetDestDir() {
  const char* env_var = getenv("TEST_UNDECLARED_OUTPUTS_DIR");
  if (env_var && env_var[0] != '\0') {
    return std::string(env_var);
  }
  return GetLoggingDirectories()[0];
}

REGISTER_MODULE_INITIALIZER(initialize_log_file, {
  start_nanos = absl::GetCurrentTimeNanos();
  const std::string dest_dir = GetDestDir();
  my_pid = getpid();
  // TODO(augie): I have to be missing some obvious way to do get process name?
  auto my_name = ProcessName(my_pid);
  auto dest_path = file::JoinPath(
      dest_dir, absl::Substitute("ctrace.$0.$1.trace", my_name, my_pid));
  LOG(INFO) << "Writing Chrome trace_events (for chrome::tracing) to "
            << dest_path;
  log_file = fopen(dest_path.c_str(), "w");
  if (!log_file) {
    PLOG(INFO) << "Failed to open " << dest_path
                << " for Chrome trace events, continuing without tracing";
    return;
  }
  fputs("[\n", log_file);
  WriteEvent(internal::TraceEvent{
      .name = "process_name",
      .phase = internal::Phase::METADATA,
      .pid = my_pid,
      .tid = GetTID(),
      .time = 0,
      .process_name = my_name,
  });
});

}  // namespace internal

PendingEvent::~PendingEvent() {
  WriteEvent(internal::TraceEvent{
      .name = name_,
      .phase = internal::Phase::END,
      .pid = internal::my_pid,
      .tid = tid_,
      .time = ((absl::GetCurrentTimeNanos() - internal::start_nanos) / 1000),
  });
}

PendingEvent Event(std::string name) { return Event(name, GetTID()); }

PendingEvent Event(std::string name, int64 tid) {
  WriteEvent(internal::TraceEvent{
      .name = name,
      .phase = internal::Phase::BEGIN,
      .pid = internal::my_pid,
      .tid = tid,
      .time = ((absl::GetCurrentTimeNanos() - internal::start_nanos) / 1000),
  });
  return PendingEvent(name, tid);
}

}  // namespace chrometracing
