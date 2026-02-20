# C++ Seat Allocator (Stage 5)

This service provides gRPC `AllocateSeat` with an in-memory bitmap/bitset model.

## Build (local)

Requirements:

- cmake
- g++
- protobuf + grpc C++ dev libraries

Commands:

```bash
cd seat-allocator
cmake -S . -B build
cmake --build build -j
./build/seat_allocator 0.0.0.0:50051
```

## Build (docker)

```bash
docker build -f seat-allocator/Dockerfile -t seat-allocator:dev .
docker run --rm -p 50051:50051 seat-allocator:dev
```

## Allocation Model

- `RouteKey = train_id|travel_date|coach_type`
- Each seat keeps `bitset<64>` segment occupancy.
- For request `[from_index, to_index)`, build mask and find first seat where:
  - `(occupied_mask & request_mask) == 0`
- On success, set bits and return seat number (`coach-seat` format).


