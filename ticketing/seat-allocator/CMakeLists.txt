cmake_minimum_required(VERSION 3.16)
project(seat_allocator CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Ubuntu 24.04 apt packages provide Module-mode cmake files + pkg-config
find_package(Protobuf REQUIRED)
find_package(PkgConfig REQUIRED)
pkg_check_modules(GRPC REQUIRED IMPORTED_TARGET grpc++ grpc)

# Locate protoc and grpc_cpp_plugin executables
find_program(PROTOC_EXECUTABLE protoc REQUIRED)
find_program(GRPC_CPP_PLUGIN grpc_cpp_plugin REQUIRED)

set(PROTO_FILE ${CMAKE_SOURCE_DIR}/../proto/seatallocator/v1/seatallocator.proto)
set(PROTO_GEN_DIR ${CMAKE_BINARY_DIR}/generated)
file(MAKE_DIRECTORY ${PROTO_GEN_DIR})

add_custom_command(
  OUTPUT
    ${PROTO_GEN_DIR}/seatallocator/v1/seatallocator.pb.cc
    ${PROTO_GEN_DIR}/seatallocator/v1/seatallocator.pb.h
    ${PROTO_GEN_DIR}/seatallocator/v1/seatallocator.grpc.pb.cc
    ${PROTO_GEN_DIR}/seatallocator/v1/seatallocator.grpc.pb.h
  COMMAND ${CMAKE_COMMAND} -E make_directory ${PROTO_GEN_DIR}/seatallocator/v1
  COMMAND ${PROTOC_EXECUTABLE}
    --proto_path=${CMAKE_SOURCE_DIR}/../proto
    --cpp_out=${PROTO_GEN_DIR}
    --grpc_out=${PROTO_GEN_DIR}
    --plugin=protoc-gen-grpc=${GRPC_CPP_PLUGIN}
    ${PROTO_FILE}
  DEPENDS ${PROTO_FILE}
)

add_executable(seat_allocator
  src/main.cc
  ${PROTO_GEN_DIR}/seatallocator/v1/seatallocator.pb.cc
  ${PROTO_GEN_DIR}/seatallocator/v1/seatallocator.grpc.pb.cc
)

target_include_directories(seat_allocator PRIVATE ${PROTO_GEN_DIR})

target_link_libraries(seat_allocator
  PRIVATE
    protobuf::libprotobuf
    PkgConfig::GRPC
)
