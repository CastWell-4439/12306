#include <grpcpp/grpcpp.h>

#include <bitset>
#include <cstdint>
#include <iostream>
#include <memory>
#include <mutex>
#include <sstream>
#include <string>
#include <unordered_map>
#include <vector>

#include "seatallocator/v1/seatallocator.grpc.pb.h"

namespace {

constexpr uint32_t kSegmentCount = 64;
constexpr uint32_t kDefaultSeatCount = 200;

using ::grpc::Server;
using ::grpc::ServerBuilder;
using ::grpc::ServerContext;
using ::grpc::Status;
using ::seatallocator::v1::AllocateSeatRequest;
using ::seatallocator::v1::AllocateSeatResponse;
using ::seatallocator::v1::SeatAllocator;

struct SeatRecord {
  std::bitset<kSegmentCount> occupied_segments;
};

class RouteInventory {
 public:
  explicit RouteInventory(uint32_t seat_count) : seats_(seat_count) {}

  bool Allocate(uint32_t from_index, uint32_t to_index, std::string* seat_no_out) {
    if (from_index >= to_index || to_index > kSegmentCount) {
      return false;
    }
    const std::bitset<kSegmentCount> demand_mask = BuildMask(from_index, to_index);
    for (size_t i = 0; i < seats_.size(); ++i) {
      auto& seat = seats_[i];
      // O-D compatibility check with bit mask: can allocate only when no overlap.
      if ((seat.occupied_segments & demand_mask).any()) {
        continue;
      }
      seat.occupied_segments |= demand_mask;
      *seat_no_out = FormatSeatNo(i);
      return true;
    }
    return false;
  }

 private:
  static std::bitset<kSegmentCount> BuildMask(uint32_t from_index, uint32_t to_index) {
    std::bitset<kSegmentCount> mask;
    for (uint32_t i = from_index; i < to_index; ++i) {
      mask.set(i);
    }
    return mask;
  }

  static std::string FormatSeatNo(size_t idx) {
    constexpr size_t seats_per_coach = 50;
    const size_t coach = idx / seats_per_coach + 1;
    const size_t seat = idx % seats_per_coach + 1;
    std::ostringstream os;
    os << coach << "-" << seat;
    return os.str();
  }

  std::vector<SeatRecord> seats_;
};

class SeatAllocatorService final : public SeatAllocator::Service {
 public:
  Status AllocateSeat(ServerContext*,
                      const AllocateSeatRequest* req,
                      AllocateSeatResponse* resp) override {
    if (req->train_id().empty() || req->travel_date().empty() || req->coach_type().empty()) {
      return Status(grpc::StatusCode::INVALID_ARGUMENT, "train_id/travel_date/coach_type required");
    }

    const std::string route_key =
        req->train_id() + "|" + req->travel_date() + "|" + req->coach_type();

    std::lock_guard<std::mutex> lock(mu_);
    auto it = routes_.find(route_key);
    if (it == routes_.end()) {
      it = routes_.emplace(route_key, RouteInventory(kDefaultSeatCount)).first;
    }

    std::string seat_no;
    if (!it->second.Allocate(req->from_index(), req->to_index(), &seat_no)) {
      return Status(grpc::StatusCode::RESOURCE_EXHAUSTED, "no seat available for requested O-D");
    }
    resp->set_seat_no(seat_no);
    return Status::OK;
  }

 private:
  std::mutex mu_;
  std::unordered_map<std::string, RouteInventory> routes_;
};

void RunServer(const std::string& addr) {
  SeatAllocatorService service;
  ServerBuilder builder;
  builder.AddListeningPort(addr, grpc::InsecureServerCredentials());
  builder.RegisterService(&service);
  std::unique_ptr<Server> server(builder.BuildAndStart());
  std::cout << "seat-allocator listening on " << addr << std::endl;
  server->Wait();
}

}  // namespace

int main(int argc, char** argv) {
  std::string addr = "0.0.0.0:50051";
  if (argc > 1) {
    addr = argv[1];
  }
  RunServer(addr);
  return 0;
}


