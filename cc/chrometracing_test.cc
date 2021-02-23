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

#include "chrometracing.h"

#include "testing/base/public/gmock.h"
#include "testing/base/public/gunit.h"

namespace {

using testing::StrEq;
namespace internal = chrometracing::internal;

TEST(ChromeTracing, EventRenderCorrectly) {
  EXPECT_THAT(
      internal::RenderEvent(internal::TraceEvent{
          .name = "process_name",
          .phase = internal::Phase::METADATA,
          .pid = 42,
          .tid = 4242,
          .time = 0,
          .process_name = "some\"awful\n name\\to escape",
      }),
      StrEq(
          R"json({name: "process_name", "ph": "M", "pid": 42, "tid": 4242, "args": {"name": "some\"awful\n name\\to escape"}, },
)json"));
  EXPECT_THAT(
      internal::RenderEvent(internal::TraceEvent{
          .name = "process_name",
          .phase = internal::Phase::BEGIN,
          .pid = 42,
          .tid = 4242,
          .time = 32767,
      }),
      StrEq(
          R"json({name: "process_name", "ph": "B", "pid": 42, "tid": 4242, "time": 32767, },
)json"));
}

}  // namespace
