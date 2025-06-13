package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/atotto/clipboard"
	"github.com/gin-gonic/gin"
	"github.com/go-toast/toast"
	gemini "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"clipboard-translate/config"
	"clipboard-translate/constants"
	log "clipboard-translate/utils/log"
)

// 历史记录项
type HistoryItem struct {
    Original   string `json:"original"`
    Translated string `json:"translated"`
    Timestamp  string `json:"timestamp"`
    ID         string `json:"id"`
    Direction  string `json:"direction"` // 翻译方向
}

var (
    historyItems  = make([]*HistoryItem, 0)
    historyMutex  sync.RWMutex
    geminiClient  *gemini.Client
    staticDirPath string // 全局变量存储静态文件目录路径
)

func translateWithGemini(ctx context.Context, client *gemini.Client, text string) (string, error) {
    // 检测输入文本语言
    isChinese := isChineseText(text)

    // 根据语言设置不同的翻译指令
    var systemPrompt string
    if isChinese {
        systemPrompt = "你是一个专业的中英文翻译助手，请将用户输入的中文内容翻译成英文，只需给出翻译结果，不要输出多余内容。"
    } else {
        systemPrompt = "你是一个专业的英中文翻译助手，请将用户输入的英文内容翻译成中文，只需给出翻译结果，不要输出多余内容。"
    }

    model := client.GenerativeModel("models/gemini-2.0-flash")

    systemInstruction := &gemini.Content{
        Parts: []gemini.Part{
            gemini.Text(systemPrompt),
        },
        Role: "system",
    }

    model.SystemInstruction = systemInstruction

    // 最大重试次数
    maxRetries := 3
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        // 如果不是第一次尝试，等待一小段时间
        if attempt > 0 {
            log.Info("翻译请求失败，正在进行第 %d 次重试...", attempt+1)
            time.Sleep(time.Duration(attempt) * time.Second)
        }

        resp, err := model.GenerateContent(ctx, gemini.Text(text))
        if err != nil {
            lastErr = err
            log.Error("翻译API调用失败: %v", err)
            continue
        }

        if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
            if responseText, ok := resp.Candidates[0].Content.Parts[0].(gemini.Text); ok {
                return string(responseText), nil
            }
        }
    }

    // 所有重试都失败了
    return "", fmt.Errorf("翻译失败，已重试%d次: %v", maxRetries, lastErr)
}

// 添加历史项
func addHistoryItem(original, translated string) {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    id := fmt.Sprintf("%d", time.Now().UnixNano())

    // 确定翻译方向
    direction := "中 → 英"
    if !isChineseText(original) {
        direction = "英 → 中"
    }

    newItem := &HistoryItem{
        Original:   original,
        Translated: translated,
        Timestamp:  timestamp,
        ID:         id,
        Direction:  direction,
    }

    historyMutex.Lock()
    historyItems = append(historyItems, newItem)
    historyMutex.Unlock()
}

// 注册热键
func registerHotKey() bool {
    // 从配置读取热键设置
    translateHotkey := config.GetConfig().Hotkeys["translate"]
    modifiers := config.ModifierFromString(translateHotkey.Modifiers)
    key := config.VirtualKeyFromString(translateHotkey.Key)

    user32 := syscall.NewLazyDLL("user32.dll")
    procRegisterHotKey := user32.NewProc("RegisterHotKey")
    ret, _, _ := procRegisterHotKey.Call(
        0,
        uintptr(constants.HOTKEY_ID),
        uintptr(modifiers),
        uintptr(key))

    return ret != 0
}

// 取消注册热键
func unregisterHotKey() {
    user32 := syscall.NewLazyDLL("user32.dll")
    procUnregisterHotKey := user32.NewProc("UnregisterHotKey")
    procUnregisterHotKey.Call(0, uintptr(constants.HOTKEY_ID))
}

// 触发翻译
func triggerTranslation(ctx context.Context, client *gemini.Client) {
    // 获取当前剪贴板内容
    content, err := clipboard.ReadAll()
    if err != nil {
        log.Error("读取剪贴板失败: %v", err)
        return
    }

    if content == "" {
        log.Warn("剪贴板内容为空")
        return
    }

    // 检测语言并记录
    isChinese := isChineseText(content)
    translationDirection := "中 → 英"
    if !isChinese {
        translationDirection = "英 → 中"
    }

    log.Info("开始翻译剪贴板内容... 方向: %s", translationDirection)
    translated, err := translateWithGemini(ctx, client, content)
    if err != nil {
        log.Error("翻译失败: %v", err)
        translated = "翻译失败: " + err.Error()
    }

    notification := toast.Notification{
        AppID:   "剪贴板翻译",
        Title:   fmt.Sprintf("翻译结果 (%s)", translationDirection),
        Message: translated,
    }

    if err := notification.Push(); err != nil {
        log.Error("发送通知失败: %v", err)
    } else {
        log.Info("已发送通知，内容: %s", translated)
    }

    // 添加到历史记录，包含翻译方向信息
    addHistoryItem(content, translated)
}

