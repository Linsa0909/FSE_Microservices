#!/bin/bash
set -e

# =============================================
#  分布式配置中心 — 一键启动脚本 (多域仿真)
# =============================================

# Ensure Go is on PATH (common locations)
for d in /root/go/bin /usr/local/go/bin /usr/lib/go/bin /snap/go/current/bin; do
    if [ -d "$d" ]; then export PATH=$PATH:$d; fi
done

# Fallback: try to find go
command -v go >/dev/null 2>&1 || {
    echo "❌ go: command not found. Please install Go first."
    echo "   https://go.dev/dl/"
    exit 1
}

echo "========================================"
echo "  分布式配置中心 — 多域仿真平台"
echo "  启动中..."
echo "========================================"

# 1. 启动后端（配置中心）
echo "[1/5] 启动配置中心后端 (:8080)..."
(cd "$(dirname "$0")/backend" && go run .) &
BACKEND_PID=$!
sleep 2

# 2. 启动火控雷达服务
echo "[2/5] 启动火控雷达服务 (:3001)..."
(cd "$(dirname "$0")/demo-service" && \
    SERVICE_TYPE=radar CONFIG_SERVICE=radar-service CONFIG_ENV=dev RADAR_PORT=3001 go run .) &
RADAR_PID=$!
sleep 1

# 3. 启动光电传感器服务
echo "[3/5] 启动光电传感器服务 (:3002)..."
(cd "$(dirname "$0")/demo-service" && \
    SERVICE_TYPE=sensor CONFIG_SERVICE=sensor-service CONFIG_ENV=dev SENSOR_PORT=3002 go run .) &
SENSOR_PID=$!
sleep 1

# 4. 启动船舶航海服务
echo "[4/5] 启动船舶航海服务 (:3003)..."
(cd "$(dirname "$0")/demo-service" && \
    SERVICE_TYPE=navigation CONFIG_SERVICE=navigation-service CONFIG_ENV=dev NAV_PORT=3003 go run .) &
NAV_PID=$!
sleep 1

# 5. 启动前端
echo "[5/5] 启动配置中心前端 (:5173)..."
(cd "$(dirname "$0")/frontend" && npm run dev) &
FRONTEND_PID=$!

trap 'kill $BACKEND_PID $RADAR_PID $SENSOR_PID $NAV_PID $FRONTEND_PID 2>/dev/null' EXIT

echo ""
echo "========================================"
echo "  ✅ 全部启动完成！"
echo "  前端页面:        http://localhost:5173"
echo "  后端接口:        http://localhost:8080/api/configs"
echo "  🔴 火控雷达:      http://localhost:3001/radar/status"
echo "  📷 光电传感器:    http://localhost:3002/sensor/status"
echo "  🚢 船舶航海:      http://localhost:3003/nav/status"
echo "  Ctrl+C 一键停止所有进程"
echo "========================================"

wait
