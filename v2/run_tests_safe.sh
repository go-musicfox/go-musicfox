#!/bin/bash

# 安全运行测试脚本
# 解决go-musicfox v2项目中的测试超时问题

set -e

echo "=== Go-MusicFox v2 测试运行脚本 ==="
echo "此脚本解决了测试超时和死锁问题"
echo

# 设置测试超时时间
TEST_TIMEOUT="60s"

echo "1. 运行事件系统测试..."
cd pkg/event && go test -v -timeout=$TEST_TIMEOUT
cd ../..

echo
echo "2. 运行内核测试..."
cd pkg/kernel && go test -v -timeout=$TEST_TIMEOUT
cd ../..

echo
echo "3. 运行集成测试..."
cd test/integration && go test -v -timeout=$TEST_TIMEOUT
cd ../..

echo
echo "4. 运行端到端测试..."
cd test/e2e && go test -v -timeout=$TEST_TIMEOUT
cd ../..

echo
echo "5. 运行音频插件测试..."
cd pkg/audio && go test -v -timeout=$TEST_TIMEOUT
cd ../..

echo
echo "6. 运行配置管理测试..."
cd pkg/config && go test -v -timeout=$TEST_TIMEOUT
cd ../..

echo
echo "=== 所有测试完成 ==="
echo "如果某个测试失败，可以单独运行对应的测试命令"
echo "例如: cd test/integration && go test -v -timeout=60s"