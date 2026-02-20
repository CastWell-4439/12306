CREATE TABLE IF NOT EXISTS query_order_view (
  order_id VARCHAR(64) PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  amount_cents BIGINT NOT NULL DEFAULT 0,
  provider_txn_id VARCHAR(128) NOT NULL DEFAULT '',
  seat_no VARCHAR(64) NOT NULL DEFAULT '',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_query_order_view_status (status),
  KEY idx_query_order_view_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


