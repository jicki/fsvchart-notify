# 升级到 v2.6.6+ 指南

## 🎉 新功能概览

### 1. 自动数据库迁移与清理

从 v2.6.6 开始，应用内置了智能数据库迁移系统，会自动：
- ✅ 检测当前数据库版本
- ✅ 执行必要的升级迁移
- ✅ **自动清理孤立的关联记录**
- ✅ 记录详细的清理统计

### 2. 完整的级联删除

删除任务时，现在会自动清理所有关联数据：
- ✅ `push_task_promql` - PromQL 关联
- ✅ `push_task_webhook` - Webhook 关联
- ✅ `push_task_send_time` - 发送时间配置
- ✅ `push_task_query` - 查询记录（旧格式）
- ✅ `push_task` - 任务主记录

### 3. 每个 PromQL 的独立配置

现在支持为每个 PromQL 单独配置：
- 📊 单位（MB, GB, %, ms 等）
- 🏷️ 指标标签（pod, namespace, container 等）
- ✏️ 自定义指标标签
- 📈 图表模板

### 4. 推送模式切换

- 📈 **图表模式**：显示时间序列图表
- 📝 **文本模式**：仅显示最新值

---

## 🚀 升级步骤

### 步骤 1：备份数据库（可选但推荐）

```bash
cd /path/to/fsvchart-notify
cp ./data/app.db ./data/app.db.backup.$(date +%Y%m%d_%H%M%S)
```

### 步骤 2：更新应用

```bash
# 如果从源码构建
git pull
make build

# 或使用新的二进制文件
# 替换 bin/fsvchart-notify
```

### 步骤 3：启动应用（自动迁移）

```bash
./bin/fsvchart-notify
```

### 步骤 4：查看迁移日志

启动应用后，在日志中查找以下信息：

```
[数据库] 当前数据库版本: 9, 最新版本: 10
[数据库] 开始迁移数据库，从版本 9 升级到 10
[数据库] 执行迁移: 版本 10 - 清理孤立的关联记录（已删除任务遗留的记录）
[数据库] 迁移版本 10 执行成功
[数据库清理] 孤立记录清理完成
[数据库清理] 当前记录数 - push_task_promql: X, push_task_webhook: X, ...
[数据库清理] 现在可以正常删除不再使用的 PromQL 了
[数据库] 迁移完成，验证表结构
```

如果看到 `数据库已是最新版本，无需迁移`，说明：
- 您已经是 v2.6.6 或更高版本
- 或者这是全新安装

---

## ✅ 验证升级成功

### 1. 验证数据库版本

```bash
sqlite3 ./data/app.db "SELECT version FROM schema_version ORDER BY version DESC LIMIT 1;"
# 应该显示: 10
```

### 2. 验证清理效果

尝试删除一个不再被使用的 PromQL：
1. 进入"PromQL 管理"页面
2. 选择一个 PromQL
3. 点击"删除"
4. 应该成功删除（如果没有任务在使用它）

### 3. 测试新功能

**测试独立 PromQL 配置**：
1. 创建新任务
2. 选择 2-3 个 PromQL
3. 观察每个 PromQL 自动展开
4. 为每个 PromQL 设置不同的单位和标签
5. 保存并查看任务列表

**测试推送模式**：
1. 创建任务时选择"文本模式"
2. 不需要选择图表模板
3. 任务执行时只显示最新值

---

## 🔍 常见问题

### Q1: 升级后，为什么我的 PromQL 还是无法删除？

**A**: 可能的原因：
1. 该 PromQL 确实正在被任务使用
   - 解决：在任务列表中搜索该 PromQL，删除或修改使用它的任务

2. 浏览器缓存问题
   - 解决：强制刷新浏览器 (`Ctrl+F5` 或 `Cmd+Shift+R`)

3. 应用未重启
   - 解决：确保重启应用以执行数据库迁移

### Q2: 如何确认迁移是否已执行？

**A**: 查看启动日志：
```bash
./bin/fsvchart-notify 2>&1 | grep "数据库"
```

或查询数据库版本：
```bash
sqlite3 ./data/app.db "SELECT version, description, applied_at FROM schema_version ORDER BY version;"
```

### Q3: 迁移失败了怎么办？

**A**: 
1. 查看详细错误日志
2. 恢复备份：
   ```bash
   cp ./data/app.db.backup.YYYYMMDD_HHMMSS ./data/app.db
   ```
3. 联系技术支持并提供错误日志

### Q4: 可以跳过自动迁移吗？

**A**: 不建议跳过。自动迁移确保：
- 数据库结构正确
- 清理无效数据
- 功能正常运行

如果确实需要，可以手动设置数据库版本（不推荐）：
```sql
INSERT INTO schema_version (version, applied_at, description) 
VALUES (10, datetime('now'), '手动设置');
```

---

## 📝 回滚指南

如果升级后遇到问题，可以回滚：

### 1. 停止应用

```bash
# 停止应用进程
pkill fsvchart-notify
```

### 2. 恢复数据库

```bash
cp ./data/app.db.backup.YYYYMMDD_HHMMSS ./data/app.db
```

### 3. 使用旧版本

```bash
# 使用之前的版本启动
./bin/fsvchart-notify.old
```

---

## 🎯 关键改进总结

| 功能 | 之前 | 现在 |
|------|------|------|
| 删除任务 | 部分清理，留下孤立记录 | 完整级联删除 |
| 删除 PromQL | 可能失败（被孤立记录阻止） | 正常删除 |
| 数据库清理 | 需要手动执行 SQL | 自动迁移清理 |
| PromQL 配置 | 任务级别统一配置 | 每个 PromQL 独立配置 |
| 推送模式 | 仅图表模式 | 图表/文本双模式 |

---

## 📞 获取帮助

如果遇到问题：
1. 查看 `scripts/README.md` 了解手动清理方法（备用）
2. 查看应用日志获取详细错误信息
3. 保留数据库备份以便恢复

---

**升级完成后，享受更稳定、更强大的功能！** 🎉

