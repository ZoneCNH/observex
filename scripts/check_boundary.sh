#!/usr/bin/env bash
set -euo pipefail

echo "checking forbidden dependency on x.go..."

DEPS="$(GOWORK=off go list -deps ./...)"
LEGACY_XGO_DEP="github.com/byte""chainx/x.go"
CURRENT_XGO_DEP="github.com/ZoneCNH/x.go"
FORBIDDEN_DEPS=(
  "$LEGACY_XGO_DEP"
  "$CURRENT_XGO_DEP"
  "go.uber.org/zap"
  "github.com/sirupsen/logrus"
  "github.com/prometheus/client_golang"
  "go.opentelemetry.io/otel"
  "github.com/grafana/loki"
  "github.com/grafana/tempo"
  "github.com/jaegertracing/jaeger"
  "github.com/redis/go-redis"
  "github.com/Shopify/sarama"
  "github.com/IBM/sarama"
  "github.com/jackc/pgx"
  "github.com/lib/pq"
  "github.com/ClickHouse/clickhouse-go"
  "github.com/tdengine/driver-go"
  "github.com/aliyun/aliyun-oss-go-sdk"
  "github.com/aws/aws-sdk-go"
  "github.com/adshao/go-binance"
  "gorm.io/gorm"
)

for dep in "${FORBIDDEN_DEPS[@]}"; do
  if grep -Fq "$dep" <<<"$DEPS"; then
    echo "ERROR: observex core must not depend on forbidden dependency: $dep"
    exit 1
  fi
done

echo "checking public package import allowlist..."
module="$(GOWORK=off go list -m)"
import_edges="$(GOWORK=off go list -f '{{.ImportPath}} {{join .Imports ","}}' ./pkg/... ./testkit/...)"
while IFS= read -r line; do
  pkg="${line%% *}"
  imports="${line#* }"
  if [[ "$pkg" == "$imports" ]]; then
    imports=""
  fi
  IFS=',' read -ra import_list <<< "$imports"
  for import_path in "${import_list[@]}"; do
    [[ -z "$import_path" ]] && continue
    if [[ "$pkg" == "$module/pkg"* ]]; then
      if [[ "$import_path" == "$module/"* && "$import_path" != "$module/internal"* ]]; then
        echo "ERROR: pkg import boundary violation: $pkg imports $import_path"
        exit 1
      fi
      if [[ "$import_path" == *.* && "$import_path" != "github.com/ZoneCNH/foundationx/pkg/foundationx" && "$import_path" != "$module/internal"* ]]; then
        echo "ERROR: pkg external dependency boundary violation: $pkg imports $import_path"
        exit 1
      fi
    fi
    if [[ "$pkg" == "$module/testkit"* ]]; then
      if [[ "$import_path" == "$module/internal"* || ( "$import_path" == *.* && "$import_path" != "$module/pkg/observex" ) ]]; then
        echo "ERROR: testkit import boundary violation: $pkg imports $import_path"
        exit 1
      fi
    fi
  done
done <<< "$import_edges"

echo "checking forbidden business terms..."

FORBIDDEN_TERMS=(
  "MacroRegime"
  "MarketRegime"
  "TradingSignal"
  "BTCUSDT"
  "Binance"
  "ETHUSDT"
  "FRED"
  "Jaeger"
  "Kline"
  "Loki"
  "OrderBook"
  "Position"
  "Postgres"
  "RiskGate"
  "TDengine"
  "Tempo"
)

for term in "${FORBIDDEN_TERMS[@]}"; do
  if grep -R --line-number --fixed-strings "$term" ./pkg ./internal ./examples ./scripts ./testkit --exclude=check_boundary.sh --exclude-dir=.git; then
    echo "ERROR: forbidden business term found: $term"
    exit 1
  fi
done

echo "checking public API boundary terms..."
for term in "${FORBIDDEN_TERMS[@]}"; do
  if grep -R --line-number --fixed-strings "$term" contracts/public_api.md contracts/public_api.snapshot; then
    echo "ERROR: forbidden business term found in public API contract: $term"
    exit 1
  fi
done

echo "boundary check passed"
