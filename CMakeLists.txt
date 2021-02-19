cmake_minimum_required(VERSION 3.10)

project(chrometracing VERSION 0.0)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED True)

add_subdirectory(abseil-cpp)

add_library(chrometracing SHARED
	chrometracing.cc
	compat.cc)
target_link_libraries(chrometracing absl::base absl::strings)
target_include_directories(chrometracing PUBLIC
	"${PROJECT_SOURCE_DIR}/abseil-cpp/absl")

install(TARGETS chrometracing DESTINATION lib)
install(FILES "${PROJECT_SOURCE_DIR}/chrometracing.h" DESTINATION include)
