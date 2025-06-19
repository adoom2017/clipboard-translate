// Electron渲染进程增强功能
(function() {
  'use strict';
  
  // 等待DOM加载完成
  document.addEventListener('DOMContentLoaded', function() {
    
    // 添加版本信息到页面
    if (window.electronAPI) {
      window.electronAPI.getVersion().then(version => {
        const footer = document.querySelector('footer .footer-info');
        if (footer) {
          const versionInfo = document.createElement('p');
          versionInfo.textContent = `版本 ${version}`;
          versionInfo.style.fontSize = '0.8rem';
          versionInfo.style.color = 'var(--text-light)';
          footer.appendChild(versionInfo);
        }
      });
    }
    
    // 增强错误处理
    window.addEventListener('error', function(e) {
      console.error('应用错误:', e.error);
    });
    
    // 增强网络错误处理
    window.addEventListener('unhandledrejection', function(e) {
      console.error('未处理的Promise拒绝:', e.reason);
    });
    
    // 桌面应用特定的UI调整
    addDesktopEnhancements();
  });
  
  // 添加桌面应用特定的增强功能
  function addDesktopEnhancements() {
    // 添加窗口控制按钮样式
    const style = document.createElement('style');
    style.textContent = `
      .electron-app {
        user-select: none;
      }
      
      .electron-app input,
      .electron-app textarea,
      .electron-app [contenteditable] {
        user-select: text;
      }
      
      /* macOS特定样式 */
      .platform-darwin header {
        padding-top: 2rem; /* 为macOS标题栏留空间 */
      }
      
      /* Windows特定样式 */
      .platform-win32 {
        /* Windows特定样式 */
      }
      
      /* 滚动条样式 */
      .electron-app ::-webkit-scrollbar {
        width: 8px;
      }
      
      .electron-app ::-webkit-scrollbar-track {
        background: var(--background-color);
      }
      
      .electron-app ::-webkit-scrollbar-thumb {
        background: var(--border-color);
        border-radius: 4px;
      }
      
      .electron-app ::-webkit-scrollbar-thumb:hover {
        background: var(--text-light);
      }
    `;
    document.head.appendChild(style);
    
    // 添加桌面应用提示
    const header = document.querySelector('header');
    if (header) {
      const appBadge = document.createElement('div');
      appBadge.style.cssText = `
        position: absolute;
        top: 10px;
        right: 10px;
        background: var(--primary-color);
        color: white;
        padding: 2px 8px;
        border-radius: 10px;
        font-size: 0.7rem;
        font-weight: 500;
      `;
      appBadge.textContent = '桌面版';
      header.style.position = 'relative';
      header.appendChild(appBadge);
    }
  }
  
})();