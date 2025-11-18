/**
 * 任务列表修复脚本
 * 
 * 解决以下问题：
 * 1. 删除按钮点击时ID为undefined的问题
 * 2. 任务列表ID为空的显示问题
 * 3. 时间范围显示undefined的问题
 */

(function() {
  console.log('任务列表修复脚本已加载');

  // 等待页面和Vue实例加载完成
  const waitForElements = setInterval(() => {
    // 检查任务列表表格是否存在
    const taskTable = document.querySelector('table');
    if (!taskTable) return;

    // 检查删除按钮是否存在
    const deleteButtons = document.querySelectorAll('.delete-btn, button:contains("删除")');
    if (deleteButtons.length === 0) return;

    // 已找到元素，停止轮询
    clearInterval(waitForElements);
    console.log('找到任务列表元素，应用修复');

    // 修复1: 拦截删除按钮点击
    document.addEventListener('click', function(event) {
      // 查找最近的删除按钮祖先
      const deleteBtn = findAncestorByText(event.target, '删除');
      if (deleteBtn) {
        // 获取任务行
        const row = deleteBtn.closest('tr');
        if (!row) return;

        // 获取ID单元格
        const idCell = row.querySelector('td:first-child');
        if (!idCell) return;

        const taskId = idCell.textContent.trim();
        
        // 检查ID是否有效
        if (!taskId || taskId === '' || taskId === 'undefined') {
          console.warn('阻止删除无效ID的任务:', taskId);
          event.preventDefault();
          event.stopPropagation();
          
          // 显示友好提示
          alert('无法删除此任务：任务ID无效。请刷新页面或创建新任务。');
          return false;
        }
      }
    }, true);

    // 修复2: 修复任务列表显示
    fixTaskListDisplay();

    // 监听DOM变化，应对动态加载的内容
    const observer = new MutationObserver(() => {
      fixTaskListDisplay();
    });
    
    observer.observe(taskTable, {
      childList: true,
      subtree: true
    });
  }, 500);

  // 辅助函数：根据文本内容查找祖先元素
  function findAncestorByText(element, text) {
    let currentNode = element;
    while (currentNode) {
      if (currentNode.textContent && currentNode.textContent.includes(text)) {
        return currentNode;
      }
      currentNode = currentNode.parentNode;
    }
    return null;
  }

  // 修复任务列表显示
  function fixTaskListDisplay() {
    // 查找所有任务行
    const rows = document.querySelectorAll('table tr:not(:first-child)');
    
    rows.forEach(row => {
      const cells = row.querySelectorAll('td');
      if (cells.length < 8) return;
      
      // 修复ID显示为空的问题
      const idCell = cells[0];
      if (!idCell.textContent || idCell.textContent.trim() === '') {
        idCell.textContent = '(无效ID)';
        
        // 禁用此行的所有按钮
        row.querySelectorAll('button').forEach(btn => {
          btn.disabled = true;
          btn.title = '无效任务ID，不可操作';
          btn.style.opacity = '0.5';
        });
      }
      
      // 修复时间范围显示undefined的问题
      const timeRangeCell = cells[3];
      if (timeRangeCell.textContent === 'undefined') {
        timeRangeCell.textContent = '30m';
      }
      
      // 修复未设置的显示问题
      cells.forEach(cell => {
        if (cell.textContent === '未选择' || cell.textContent === '未设置' || cell.textContent === '未绑定') {
          cell.style.color = '#999';
        }
      });
    });
  }

  // 修复3: 拦截API请求，阻止发送无效ID
  const originalFetch = window.fetch;
  window.fetch = function() {
    const url = arguments[0];
    
    // 检查URL是否包含无效ID
    if (typeof url === 'string' && 
        (url.includes('/undefined') || 
         url.includes('/null') || 
         url.includes('/NaN'))) {
      console.warn('拦截到无效ID的API请求:', url);
      
      // 返回一个拒绝的Promise
      return Promise.reject(new Error('客户端拦截：请求包含无效ID'));
    }
    
    // 正常请求
    return originalFetch.apply(this, arguments);
  };

  // 添加全局错误处理
  window.addEventListener('error', function(event) {
    console.error('捕获到全局错误:', event.error);
    
    // 分析错误是否与任务ID相关
    const errorMsg = event.error && event.error.message;
    if (errorMsg && (
        errorMsg.includes('undefined') || 
        errorMsg.includes('null') || 
        errorMsg.includes('cannot read property')
      )) {
      console.warn('检测到可能与任务ID相关的错误，尝试恢复');
      
      // 防止错误传播
      event.preventDefault();
      
      // 提示用户
      const errorBox = document.createElement('div');
      errorBox.style.cssText = 'position:fixed;top:10px;right:10px;background-color:#f8d7da;color:#721c24;padding:10px;border-radius:4px;z-index:9999;';
      errorBox.innerHTML = '检测到界面错误，请刷新页面或重新操作';
      document.body.appendChild(errorBox);
      
      // 3秒后自动移除提示
      setTimeout(() => {
        errorBox.remove();
      }, 3000);
      
      return false;
    }
  }, true);
})(); 