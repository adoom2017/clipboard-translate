const { contextBridge, ipcRenderer } = require('electron');

// 暴露安全的API给渲染进程
contextBridge.exposeInMainWorld('electronAPI', {
  // 获取应用版本
  getVersion: () => ipcRenderer.invoke('get-version'),
  
  // 文件对话框
  showSaveDialog: () => ipcRenderer.invoke('show-save-dialog'),
  showOpenDialog: () => ipcRenderer.invoke('show-open-dialog'),
  
  // 平台信息
  platform: process.platform,
  
  // 应用信息
  isElectron: true
});

// 当DOM准备就绪时添加Electron特定的样式和功能
window.addEventListener('DOMContentLoaded', () => {
  // 添加Electron标识类
  document.body.classList.add('electron-app');
  
  // 添加平台特定类
  document.body.classList.add(`platform-${process.platform}`);
  
  // 禁用右键菜单（可选）
  document.addEventListener('contextmenu', (e) => {
    if (process.env.NODE_ENV !== 'development') {
      e.preventDefault();
    }
  });
  
  // 禁用某些快捷键（可选）
  document.addEventListener('keydown', (e) => {
    if (process.env.NODE_ENV !== 'development') {
      // 禁用F12开发者工具
      if (e.key === 'F12') {
        e.preventDefault();
      }
      // 禁用Ctrl+Shift+I
      if (e.ctrlKey && e.shiftKey && e.key === 'I') {
        e.preventDefault();
      }
    }
  });
});