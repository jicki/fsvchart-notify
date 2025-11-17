# 数据库维护脚本

## ✅ 自动清理（推荐）

从版本 v2.6.6+ 开始，应用已经内置了**自动数据库迁移**功能：

### 自动执行的操作

当您启动应用时，系统会自动：
1. ✅ 检测数据库版本
2. ✅ 执行必要的数据库迁移（从当前版本升级到最新版本）
3. ✅ **自动清理孤立的关联记录**（版本 10 迁移）
4. ✅ 记录清理统计信息到日志

### 查看自动清理结果

启动应用后，在日志中查看：

```
[数据库] 开始迁移数据库，从版本 9 升级到 10
[数据库] 执行迁移: 版本 10 - 清理孤立的关联记录（已删除任务遗留的记录）
[数据库] 迁移版本 10 执行成功
[数据库清理] 孤立记录清理完成
[数据库清理] 当前记录数 - push_task_promql: X, push_task_webhook: X, push_task_send_time: X, push_task_query: X
[数据库清理] 现在可以正常删除不再使用的 PromQL 了
```

### 无需手动操作

⚡ **只需要重启应用，系统会自动完成所有清理工作！**

---

## 🔧 手动清理（备用方案）

如果您需要手动清理（通常不需要），可以使用以下工具：

### 问题描述
当删除任务时，如果没有正确清理关联表中的记录，会导致：
- 无法删除 PromQL（提示"正在被任务使用"）
- 数据库中存在无效的关联记录

### 手动清理方案

#### 方法 1：使用 SQLite 命令行（推荐）

```bash
# 1. 进入项目目录
cd /Users/jicki/jicki/github/fsvchart-notify

# 2. 备份数据库（重要！）
cp ./data/app.db ./data/app.db.backup

# 3. 执行清理脚本
sqlite3 ./data/app.db < scripts/cleanup_orphaned_records.sql
```

#### 方法 2：手动执行 SQL（如果需要逐步操作）

```bash
# 打开数据库
sqlite3 ./data/app.db

# 查看孤立的记录
SELECT ptp.task_id, ptp.promql_id, p.name as promql_name
FROM push_task_promql ptp
LEFT JOIN push_task pt ON ptp.task_id = pt.id
LEFT JOIN promql p ON ptp.promql_id = p.id
WHERE pt.id IS NULL;

# 删除孤立的记录
DELETE FROM push_task_promql WHERE task_id NOT IN (SELECT id FROM push_task);
DELETE FROM push_task_webhook WHERE task_id NOT IN (SELECT id FROM push_task);
DELETE FROM push_task_send_time WHERE task_id NOT IN (SELECT id FROM push_task);
DELETE FROM push_task_query WHERE task_id NOT IN (SELECT id FROM push_task);

# 退出
.quit
```

#### 方法 3：在线工具

如果您使用 SQLite 图形化工具（如 DB Browser for SQLite），可以：
1. 打开 `./data/app.db`
2. 执行查询标签页
3. 复制并执行 `cleanup_orphaned_records.sql` 中的 SQL 语句

### 验证清理结果

清理完成后，重启应用并尝试删除 PromQL：

```bash
# 重启应用
./bin/fsvchart-notify
```

现在应该可以成功删除不再被使用的 PromQL 了。

### 预防措施

从现在开始，应用已经修复了删除任务的逻辑，会自动清理所有关联记录，不会再出现孤立记录的问题。

### 注意事项

⚠️ **在执行清理脚本之前，务必备份数据库！**

```bash
cp ./data/app.db ./data/app.db.backup.$(date +%Y%m%d_%H%M%S)
```

