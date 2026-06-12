#!/bin/bash
# =============================================
#  冒烟测试 (Smoke Test)
#  验证: 后端启动 → API 可用 → 编辑发布流程
#        → demo-service 拉取配置 → 清理
# =============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0
BASE_URL="http://localhost:8080"
DEMO_URL="http://localhost:3000"

info()  { echo -e "${NC}[INFO] $*"; }
pass() { echo -e "${GREEN}[PASS]${NC} $*"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}[FAIL]${NC} $*"; FAIL=$((FAIL+1)); }

cleanup() {
    info "清理进程..."
    kill $BACKEND_PID 2>/dev/null || true
    kill $DEMO_PID 2>/dev/null || true
    wait $BACKEND_PID 2>/dev/null || true
    wait $DEMO_PID 2>/dev/null || true
}
trap cleanup EXIT

# ============================================
# Step 1: 编译
# ============================================
info "=== Step 1: 编译 ==="

cd "$REPO_DIR/backend"
export PATH=${PATH}:/root/go/bin
GOTOOLCHAIN=local go build -o config-center . && pass "backend build" || fail "backend build"

cd "$REPO_DIR/demo-service"
GOTOOLCHAIN=local go build -o demo-service . && pass "demo-service build" || fail "demo-service build"

if [ $FAIL -gt 0 ]; then
    fail "编译失败，中止"
    exit 1
fi

# ============================================
# Step 2: 启动后端
# ============================================
info "=== Step 2: 启动后端 ==="

cd "$REPO_DIR/backend"
./config-center &
BACKEND_PID=$!
sleep 2

# 验证后端存活
if kill -0 $BACKEND_PID 2>/dev/null; then
    pass "后端进程启动 (PID=$BACKEND_PID)"
else
    fail "后端进程未启动"
    exit 1
fi

# ============================================
# Step 3: API 基础验证
# ============================================
info "=== Step 3: API 基础验证 ==="

# 3.1 GET /api/configs — 种子数据
CONFIGS=$(curl -sf "$BASE_URL/api/configs" 2>/dev/null)
if echo "$CONFIGS" | grep -q "order-service"; then
    pass "GET /api/configs — 种子数据包含 order-service"
else
    fail "GET /api/configs — 种子数据缺失"
fi

COUNT=$(echo "$CONFIGS" | grep -o '"service"' | wc -l)
if [ "$COUNT" -ge 3 ]; then
    pass "GET /api/configs — 共 $COUNT 个配置组 (期望 >= 3)"
else
    fail "GET /api/configs — 只有 $COUNT 个配置组 (期望 >= 3)"
fi

# 3.2 GET /api/configs/order-service/dev — 单个配置组
DETAIL=$(curl -sf "$BASE_URL/api/configs/order-service/dev" 2>/dev/null)
if echo "$DETAIL" | grep -q '"db.url"'; then
    pass "GET /api/configs/order-service/dev — 包含 db.url 配置项"
else
    fail "GET /api/configs/order-service/dev — 缺少配置项"
fi

# 3.3 GET /api/configs/order-service/dev/published — 微服务接口
PUBLISHED=$(curl -sf "$BASE_URL/api/configs/order-service/dev/published" 2>/dev/null)
if echo "$PUBLISHED" | grep -q '"db.url":"localhost:3306"'; then
    pass "GET /api/configs/order-service/dev/published — 返回线上配置"
else
    fail "GET /api/configs/order-service/dev/published — 配置不正确"
fi

# 3.4 预留接口
PUSH=$(curl -sf -X POST "$BASE_URL/api/configs/order-service/dev/push" 2>/dev/null)
if echo "$PUSH" | grep -q "reserved"; then
    pass "POST /push — 预留接口正常"
else
    fail "POST /push — 预留接口异常"
fi

WATCH=$(curl -sf "$BASE_URL/api/configs/watch" 2>/dev/null)
if echo "$WATCH" | grep -q "reserved"; then
    pass "GET /watch — 预留接口正常"
else
    fail "GET /watch — 预留接口异常"
fi

# ============================================
# Step 4: 编辑 → 发布流程
# ============================================
info "=== Step 4: 编辑 → 发布流程 ==="

# 4.1 修改配置
curl -sf -X PUT "$BASE_URL/api/configs/order-service/dev/keys/smoke.test" \
    -H "Content-Type: application/json" \
    -d '{"value":"smoke_test_value"}' > /dev/null
pass "PUT /keys/smoke.test — 新增测试配置项"

# 4.2 发布
PUBLISH_RESP=$(curl -sf -X POST "$BASE_URL/api/configs/order-service/dev/publish" 2>/dev/null)
if echo "$PUBLISH_RESP" | grep -q "published"; then
    pass "POST /publish — 发布成功"
else
    fail "POST /publish — 发布失败"
fi

# 4.3 验证发布后版本号递增
VER=$(curl -sf "$BASE_URL/api/configs/order-service/dev" 2>/dev/null | grep -o '"publishedVersion":[0-9]*' | head -1 | cut -d: -f2)
if [ "$VER" -ge 2 ]; then
    pass "版本号递增: v$VER (期望 >= 2)"
else
    fail "版本号未递增: v$VER"
fi

# 4.4 验证新的配置值已生效
NEWVAL=$(curl -sf "$BASE_URL/api/configs/order-service/dev/published" 2>/dev/null | grep -o '"smoke.test":"[^"]*"' | cut -d: -f2 | tr -d '"')
if [ "$NEWVAL" = "smoke_test_value" ]; then
    pass "配置项 smoke.test 已发布生效"
else
    fail "配置项 smoke.test 未生效: $NEWVAL"
fi

# 4.5 删除 → 再次发布
curl -sf -X DELETE "$BASE_URL/api/configs/order-service/dev/keys/smoke.test" > /dev/null
curl -sf -X POST "$BASE_URL/api/configs/order-service/dev/publish" > /dev/null
DELCHECK=$(curl -sf "$BASE_URL/api/configs/order-service/dev/published" 2>/dev/null)
if ! echo "$DELCHECK" | grep -q "smoke.test"; then
    pass "删除 smoke.test 后发布 → 已从线上移除"
else
    fail "删除 smoke.test 后发布 → 仍然存在"
fi

# ============================================
# Step 5: Demo Service 拉取配置
# ============================================
info "=== Step 5: Demo Service 拉取配置 ==="

cd "$REPO_DIR/demo-service"
./demo-service &
DEMO_PID=$!
sleep 2

if kill -0 $DEMO_PID 2>/dev/null; then
    pass "Demo Service 启动 (PID=$DEMO_PID)"
else
    fail "Demo Service 未启动"
    exit 1
fi

DEMO_RESP=$(curl -sf "$DEMO_URL/" 2>/dev/null)
if echo "$DEMO_RESP" | grep -q "order-service"; then
    pass "GET / — 返回 order-service 标识"
else
    fail "GET / — 未返回服务标识"
fi

if echo "$DEMO_RESP" | grep -q "db.url"; then
    pass "Demo Service 成功拉取到配置 (含 db.url)"
else
    fail "Demo Service 未拉取到配置"
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

# 清理
cleanup
trap - EXIT

exit $FAIL
