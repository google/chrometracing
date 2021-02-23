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

#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>

#include "absl/strings/substitute.h"

#include "chrometracing.h"

namespace chrometracing {

std::string ProcessName(pid_t pid) {
    if (!pid)
        pid = getpid();
    std::string filename = absl::Substitute("/proc/$0/comm", pid);

    int fd = open(filename.c_str(), O_RDONLY);
    if (fd == -1)
        return {};
    char buf[64];
    int len = read(fd, buf, sizeof (buf));
    close(fd);
    if (len == -1)
        return {};
    if (len > 0 && buf[len-1] == '\n')
        len--;
    return std::string(buf, len);
}


//}  // namespace chrometracing

}
