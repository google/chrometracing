# Copyright 2021 Google LLC
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

add_library(chrometracing
	chrometracing.cc
	compat.cc)
target_link_libraries(chrometracing absl::base absl::strings)
target_include_directories(chrometracing PUBLIC
	"${PROJECT_SOURCE_DIR}/abseil-cpp/absl")

enable_testing()

add_executable(chrometracing_init_test chrometracing_init_test.cc)
target_link_libraries(chrometracing_init_test chrometracing absl::base absl::strings absl::time)
add_test(init chrometracing_init_test)

install(TARGETS chrometracing DESTINATION lib)
install(FILES chrometracing.h DESTINATION include)
