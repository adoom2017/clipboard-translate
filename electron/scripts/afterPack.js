const fs = require('fs');
const path = require('path');

module.exports = async function(context) {
  console.log('开始语言包清理...');

  const { electronPlatformName, appOutDir } = context;

  if (electronPlatformName === 'win32') {
    // Windows平台处理
    const localesDir = path.join(appOutDir, 'locales');

    if (fs.existsSync(localesDir)) {
      const allLocales = fs.readdirSync(localesDir);

      // 保留的语言包：英文、中文简体、中文繁体
      const keepLocales = [
        'en-US.pak',
        'zh-CN.pak',
        'zh-TW.pak'
      ];

      let removedCount = 0;
      let totalSize = 0;

      allLocales.forEach(locale => {
        if (!keepLocales.includes(locale)) {
          const localePath = path.join(localesDir, locale);
          const stats = fs.statSync(localePath);
          totalSize += stats.size;

          fs.unlinkSync(localePath);
          removedCount++;
          console.log(`已删除语言包: ${locale}`);
        }
      });

      console.log(`语言包清理完成! 删除了 ${removedCount} 个文件，节省空间 ${(totalSize / 1024 / 1024).toFixed(2)} MB`);
    }

    // 删除不必要的DLL文件
    const dllsToRemove = [
      'libGLESv2.dll',
      'libEGL.dll',
      'vk_swiftshader.dll',
      'vulkan-1.dll',
      'd3dcompiler_47.dll'
    ];

    let dllRemovedSize = 0;
    dllsToRemove.forEach(dll => {
      const dllPath = path.join(appOutDir, dll);
      if (fs.existsSync(dllPath)) {
        const stats = fs.statSync(dllPath);
        dllRemovedSize += stats.size;
        fs.unlinkSync(dllPath);
        console.log(`已删除DLL: ${dll}`);
      }
    });

    if (dllRemovedSize > 0) {
      console.log(`DLL清理完成! 节省空间 ${(dllRemovedSize / 1024 / 1024).toFixed(2)} MB`);
    }
  } else if (electronPlatformName === 'darwin') {
    // macOS平台处理
    const localesDir = path.join(appOutDir, 'ClipboardTranslate.app', 'Contents', 'Frameworks', 'Electron Framework.framework', 'Versions', 'A', 'Resources');

    if (fs.existsSync(localesDir)) {
      const allLocales = fs.readdirSync(localesDir).filter(file => file.endsWith('.lproj'));

      const keepLocales = [
        'en.lproj',
        'zh_CN.lproj',
        'zh_TW.lproj'
      ];

      allLocales.forEach(locale => {
        if (!keepLocales.includes(locale)) {
          const localePath = path.join(localesDir, locale);
          fs.rmSync(localePath, { recursive: true, force: true });
          console.log(`已删除macOS语言包: ${locale}`);
        }
      });
    }
  }

  console.log('构建后处理完成!');
};