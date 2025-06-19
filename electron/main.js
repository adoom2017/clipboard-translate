const { app, BrowserWindow, Tray, Menu, shell, ipcMain, dialog, nativeImage } = require('electron');
const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs');
const AutoLaunch = require('auto-launch');

// 全局变量
let mainWindow;
let tray;
let goProcess;
let isQuitting = false;

// 应用配置
const config = {
  port: 8080,
  isDev: process.env.NODE_ENV === 'development',
  appName: '剪贴板翻译'
};

// 自启动配置
const autoLauncher = new AutoLaunch({
  name: config.appName,
  path: app.getPath('exe')
});

// 获取资源路径
function getResourcePath(relativePath) {
  if (config.isDev) {
    return path.join(__dirname, '..', relativePath);
  } else {
    return path.join(process.resourcesPath, 'app', relativePath);
  }
}

// 创建主窗口
function createMainWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js'),
      webSecurity: true
    },
    icon: path.join(__dirname, 'assets', 'icon.png'),
    title: config.appName,
    show: false, // 初始隐藏，等待ready-to-show事件
    autoHideMenuBar: true // 隐藏菜单栏
  });

  // 窗口准备就绪时显示
  mainWindow.once('ready-to-show', () => {
    // 检查Go服务是否已启动
    checkGoService().then(() => {
      mainWindow.show();
      if (config.isDev) {
        mainWindow.webContents.openDevTools();
      }
    }).catch(err => {
      console.error('Go service check failed:', err);
      showErrorDialog('启动失败', 'Go后端服务启动失败，请检查配置。');
    });
  });

  // 处理窗口关闭事件
  mainWindow.on('close', (event) => {
    if (!isQuitting) {
      event.preventDefault();
      mainWindow.hide();

      // 首次最小化时显示提示
      if (!app.isQuitting) {
        showTrayNotification('应用已最小化到系统托盘', '点击托盘图标可重新打开窗口');
      }
    }
  });

  // 窗口关闭后清理
  mainWindow.on('closed', () => {
    mainWindow = null;
  });

  // 处理外部链接
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: 'deny' };
  });

  // 加载应用页面
  loadAppPage();
}

// 加载应用页面
async function loadAppPage() {
  try {
    await checkGoService();
    mainWindow.loadURL(`http://localhost:${config.port}`);
  } catch (error) {
    console.error('Failed to load app page:', error);
    // 加载错误页面
    mainWindow.loadFile(path.join(__dirname, 'error.html'));
  }
}

// 检查Go服务状态
function checkGoService(maxAttempts = 30) {
  return new Promise((resolve, reject) => {
    let attempts = 0;

    const check = () => {
      attempts++;

      const http = require('http');
      const req = http.get(`http://localhost:${config.port}/api/health`, (res) => {
        if (res.statusCode === 200) {
          resolve();
        } else {
          if (attempts < maxAttempts) {
            setTimeout(check, 1000);
          } else {
            reject(new Error('Service health check failed'));
          }
        }
      });

      req.on('error', () => {
        if (attempts < maxAttempts) {
          setTimeout(check, 1000);
        } else {
          reject(new Error('Service not available'));
        }
      });

      req.setTimeout(2000, () => {
        req.destroy();
        if (attempts < maxAttempts) {
          setTimeout(check, 1000);
        } else {
          reject(new Error('Service timeout'));
        }
      });
    };

    check();
  });
}

// 启动Go后端服务
function startGoService() {
  const exePath = getResourcePath('clipboard-translate.exe');

  console.log('Starting Go service from:', exePath);

  if (!fs.existsSync(exePath)) {
    showErrorDialog('错误', `找不到应用程序文件: ${exePath}`);
    app.quit();
    return;
  }

  // 设置环境变量
  const env = { ...process.env };

  // 启动Go进程
  goProcess = spawn(exePath, [], {
    cwd: path.dirname(exePath),
    env: env,
    windowsHide: true
  });

  goProcess.stdout.on('data', (data) => {
    console.log(`Go服务输出: ${data}`);
  });

  goProcess.stderr.on('data', (data) => {
    console.error(`Go服务错误: ${data}`);
  });

  goProcess.on('error', (err) => {
    console.error('启动Go服务失败:', err);
    showErrorDialog('启动失败', `无法启动后端服务: ${err.message}`);
  });

  goProcess.on('close', (code) => {
    console.log(`Go服务进程退出，代码: ${code}`);
    if (!isQuitting && code !== 0) {
      showErrorDialog('服务异常', `后端服务异常退出，退出代码: ${code}`);
    }
  });
}

