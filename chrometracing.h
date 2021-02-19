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

#ifndef CHROMETRACING_CHROMETRACING_H_
#define CHROMETRACING_CHROMETRACING_H_
// Offers Chrome tracing support for C++. Example use:
//
// {
//   auto e = chrometracing::Event("frog blast the vent core");
//
//   // actually frog blast the vent core
//
//   // when e falls out of scope the event span will get closed automatically
// }
//
// See go/chrometracing for more details on how tracing works.

#include <sys/types.h>

#include <optional>
#include <string>
#include <utility>

#include "absl/base/internal/sysinfo.h"

namespace chrometracing {
namespace internal {

enum class Phase: char {
  BEGIN = 'B',
  END = 'E',
  METADATA = 'M',
};

struct TraceEvent {
  std::string name;
  Phase phase;
  int64_t pid;
  int64_t tid;
  int64_t time;
  std::optional<std::string> process_name;
};

std::string RenderEvent(TraceEvent e);
}  // namespace internal

class PendingEvent {
 public:
  PendingEvent(std::string name, pid_t tid)
      : name_(std::move(name)), tid_(tid) {}
  ~PendingEvent();

 private:
  std::string name_;
  pid_t tid_;
};

PendingEvent Event(std::string name, int64_t explicit_tid);
PendingEvent Event(std::string name);


// Compatibility definitions:

std::string ProcessName(pid_t pid);


}  // namespace chrometracing

#endif  // CHROMETRACING_CHROMETRACING_H_
