-- 数据库清理脚本：删除孤立的记录
-- 用途：清理已删除任务留下的关联记录，这些记录会阻止删除 PromQL

-- 查看孤立的 push_task_promql 记录（任务已被删除但关联还在）
SELECT 
    ptp.task_id,
    ptp.promql_id,
    p.name as promql_name
FROM push_task_promql ptp
LEFT JOIN push_task pt ON ptp.task_id = pt.id
LEFT JOIN promql p ON ptp.promql_id = p.id
WHERE pt.id IS NULL;

-- 删除孤立的 push_task_promql 记录
DELETE FROM push_task_promql 
WHERE task_id NOT IN (SELECT id FROM push_task);

-- 查看孤立的 push_task_webhook 记录
SELECT task_id, webhook_id
FROM push_task_webhook
WHERE task_id NOT IN (SELECT id FROM push_task);

-- 删除孤立的 push_task_webhook 记录
DELETE FROM push_task_webhook 
WHERE task_id NOT IN (SELECT id FROM push_task);

-- 查看孤立的 push_task_send_time 记录
SELECT task_id, weekday, send_time
FROM push_task_send_time
WHERE task_id NOT IN (SELECT id FROM push_task);

-- 删除孤立的 push_task_send_time 记录
DELETE FROM push_task_send_time 
WHERE task_id NOT IN (SELECT id FROM push_task);

-- 查看孤立的 push_task_query 记录（旧格式）
SELECT task_id, query
FROM push_task_query
WHERE task_id NOT IN (SELECT id FROM push_task);

-- 删除孤立的 push_task_query 记录
DELETE FROM push_task_query 
WHERE task_id NOT IN (SELECT id FROM push_task);

-- 显示清理结果统计
SELECT 
    'push_task_promql' as table_name,
    COUNT(*) as remaining_records
FROM push_task_promql
UNION ALL
SELECT 
    'push_task_webhook',
    COUNT(*)
FROM push_task_webhook
UNION ALL
SELECT 
    'push_task_send_time',
    COUNT(*)
FROM push_task_send_time
UNION ALL
SELECT 
    'push_task_query',
    COUNT(*)
FROM push_task_query;

