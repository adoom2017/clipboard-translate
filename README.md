# 剪贴板翻译工具 (Clipboard Translate)

这是一个基于 Go 和 Electron 的桌面应用，旨在提供一个智能、高效的剪贴板翻译体验。通过全局快捷键，用户可以快速翻译剪贴板中的文本，并通过简洁的界面查看翻译历史。

## ✨ 功能特性

*   **剪贴板翻译**: 监控剪贴板内容，通过快捷键快速翻译。
*   **多 AI 服务商支持**: 支持 Gemini, OpenAI, Claude, 和 Ollama 等多种翻译引擎。
*   **自定义快捷键**: 用户可以根据自己的习惯在 `config.json` 中设置翻译快捷键。
*   **翻译历史**: 在本地通过 Web 界面查看和管理翻译历史记录。
*   **跨平台**: 使用 Go 作为后端，Electron 作为 GUI，理论上可以打包成多平台应用。
*   **开机自启**: 可配置是否在系统启动时自动运行。


## 🚀 使用方法

### 1. 配置

在首次运行前，请在项目根目录创建一个 `config.json` 文件。你可以参考下面的模板进行配置。

```json
{
  "hotkeys": {
    "translate": {
      "modifiers": ["control", "alt"],
      "key": "t"
    }
  },
  "api": {
    "provider": "gemini",
    "api_key": "YOUR_API_KEY",
    "model": "gemini-pro",
    "base_url": "",
    "use_env_key": false
  },
  "translation": {
    "target_language": "zh-CN",
    "auto_translate": false,
    "show_notification": true
  },
  "ui": {
    "port": 8080,
    "theme": "light"
  },
  "system": {
    "auto_start": true,
    "max_history_items": 100
  },
  "database": {
    "type": "sqlite",
    "connection": "clipboard-translate.db"
  }
}
```

**配置说明:**

*   `hotkeys`: 设置全局快捷键。
    *   `translate`: 翻译功能的快捷键。
*   `api`: 配置 AI 翻译服务。
    *   `provider`: AI 服务商 (`gemini`, `openai`, `claude`, `ollama`)。
    *   `api_key`: 你的 API 密钥。如果 `use_env_key` 为 `true`，则会从环境变量读取。
    *   `model`: 使用的具体模型。
*   `ui`: Web 界面的配置。
    *   `port`: 访问翻译历史的本地端口。

### 2. 构建和运行

本项目使用 `Makefile` 进行构建管理。

*   **构建应用**:
    ```bash
    make build
    ```
    该命令会编译 Go 后端，并将所有必要的静态资源复制到 `build` 目录。

*   **运行应用**:
    ```bash
    make run
    ```
    此命令会先执行构建，然后启动 `build` 目录下的可执行文件。

*   **清理构建目录**:
    ```bash
    make clean
    ```

### 3. 快捷键

*   **翻译**: 默认快捷键为 `Ctrl + Alt + T`。复制文本后，按下此快捷键即可进行翻译。
*   **查看历史**: 打开浏览器并访问 `http://localhost:8080` (端口可在 `config.json` 中修改)。

## 📦 打包分发

本项目使用 `Electron` 将 Go 后端和 Web UI 打包成桌面应用。

1.  **安装依赖**:
    进入 `electron` 目录并安装 npm 依赖。
    ```bash
    cd electron
    npm install
    ```

2.  **打包应用**:
    在 `electron` 目录中，运行以下命令之一来为不同平台打包：
    ```bash
    # For Windows
    npm run build-win

    # For macOS
    npm run build-mac

    # For Linux
    npm run build-linux
    ```
    打包前，脚本会自动执行 `make package` 来准备 Go 后端程序和资源。打包好的应用会存放在 `electron/dist` 目录下。
