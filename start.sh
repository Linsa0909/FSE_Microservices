#!/bin/bash
set -e

# =============================================
#  分布式配置中心 — 一键启动脚本
# =============================================

echo "========================================"
echo "  分布式配置中心 (Config Center)"
echo "  启动中..."
echo "========================================"

# 1. 启动后端（配置中心）
echo "[1/3] 启动配置中心后端 (:8080)..."
(cd "$(dirname "$0")/backend" && go run .) &
BACKEND_PID=$!
sleep 2

# 2. 启动微服务示例
echo "[2/3] 启动示例微服务 (:3000)..."
(cd "$(dirname "$0")/demo-service" && go run .) &
DEMO_PID=$!
sleep 1

# 3. 启动前端
echo "[3/3] 启动配置中心前端 (:5173)..."
(cd "$(dirname "$0")/frontend" && npm run dev) &
FRONTEND_PID=$!

trap 'kill $BACKEND_PID $DEMO_PID $FRONTEND_PID 2>/dev/null' EXIT

echo ""
echo "========================================"
echo "  ✅ 启动完成！"
echo "  前端页面:  http://localhost:5173"
echo "  后端接口:  http://localhost:8080/api/configs"
echo "  微服务:    http://localhost:3000"
echo "  Ctrl+C 一键停止所有进程"
echo "========================================"

wait
