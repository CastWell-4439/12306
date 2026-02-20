package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName string
	Env         string
	Version     string
	HTTPPort    int

	MySQLDSN string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	KafkaBrokers []string

	InventoryServiceURL        string
	OrderInventoryPartitionKey string
	OrderInventoryDefaultQty   int
	OrderInventoryCapacity     int
	PaymentCallbackSignKey     string

	InventoryShardCount           int
	InventoryWALBuffer            int
	InventorySnapshotIntervalSecs int
	InventorySnapshotOpsThreshold int64
	InventoryHoldTTLSecs          int

	SeatAllocatorMode       string
	SeatAllocatorAddr       string
	SeatAllocatorTrainID    string
	SeatAllocatorTravelDate string
	SeatAllocatorCoachType  string
	SeatAllocatorFromIndex  int
	SeatAllocatorToIndex    int
}

func Load(serviceName string) Config {
	return Config{
		ServiceName:                   serviceName,
		Env:                           getenv("APP_ENV", "dev"),
		Version:                       getenv("APP_VERSION", "stage1"),
		HTTPPort:                      getenvInt("HTTP_PORT", 8080),
		MySQLDSN:                      getenv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/ticketing?parseTime=true"),
		RedisAddr:                     getenv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:                 getenv("REDIS_PASSWORD", ""),
		RedisDB:                       getenvInt("REDIS_DB", 0),
		KafkaBrokers:                  splitCSV(getenv("KAFKA_BROKERS", "127.0.0.1:9092")),
		InventoryServiceURL:           getenv("INVENTORY_SERVICE_URL", "http://127.0.0.1:8082"),
		OrderInventoryPartitionKey:    getenv("ORDER_INVENTORY_PARTITION_KEY", "G123|2026-02-11|2nd"),
		OrderInventoryDefaultQty:      getenvInt("ORDER_INVENTORY_DEFAULT_QTY", 1),
		OrderInventoryCapacity:        getenvInt("ORDER_INVENTORY_CAPACITY", 500),
		PaymentCallbackSignKey:        getenv("PAYMENT_CALLBACK_SIGN_KEY", ""),
		InventoryShardCount:           getenvInt("INVENTORY_SHARD_COUNT", 32),
		InventoryWALBuffer:            getenvInt("INVENTORY_WAL_BUFFER", 4096),
		InventorySnapshotIntervalSecs: getenvInt("INVENTORY_SNAPSHOT_INTERVAL_SECS", 10),
		InventorySnapshotOpsThreshold: int64(getenvInt("INVENTORY_SNAPSHOT_OPS_THRESHOLD", 500)),
		InventoryHoldTTLSecs:          getenvInt("INVENTORY_HOLD_TTL_SECS", 120),
		SeatAllocatorMode:             getenv("SEAT_ALLOCATOR_MODE", "mock"),
		SeatAllocatorAddr:             getenv("SEAT_ALLOCATOR_ADDR", "127.0.0.1:50051"),
		SeatAllocatorTrainID:          getenv("SEAT_ALLOCATOR_TRAIN_ID", "G123"),
		SeatAllocatorTravelDate:       getenv("SEAT_ALLOCATOR_TRAVEL_DATE", "2026-02-11"),
		SeatAllocatorCoachType:        getenv("SEAT_ALLOCATOR_COACH_TYPE", "2nd"),
		SeatAllocatorFromIndex:        getenvInt("SEAT_ALLOCATOR_FROM_INDEX", 1),
		SeatAllocatorToIndex:          getenvInt("SEAT_ALLOCATOR_TO_INDEX", 3),
	}
}

func DefaultRequestTimeout() time.Duration {
	return 2 * time.Second
}

func getenv(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func getenvInt(key string, defaultValue int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return v
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return []string{"127.0.0.1:9092"}
	}
	return out
}
