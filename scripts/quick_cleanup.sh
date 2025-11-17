#!/bin/bash

# 数据库快速清理脚本
# 用途：清理孤立的关联记录，允许删除 PromQL

set -e

DB_PATH="./data/app.db"
BACKUP_PATH="./data/app.db.backup.$(date +%Y%m%d_%H%M%S)"

echo "========================================"
echo "数据库清理工具"
echo "========================================"
echo ""

# 检查数据库是否存在
if [ ! -f "$DB_PATH" ]; then
    echo "❌ 错误：数据库文件不存在: $DB_PATH"
    exit 1
fi

# 备份数据库
echo "📦 正在备份数据库..."
cp "$DB_PATH" "$BACKUP_PATH"
echo "✅ 备份完成: $BACKUP_PATH"
echo ""

# 查看孤立记录
echo "🔍 检查孤立记录..."
ORPHANED_PROMQL=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM push_task_promql WHERE task_id NOT IN (SELECT id FROM push_task);")
ORPHANED_WEBHOOK=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM push_task_webhook WHERE task_id NOT IN (SELECT id FROM push_task);")
ORPHANED_SENDTIME=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM push_task_send_time WHERE task_id NOT IN (SELECT id FROM push_task);")
ORPHANED_QUERY=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM push_task_query WHERE task_id NOT IN (SELECT id FROM push_task);")

echo "   push_task_promql 孤立记录: $ORPHANED_PROMQL"
echo "   push_task_webhook 孤立记录: $ORPHANED_WEBHOOK"
echo "   push_task_send_time 孤立记录: $ORPHANED_SENDTIME"
echo "   push_task_query 孤立记录: $ORPHANED_QUERY"
echo ""

TOTAL_ORPHANED=$((ORPHANED_PROMQL + ORPHANED_WEBHOOK + ORPHANED_SENDTIME + ORPHANED_QUERY))

if [ $TOTAL_ORPHANED -eq 0 ]; then
    echo "✅ 没有发现孤立记录，数据库状态良好！"
    echo ""
    rm -f "$BACKUP_PATH"
    echo "已删除备份文件（不需要）"
    exit 0
fi

# 清理孤立记录
echo "🧹 正在清理孤立记录..."
sqlite3 "$DB_PATH" "DELETE FROM push_task_promql WHERE task_id NOT IN (SELECT id FROM push_task);"
sqlite3 "$DB_PATH" "DELETE FROM push_task_webhook WHERE task_id NOT IN (SELECT id FROM push_task);"
sqlite3 "$DB_PATH" "DELETE FROM push_task_send_time WHERE task_id NOT IN (SELECT id FROM push_task);"
sqlite3 "$DB_PATH" "DELETE FROM push_task_query WHERE task_id NOT IN (SELECT id FROM push_task);"
echo "✅ 清理完成！"
echo ""

# 显示清理结果
echo "📊 清理结果："
echo "   删除了 $ORPHANED_PROMQL 条 push_task_promql 记录"
echo "   删除了 $ORPHANED_WEBHOOK 条 push_task_webhook 记录"
echo "   删除了 $ORPHANED_SENDTIME 条 push_task_send_time 记录"
echo "   删除了 $ORPHANED_QUERY 条 push_task_query 记录"
echo ""
echo "✅ 数据库清理成功！现在可以删除不再使用的 PromQL 了。"
echo ""
echo "💾 备份文件保存在: $BACKUP_PATH"
echo "   如果一切正常，可以手动删除备份文件。"
echo ""
