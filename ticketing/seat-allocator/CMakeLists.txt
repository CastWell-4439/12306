cmake_minimum_required(VERSION 3.16)
project(seat_allocator CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Protobuf CONFIG REQUIRED)
find_package(gRPC CONFIG REQUIRED)

set(PROTO_FILE ${CMAKE_SOURCE_DIR}/../proto/seatallocator/v1/seatallocator.proto)
set(PROTO_GEN_DIR ${CMAKE_BINARY_DIR}/generated)
file(MAKE_DIRECTORY ${PROTO_GEN_DIR})

add_custom_command(
  OUTPUT
    ${PROTO_GEN_DIR}/seatallocator.pb.cc
    ${PROTO_GEN_DIR}/seatallocator.pb.h
    ${PROTO_GEN_DIR}/seatallocator.grpc.pb.cc
    ${PROTO_GEN_DIR}/seatallocator.grpc.pb.h
  COMMAND protobuf::protoc
  ARGS
    --proto_path=${CMAKE_SOURCE_DIR}/../proto
    --cpp_out=${PROTO_GEN_DIR}
    --grpc_out=${PROTO_GEN_DIR}
    --plugin=protoc-gen-grpc=$<TARGET_FILE:gRPC::grpc_cpp_plugin>
    ${PROTO_FILE}
  DEPENDS ${PROTO_FILE}
)

add_executable(seat_allocator
  src/main.cc
  ${PROTO_GEN_DIR}/seatallocator.pb.cc
  ${PROTO_GEN_DIR}/seatallocator.grpc.pb.cc
)

target_include_directories(seat_allocator PRIVATE ${PROTO_GEN_DIR})

target_link_libraries(seat_allocator
  PRIVATE
    protobuf::libprotobuf
    gRPC::grpc++
)