// 监听热键
func listenHotkey(ctx context.Context, client *gemini.Client) {
    // 注册热键
    if !registerHotKey() {
        log.Error("注册热键失败，按 Ctrl+Alt+T 触发翻译可能不可用")
    } else {
        log.Info("已注册热键: Ctrl+Alt+T 触发翻译")
        defer unregisterHotKey()
    }

    // 加载所需的 Windows API
    user32 := syscall.NewLazyDLL("user32.dll")
    procGetMessage := user32.NewProc("GetMessageW")
    procTranslateMessage := user32.NewProc("TranslateMessage")
    procDispatchMessage := user32.NewProc("DispatchMessageW")

    // 定义消息结构体
    type MSG struct {
        HWND   uintptr
        UINT   uint32
        WPARAM uintptr
        LPARAM uintptr
        Time   uint32
        Pt     struct {
            X int32
            Y int32
        }
    }

    // 消息循环
    var msg MSG
    for {
        // 获取Windows消息
        ret, _, _ := procGetMessage.Call(
            uintptr(unsafe.Pointer(&msg)),
            0,
            0,
            0,
        )

        if ret == 0 {
            // WM_QUIT 消息，退出循环
            break
        }

        // 检查是否是热键消息
        if msg.UINT == 0x0312 /* WM_HOTKEY */ && msg.WPARAM == uintptr(constants.HOTKEY_ID) {
            log.Info("检测到热键 Ctrl+Alt+T，开始翻译...")
            go triggerTranslation(ctx, client)
        }

        // 处理消息
        procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
        procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
    }
}

// 设置Gin路由
func setupRouter() *gin.Engine {
    // 设置为发布模式
    gin.SetMode(gin.ReleaseMode)

    // 创建带有默认中间件的路由
    r := gin.Default()

    // 日志中间件
    r.Use(func(c *gin.Context) {
        // 开始时间
        start := time.Now()

        // 处理请求
        c.Next()

        // 结束时间
        end := time.Now()

        // 执行时间
        latency := end.Sub(start)

        // 请求方式
        reqMethod := c.Request.Method

        // 请求路由
        reqURI := c.Request.RequestURI

        // 状态码
        statusCode := c.Writer.Status()

        // 请求IP
        clientIP := c.ClientIP()

        // 日志格式
        log.Info("[GIN] %v | %3d | %13v | %15s | %s %s",
            end.Format("2006/01/02 - 15:04:05"),
            statusCode,
            latency,
            clientIP,
            reqMethod,
            reqURI,
        )
    })

    // API 路由组
    api := r.Group("/api")
    {
        // 获取历史记录
        api.GET("/history", func(c *gin.Context) {
            historyMutex.RLock()
            defer historyMutex.RUnlock()

            c.JSON(http.StatusOK, historyItems)
        })

        // 清空历史记录
        api.POST("/clear", func(c *gin.Context) {
            historyMutex.Lock()
            historyItems = []*HistoryItem{}
            historyMutex.Unlock()

            c.Status(http.StatusOK)
        })

        // 手动刷新剪贴板
        api.POST("/refresh", func(c *gin.Context) {
            go triggerTranslation(context.Background(), geminiClient)
            c.Status(http.StatusOK)
        })

        // 获取配置
        api.GET("/config", func(c *gin.Context) {
            c.JSON(http.StatusOK, config.GetConfig())
        })

        // 更新配置
        api.POST("/config", func(c *gin.Context) {
            var newConfig config.Config
            if err := c.BindJSON(&newConfig); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置数据"})
                return
            }

            // 保存旧热键设置以检查是否需要重新注册
            oldTranslateHotkey := config.GetConfig().Hotkeys["translate"]

            // 保存到文件
            if err := config.SaveConfig(&newConfig); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败"})
                return
            }

            // 检查热键是否已更改
            newTranslateHotkey := config.GetConfig().Hotkeys["translate"]
            if !newTranslateHotkey.Equals(oldTranslateHotkey) {
                // 取消注册旧热键
                unregisterHotKey()

                // 注册新热键
                if !registerHotKey() {
                    log.Error("重新注册热键失败")
                }
            }

            c.Status(http.StatusOK)
        })

        // 健康检查端点
        api.GET("/health", func(c *gin.Context) {
            c.String(http.StatusOK, "OK")
        })
    }

    // 使用 StaticFile 处理单个文件
    r.StaticFile("/", filepath.Join(staticDirPath, "index.html"))       // 根路径
    r.StaticFile("/index.html", filepath.Join(staticDirPath, "index.html"))
    r.StaticFile("/config.html", filepath.Join(staticDirPath, "config.html"))

    // 静态文件服务 - 分别为CSS和JS添加路由
    r.Static("/css", filepath.Join(staticDirPath, "css"))
    r.Static("/js", filepath.Join(staticDirPath, "js"))

    return r
}

