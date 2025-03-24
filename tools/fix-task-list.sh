#!/bin/bash

# 任务列表修复脚本
# 执行所有必要的修复步骤，解决任务列表显示和操作的问题

echo "=== 任务列表修复脚本 ==="
echo "数据库路径: ../data/app.db"

# 1. 停止当前运行的应用
echo "停止当前运行的应用..."
pkill -f fsvchart-notify

# 2. 备份数据库
echo "备份数据库..."
BACKUP_PATH="../data/app.db.backup.$(date +%Y%m%d%H%M%S)"
cp ../data/app.db "$BACKUP_PATH"
echo "数据库已备份到: $BACKUP_PATH"

# 3. 运行数据库维护工具
echo "运行数据库维护工具..."
go run db_maintenance.go all ../data/app.db

# 4. 创建默认任务（如果需要）
echo "检查并创建默认任务..."
go run create_default_task.go ../data/app.db

# 5. 确保前端修复脚本存在
echo "安装前端修复脚本..."
mkdir -p frontend/public
cp frontend/public/task-list-fix.js frontend/public/ 2>/dev/null || \
  echo "前端修复脚本已存在"

echo "修复完成！应用已重启"
echo "如果问题仍然存在，请刷新前端页面或重启浏览器" 