FROM ubuntu:24.04 AS build

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    git \
    pkg-config \
    protobuf-compiler \
    protobuf-compiler-grpc \
    libprotobuf-dev \
    libgrpc++-dev \
    libgrpc-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . /app

WORKDIR /app/seat-allocator
RUN cmake -S . -B build && cmake --build build -j

FROM ubuntu:24.04
RUN apt-get update && apt-get install -y --no-install-recommends \
    libprotobuf-dev \
    libgrpc++-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /app/seat-allocator/build/seat_allocator /usr/local/bin/seat_allocator

EXPOSE 50051
ENTRYPOINT ["/usr/local/bin/seat_allocator"]


