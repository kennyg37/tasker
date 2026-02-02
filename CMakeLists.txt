cmake_minimum_required(VERSION 3.10)
project(Tasker VERSION 1.0)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include_directories(${PROJECT_SOURCE_DIR}/include)

set(SOURCES
    src/main.cpp
    src/Task.cpp
    src/TodoParser.cpp
    src/TodoManager.cpp
    src/Display.cpp
)

add_executable(tasker ${SOURCES})

install(TARGETS tasker DESTINATION bin)
