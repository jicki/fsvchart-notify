/**
 * 前端补丁脚本 - 修复任务ID为null的问题
 * 
 * 使用方法:
 * 1. 将此文件放在web目录下
 * 2. 在index.html中添加引用: <script src="fix-null-task.js"></script>
 */

(function() {
  // 监听所有AJAX响应
  const originalOpen = XMLHttpRequest.prototype.open;
  const originalSend = XMLHttpRequest.prototype.send;
  
  // 拦截XHR打开
  XMLHttpRequest.prototype.open = function() {
    this._url = arguments[1];
    return originalOpen.apply(this, arguments);
  };
  
  // 拦截XHR发送
  XMLHttpRequest.prototype.send = function() {
    const xhr = this;
    
    // 添加响应处理器
    xhr.addEventListener('load', function() {
      if (xhr._url && xhr._url.includes('/api/push_task') && xhr.status === 200) {
        try {
          // 尝试解析响应
          const response = JSON.parse(xhr.responseText);
          
          // 修复响应数据
          if (response) {
            // 确保data属性存在且是数组
            if (!response.data) {
              response.data = [];
            } else if (!Array.isArray(response.data)) {
              response.data = [response.data];
            }
            
            // 过滤掉ID为null的任务
            response.data = response.data.filter(task => task && task.id != null);
            
            // 替换原始响应
            Object.defineProperty(xhr, 'responseText', {
              get: function() {
                return JSON.stringify(response);
              }
            });
          }
        } catch (e) {
          console.error('修复任务数据时出错:', e);
        }
      }
    });
    
    return originalSend.apply(this, arguments);
  };
  
  // 全局错误处理器
  window.addEventListener('error', function(event) {
    // 捕获"Cannot read properties of null (reading 'id')"错误
    if (event.error && event.error.message && event.error.message.includes("Cannot read properties of null (reading 'id')")) {
      console.warn('捕获到空ID错误，正在尝试恢复...');
      event.preventDefault();
      
      // 你可以在这里添加自定义的恢复逻辑，比如刷新组件或页面
      setTimeout(() => {
        location.reload();
      }, 1000);
      
      return false;
    }
  }, true);
  
  console.log('任务ID修复脚本已加载');
})(); 