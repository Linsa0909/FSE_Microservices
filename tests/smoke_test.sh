#!/bin/bash
# =============================================
#  冒烟测试 (Smoke Test) — 多域仿真平台
#  验证: 后端启动 → 10 配置组 → 编辑发布流程
#        → 3 领域服务拉取配置 → 健康检查 → 清理
# =============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

PASS=0
FAIL=0
BASE_URL="http://localhost:8080"

info()  { echo -e "${NC}[INFO] $*"; }
pass() { echo -e "${GREEN}[PASS]${NC} $*"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}[FAIL]${NC} $*"; FAIL=$((FAIL+1)); }

cleanup() {
    info "清理进程..."
    kill $BACKEND_PID 2>/dev/null || true
    kill $RADAR_PID 2>/dev/null || true
    kill $SENSOR_PID 2>/dev/null || true
    kill $NAV_PID 2>/dev/null || true
    wait $BACKEND_PID 2>/dev/null || true
    wait $RADAR_PID 2>/dev/null || true
    wait $SENSOR_PID 2>/dev/null || true
    wait $NAV_PID 2>/dev/null || true
}
trap cleanup EXIT

# ============================================
# Step 1: 编译验证
# ============================================
info "=== Step 1: 编译验证 ==="
export PATH=${PATH}:/root/go/bin

cd "$REPO_DIR/backend"
GOTOOLCHAIN=local go build -buildvcs=false -o /dev/null . && pass "backend build" || fail "backend build"

cd "$REPO_DIR/demo-service"
GOTOOLCHAIN=local go build -buildvcs=false -o /dev/null . && pass "demo-service build" || fail "demo-service build"

if [ $FAIL -gt 0 ]; then
    fail "编译失败，中止"
    exit 1
fi

# ============================================
# Step 2: 启动后端
# ============================================
info "=== Step 2: 启动后端 (:8080) ==="
cd "$REPO_DIR/backend"
go run . &
BACKEND_PID=$!
sleep 3

if kill -0 $BACKEND_PID 2>/dev/null; then
    pass "后端进程启动 (PID=$BACKEND_PID)"
else
    fail "后端进程未启动"
    exit 1
fi

# ============================================
# Step 3: 种子数据验证 (10 配置组)
# ============================================
info "=== Step 3: 种子数据验证 ==="

CONFIGS=$(curl -sf "$BASE_URL/api/configs" 2>/dev/null)
COUNT=$(echo "$CONFIGS" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['configs']))" 2>/dev/null || echo "0")

if [ "$COUNT" -ge 10 ]; then
    pass "种子数据: $COUNT 个配置组 (期望 >= 10)"
else
    fail "种子数据: $COUNT 个配置组 (期望 >= 10)"
fi

# 验证领域服务配置存在
for svc in "radar-service" "sensor-service" "navigation-service"; do
    if echo "$CONFIGS" | grep -q "$svc"; then
        pass "  $svc 配置组存在"
    else
        fail "  $svc 配置组缺失"
    fi
done

# 验证雷达 DEV 有 5 个配置项
RADAR_KEYS=$(curl -sf "$BASE_URL/api/configs/radar-service/dev" 2>/dev/null | python3 -c "import sys,json; print(len(json.load(sys.stdin)['config']['publishedData']))" 2>/dev/null || echo "0")
if [ "$RADAR_KEYS" -eq 5 ]; then
    pass "radar-service:dev 含 $RADAR_KEYS 个配置项 (期望 5)"
else
    fail "radar-service:dev 含 $RADAR_KEYS 个配置项 (期望 5)"
fi

# ============================================
# Step 4: 编辑 → 发布流程
# ============================================
info "=== Step 4: 编辑 → 发布流程 ==="

# 4.1 修改雷达模式: search → track
curl -sf -X PUT "$BASE_URL/api/configs/radar-service/dev/keys/radar.mode" \
    -H "Content-Type: application/json" \
    -d '{"value":"track"}' > /dev/null
pass "PUT /radar-service/dev/keys/radar.mode → track"

# 4.2 发布
PUBLISH_RESP=$(curl -sf -X POST "$BASE_URL/api/configs/radar-service/dev/publish" 2>/dev/null)
if echo "$PUBLISH_RESP" | grep -q "published"; then
    pass "POST /radar-service/dev/publish — 发布成功"
else
    fail "POST /radar-service/dev/publish — 发布失败"
fi

# 4.3 验证 PublishedData 已更新
MODE=$(curl -sf "$BASE_URL/api/configs/radar-service/dev/published" 2>/dev/null | grep -o '"radar.mode":"[^"]*"' | cut -d: -f2 | tr -d '"')
if [ "$MODE" = "track" ]; then
    pass "radar.mode 已发布生效: $MODE"
else
    fail "radar.mode 未生效: $MODE (期望 track)"
fi

