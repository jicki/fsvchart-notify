/**
 * 前端数据修复脚本
 * 
 * 此脚本用于修复前端显示的问题，确保数据正确显示在任务列表中
 * 使用方法：在网页控制台运行这段代码
 */

(function() {
    console.log('开始应用前端数据修复脚本...');
    
    // 修复HTTP响应拦截
    const originalFetch = window.fetch;
    window.fetch = async function(url, options) {
        const response = await originalFetch(url, options);
        
        // 只拦截push_task相关请求
        if (url.includes('/api/push_task') && !url.includes('/api/push_task/')) {
            const clone = response.clone();
            clone.json().then(data => {
                console.log('拦截到任务列表数据:', data);
                
                // 检查并修复数据格式问题
                if (Array.isArray(data)) {
                    console.log('数据是数组格式，正在处理...');
                    fixTaskData(data);
                } else if (data && data.data && Array.isArray(data.data)) {
                    console.log('数据是{data:[...]}格式，正在处理...');
                    fixTaskData(data.data);
                }
            }).catch(err => {
                console.error('解析响应数据失败:', err);
            });
        }
        
        return response;
    };
    
    // 修复任务数据
    function fixTaskData(tasks) {
        tasks.forEach(task => {
            // 确保ID是有效值
            if (!task.id || task.id < 0) {
                console.warn('发现无效ID:', task.id);
                // 避免使用负数ID，可能导致删除操作问题
                task.id = null;
            }
            
            // 确保时间范围有效
            if (!task.time_range || task.time_range === 'undefined') {
                console.warn('修复无效时间范围');
                task.time_range = '30m';
            }
            
            // 确保初次发送时间有效
            if (!task.initial_send_time || task.initial_send_time === 'undefined') {
                console.warn('修复无效初次发送时间');
                task.initial_send_time = '08:00';
            }
            
            // 确保启用状态是布尔值
            if (typeof task.enabled !== 'boolean') {
                task.enabled = Boolean(task.enabled);
            }
            
            // 确保所有数组字段都存在
            if (!Array.isArray(task.bound_webhooks)) {
                task.bound_webhooks = [];
            }
            
            if (!Array.isArray(task.queries)) {
                task.queries = [];
            }
        });
        
        console.log('数据修复完成');
    }
    
    // 为任务列表页面添加点击事件监听器
    function setupClickHandlers() {
        console.log('设置点击事件处理...');
        
        // 5秒后检查DOM是否包含任务列表
        setTimeout(() => {
            // 查找所有按钮
            const buttons = document.querySelectorAll('button');
            
            buttons.forEach(button => {
                // 删除按钮的特殊处理
                if (button.innerText.includes('删除') || button.className.includes('delete')) {
                    button.addEventListener('click', function(event) {
                        // 获取任务ID
                        const taskId = extractTaskIdFromButton(button);
                        
                        if (!taskId || taskId === 'undefined' || taskId < 0) {
                            console.warn('阻止删除无效ID的任务:', taskId);
                            event.preventDefault();
                            event.stopPropagation();
                            alert('无法删除此任务：ID无效');
                            return false;
                        }
                    }, true);
                }
                
                // 编辑按钮的特殊处理
                if (button.innerText.includes('编辑') || button.className.includes('edit')) {
                    button.addEventListener('click', function(event) {
                        // 获取任务ID
                        const taskId = extractTaskIdFromButton(button);
                        
                        if (!taskId || taskId === 'undefined' || taskId < 0) {
                            console.warn('阻止编辑无效ID的任务:', taskId);
                            event.preventDefault();
                            event.stopPropagation();
                            alert('无法编辑此任务：ID无效');
                            return false;
                        }
                    }, true);
                }
            });
            
            console.log('点击事件处理设置完成');
        }, 5000);
    }
    
    // 从按钮元素提取任务ID的函数
    function extractTaskIdFromButton(button) {
        // 尝试从按钮的data属性获取
        if (button.dataset && button.dataset.id) {
            return button.dataset.id;
        }
        
        // 尝试从最近的父元素获取
        let parentRow = button.closest('tr');
        if (parentRow) {
            // 尝试从行的data属性获取
            if (parentRow.dataset && parentRow.dataset.id) {
                return parentRow.dataset.id;
            }
            
            // 尝试从第一个单元格获取
            const firstCell = parentRow.querySelector('td:first-child');
            if (firstCell && firstCell.textContent) {
                return firstCell.textContent.trim();
            }
        }
        
        return null;
    }
    
    // 设置页面监听器，当DOM变化时重新应用点击处理
    const observer = new MutationObserver(function(mutations) {
        mutations.forEach(function(mutation) {
            if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
                setupClickHandlers();
            }
        });
    });
    
    // 等待 DOM 加载完成后再开始监听
    function initObserver() {
        if (document.body) {
            // 开始监听页面变化
            observer.observe(document.body, { childList: true, subtree: true });
            
            // 立即设置点击处理
            setupClickHandlers();
            
            console.log('前端数据修复脚本已加载');
        } else {
            console.warn('document.body 尚未加载，等待...');
            // 如果 body 还不存在，等待一下再尝试
            setTimeout(initObserver, 100);
        }
    }
    
    // 如果 DOM 已经加载完成，立即初始化；否则等待 DOMContentLoaded 事件
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initObserver);
    } else {
        initObserver();
    }
})(); 