// 检测输入文本的主要语言
func isChineseText(text string) bool {
    // 统计中文和英文字符的数量
    var chineseCount, englishCount int

    for _, r := range text {
        // 中文字符范围
        if (r >= 0x4E00 && r <= 0x9FFF) || // 基本汉字
           (r >= 0x3400 && r <= 0x4DBF) || // 扩展A
           (r >= 0x20000 && r <= 0x2A6DF) || // 扩展B
           (r >= 0x2A700 && r <= 0x2B73F) || // 扩展C
           (r >= 0x2B740 && r <= 0x2B81F) || // 扩展D
           (r >= 0x2B820 && r <= 0x2CEAF) || // 扩展E
           (r >= 0xF900 && r <= 0xFAFF) || // 兼容汉字
           (r >= 0x2F800 && r <= 0x2FA1F) { // 兼容扩展
            chineseCount++
        }
        // 英文字符范围 (Basic Latin)
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
            englishCount++
        }
    }

    // 如果中文字符占比较大，则判断为中文文本
    if chineseCount > 0 && chineseCount >= englishCount/2 {
        return true
    }

    return false
}

// 找到静态目录
func getStaticDir() string {
    // 检查环境变量中是否指定了静态资源路径
    staticDir := os.Getenv("STATIC_PATH")
    if staticDir == "" {
        // 尝试在不同位置查找静态资源
        possiblePaths := []string{
            "./static",                              // 开发环境
            "../static",                             // 打包后的相对路径
            path.Join(filepath.Dir(os.Args[0]), "static"), // 可执行文件同级目录
            path.Join(os.Getenv("APPDATA"), "clipboard-translate", "static"), // Windows用户数据目录
        }

        for _, p := range possiblePaths {
            if _, err := os.Stat(p); err == nil {
                staticDir = p
                break
            }
        }

        if staticDir == "" {
            log.Fatal("找不到静态文件目录")
        }
    }

    return staticDir
}

// main函数
func main() {
    // 设置工作目录为可执行文件所在目录
    execDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
        log.Fatal("无法获取可执行文件目录: %v", err)
    }
    err = os.Chdir(execDir)
    if err != nil {
        log.Fatal("无法切换工作目录: %v", err)
    }
    log.Info("当前工作目录: %s", execDir)

    // 找到静态目录
    staticDirPath = getStaticDir()
    log.Info("使用静态文件目录: %s", staticDirPath)

    // 设置日志
    logDir := "./logs"
    logPath := filepath.Join(logDir, "clipboard-translate.log")

    // 设置日志配置
    log.SetLogConfig(log.LogConfig{
        Level:      log.INFO,
        Filename:   logPath,
        MaxSize:    10,
        MaxBackups: 3,
        Compress:   true,
    })

    // 加载配置
    if err := config.LoadConfig(); err != nil {
        log.Fatal("加载配置文件失败: %v", err)
    }
    log.Info("配置加载成功")

    // 初始化Gemini客户端
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        apiKey = ""
    }

    ctx := context.Background()
    client, err := gemini.NewClient(ctx, option.WithAPIKey(apiKey))
    if err != nil {
        log.Fatal("Gemini client 初始化失败: %v", err)
    }
    geminiClient = client

    // 检查API可用性及模型列表
    log.Info("正在检查可用的模型...")
    modelInfo := client.ListModels(ctx)
    for {
        m, err := modelInfo.Next()
        if err != nil {
            break
        }
        log.Info("可用模型: %s", m.Name)
    }

    // 使用配置中的端口
    port := config.GetConfig().UI.Port

    // 启动热键监听
    go listenHotkey(ctx, client)

    // 设置路由
    router := setupRouter()

    // 启动Web服务器
    log.Info("启动Web服务器，端口 %d...", port)
    log.Info("请访问 http://localhost:%d 查看剪贴板翻译历史", port)
    log.Info("按 Ctrl+Alt+T 触发翻译，将自动翻译当前剪贴板内容")
    if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
        log.Fatal("启动服务器失败: %v", err)
    }
}