# 4.4 版本号递增
VER=$(curl -sf "$BASE_URL/api/configs/radar-service/dev" 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)['config']['publishedVersion'])" 2>/dev/null)
if [ "$VER" = "1.0.1" ]; then
    pass "版本号递增: $VER"
else
    fail "版本号: $VER (期望 1.0.1)"
fi

# ============================================
# Step 5: 启动领域服务 + 健康检查
# ============================================
info "=== Step 5: 领域服务健康检查 ==="

# 恢复雷达模式的默认值 (方便后续)
curl -sf -X PUT "$BASE_URL/api/configs/radar-service/dev/keys/radar.mode" \
    -H "Content-Type: application/json" \
    -d '{"value":"search"}' > /dev/null
curl -sf -X POST "$BASE_URL/api/configs/radar-service/dev/publish" > /dev/null

cd "$REPO_DIR/demo-service"

# 5.1 火控雷达 :3001
SERVICE_TYPE=radar CONFIG_SERVICE=radar-service CONFIG_ENV=dev RADAR_PORT=3001 go run . &
RADAR_PID=$!
sleep 2
if kill -0 $RADAR_PID 2>/dev/null; then
    pass "雷达服务启动 (PID=$RADAR_PID)"
else
    fail "雷达服务未启动"
fi

RADAR_HEALTH=$(curl -sf http://localhost:3001/health 2>/dev/null)
if echo "$RADAR_HEALTH" | grep -q '"config_from_center":true'; then
    pass "雷达 /health: config_from_center=true"
else
    fail "雷达 /health 异常: $RADAR_HEALTH"
fi

RADAR_STATUS=$(curl -sf http://localhost:3001/radar/status 2>/dev/null)
if echo "$RADAR_STATUS" | grep -q "active_targets"; then
    pass "雷达 /radar/status 正常 (含 active_targets)"
else
    fail "雷达 /radar/status 异常"
fi

# 5.2 光电传感器 :3002
SERVICE_TYPE=sensor CONFIG_SERVICE=sensor-service CONFIG_ENV=dev SENSOR_PORT=3002 go run . &
SENSOR_PID=$!
sleep 2
if kill -0 $SENSOR_PID 2>/dev/null; then
    pass "传感器服务启动 (PID=$SENSOR_PID)"
else
    fail "传感器服务未启动"
fi

SENSOR_HEALTH=$(curl -sf http://localhost:3002/health 2>/dev/null)
if echo "$SENSOR_HEALTH" | grep -q '"config_from_center":true'; then
    pass "传感器 /health: config_from_center=true"
else
    fail "传感器 /health 异常"
fi

SENSOR_FRAME=$(curl -sf http://localhost:3002/sensor/frame 2>/dev/null)
if echo "$SENSOR_FRAME" | grep -q "frame_id"; then
    pass "传感器 /sensor/frame 正常 (含 frame_id)"
else
    fail "传感器 /sensor/frame 异常"
fi

# 5.3 船舶航海 :3003
SERVICE_TYPE=navigation CONFIG_SERVICE=navigation-service CONFIG_ENV=dev NAV_PORT=3003 go run . &
NAV_PID=$!
sleep 2
if kill -0 $NAV_PID 2>/dev/null; then
    pass "航海服务启动 (PID=$NAV_PID)"
else
    fail "航海服务未启动"
fi

NAV_HEALTH=$(curl -sf http://localhost:3003/health 2>/dev/null)
if echo "$NAV_HEALTH" | grep -q '"config_from_center":true'; then
    pass "航海 /health: config_from_center=true"
else
    fail "航海 /health 异常"
fi

NAV_POS=$(curl -sf http://localhost:3003/nav/position 2>/dev/null)
if echo "$NAV_POS" | grep -q '"lat"'; then
    pass "航海 /nav/position 正常 (含 lat/lng)"
else
    fail "航海 /nav/position 异常"
fi

# 5.4 雷达模式动态切换
OLD_MODE=$(curl -sf http://localhost:3001/radar/status 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)['config']['mode'])" 2>/dev/null)
curl -sf -X POST http://localhost:3001/radar/mode -H 'Content-Type: application/json' -d '{"mode":"engage"}' > /dev/null
NEW_MODE=$(curl -sf http://localhost:3001/radar/status 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)['config']['mode'])" 2>/dev/null)
if [ "$NEW_MODE" = "engage" ]; then
    pass "雷达模式动态切换: $OLD_MODE → $NEW_MODE"
else
    fail "雷达模式切换失败: $NEW_MODE (期望 engage)"
fi

# ============================================
# Step 6: 结果汇总
# ============================================
info ""
info "=========================================="
TOTAL=$((PASS+FAIL))
if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}  ✅ 冒烟测试全部通过: $PASS/$TOTAL${NC}"
else
    echo -e "${RED}  ❌ 冒烟测试失败: $PASS PASS / $FAIL FAIL / $TOTAL TOTAL${NC}"
fi
info "=========================================="

cleanup
trap - EXIT

exit $FAIL