// 创建系统托盘
function createTray() {
  const iconPath = path.join(__dirname, 'assets', 'icon.png');
  const trayIcon = nativeImage.createFromPath(iconPath);

  tray = new Tray(trayIcon.resize({ width: 16, height: 16 }));

  const contextMenu = Menu.buildFromTemplate([
    {
      label: '显示主窗口',
      click: () => {
        if (mainWindow) {
          mainWindow.show();
          mainWindow.focus();
        }
      }
    },
    {
      label: '刷新翻译',
      click: () => {
        triggerTranslation();
      }
    },
    { type: 'separator' },
    {
      label: '开机自启',
      type: 'checkbox',
      checked: false,
      click: (menuItem) => {
        toggleAutoLaunch(menuItem.checked);
      }
    },
    { type: 'separator' },
    {
      label: '关于',
      click: () => {
        showAboutDialog();
      }
    },
    {
      label: '退出',
      click: () => {
        quitApp();
      }
    }
  ]);

  tray.setToolTip(config.appName);
  tray.setContextMenu(contextMenu);

  // 点击托盘图标显示/隐藏窗口
  tray.on('click', () => {
    if (mainWindow) {
      if (mainWindow.isVisible()) {
        mainWindow.hide();
      } else {
        mainWindow.show();
        mainWindow.focus();
      }
    }
  });

  // 检查自启动状态
  autoLauncher.isEnabled().then((isEnabled) => {
    const menu = tray.getContextMenu();
    const autoStartItem = menu.items.find(item => item.label === '开机自启');
    if (autoStartItem) {
      autoStartItem.checked = isEnabled;
    }
  });
}

// 切换自启动
function toggleAutoLaunch(enable) {
  if (enable) {
    autoLauncher.enable().then(() => {
      console.log('自启动已启用');
    }).catch(err => {
      console.error('启用自启动失败:', err);
    });
  } else {
    autoLauncher.disable().then(() => {
      console.log('自启动已禁用');
    }).catch(err => {
      console.error('禁用自启动失败:', err);
    });
  }
}

// 触发翻译
function triggerTranslation() {
  if (goProcess) {
    // 发送API请求触发翻译
    const http = require('http');
    const postData = '';

    const options = {
      hostname: 'localhost',
      port: config.port,
      path: '/api/refresh',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData)
      }
    };

    const req = http.request(options, (res) => {
      console.log('Translation triggered');
    });

    req.on('error', (err) => {
      console.error('Failed to trigger translation:', err);
    });

    req.write(postData);
    req.end();
  }
}

// 显示托盘通知
function showTrayNotification(title, body) {
  if (tray) {
    tray.displayBalloon({
      iconType: 'info',
      title: title,
      content: body
    });
  }
}

// 显示错误对话框
function showErrorDialog(title, content) {
  dialog.showErrorBox(title, content);
}

// 显示关于对话框
function showAboutDialog() {
  dialog.showMessageBox(mainWindow, {
    type: 'info',
    title: '关于',
    message: config.appName,
    detail: `版本: ${app.getVersion()}\n\n智能剪贴板翻译工具，支持多种AI翻译服务。\n\n使用 Ctrl+Alt+T 快捷键快速翻译剪贴板内容。`,
    buttons: ['确定']
  });
}

// 退出应用
function quitApp() {
  isQuitting = true;

  // 终止Go进程
  if (goProcess && !goProcess.killed) {
    console.log('正在终止Go服务...');
    goProcess.kill('SIGTERM');

    // 如果进程没有正常退出，强制终止
    setTimeout(() => {
      if (!goProcess.killed) {
        goProcess.kill('SIGKILL');
      }
    }, 5000);
  }

  app.quit();
}

// IPC处理器
ipcMain.handle('get-version', () => {
  return app.getVersion();
});

ipcMain.handle('show-save-dialog', async () => {
  const result = await dialog.showSaveDialog(mainWindow, {
    filters: [
      { name: 'JSON文件', extensions: ['json'] }
    ]
  });
  return result;
});

ipcMain.handle('show-open-dialog', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    filters: [
      { name: 'JSON文件', extensions: ['json'] }
    ],
    properties: ['openFile']
  });
  return result;
});

// 应用事件处理
app.whenReady().then(() => {
  // 启动Go服务
  startGoService();

  // 创建主窗口和托盘
  createMainWindow();
  createTray();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createMainWindow();
    } else if (mainWindow) {
      mainWindow.show();
    }
  });
});

app.on('window-all-closed', () => {
  // macOS上保持应用运行
  if (process.platform !== 'darwin') {
    quitApp();
  }
});

app.on('before-quit', () => {
  isQuitting = true;
});

app.on('will-quit', (event) => {
  if (goProcess && !goProcess.killed) {
    event.preventDefault();
    goProcess.kill();
    setTimeout(() => {
      app.quit();
    }, 2000);
  }
});

// 阻止多实例运行
const gotTheLock = app.requestSingleInstanceLock();

if (!gotTheLock) {
  app.quit();
} else {
  app.on('second-instance', () => {
    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore();
      mainWindow.focus();
      mainWindow.show();
    }
  });
}