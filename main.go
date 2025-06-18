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

	"clipboard-translate/ai"
	"clipboard-translate/config"
	"clipboard-translate/constants"
	"clipboard-translate/database"
	log "clipboard-translate/utils/log"
)

var (
	aiClient      ai.AIClient       // 替换 geminiClient
	staticDirPath string            // 全局变量存储静态文件目录路径
	db            database.Database // 数据库实例
	dbMutex       sync.Mutex        // 数据库操作互斥锁
)

// 翻译函数
func translateWithAI(ctx context.Context, text string) (string, error) {

	// 使用AI客户端进行翻译
	return aiClient.Translate(ctx, text)
}

// 添加历史项
func addHistoryItem(original, translated string) {
	// 确定翻译方向
	direction := "中 → 英"
	if !isChineseText(original) {
		direction = "英 → 中"
	}

	// 创建历史记录项
	newItem := &database.HistoryItem{
		Original:   original,
		Translated: translated,
		Timestamp:  time.Now(),
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		Direction:  direction,
	}

	// 使用互斥锁保护数据库操作
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// 添加到数据库
	if err := db.AddHistoryItem(newItem); err != nil {
		log.Error("添加历史记录失败: %v", err)
		return
	}

	// 检查是否需要清理旧记录
	maxHistory := config.GetConfig().System.MaxHistoryItems
	if maxHistory > 0 {
		count, err := db.GetHistoryCount()
		if err != nil {
			log.Error("获取历史记录数量失败: %v", err)
			return
		}

		if count > maxHistory {
			if err := db.PruneHistory(maxHistory); err != nil {
				log.Error("清理旧历史记录失败: %v", err)
			}
		}
	}
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
func triggerTranslation(ctx context.Context) {
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

	log.Info("开始翻译剪贴板内容... 方向: %s, 使用: %s", translationDirection, aiClient.GetName())
	translated, err := translateWithAI(ctx, content)
	if err != nil {
		log.Error("翻译失败: %v", err)
		translated = "翻译失败: " + err.Error()
	}

	notification := toast.Notification{
		AppID:   "剪贴板翻译",
		Title:   fmt.Sprintf("翻译结果 (%s - %s)", aiClient.GetName(), translationDirection),
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
func listenHotkey(ctx context.Context) {
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
			go triggerTranslation(ctx)
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
			dbMutex.Lock()
			defer dbMutex.Unlock()

			items, err := db.GetHistoryItems()
			if err != nil {
				log.Error("获取历史记录失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
				return
			}

			c.JSON(http.StatusOK, items)
		})

		// 清空历史记录
		api.POST("/clear", func(c *gin.Context) {
			dbMutex.Lock()
			defer dbMutex.Unlock()

			if err := db.ClearHistory(); err != nil {
				log.Error("清空历史记录失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "清空历史记录失败"})
				return
			}

			c.Status(http.StatusOK)
		})

		// 手动刷新剪贴板
		api.POST("/refresh", func(c *gin.Context) {
			go triggerTranslation(context.Background())
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
			oldDBType := config.GetConfig().Database.Type
			oldDBConn := config.GetConfig().Database.Connection

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

			// 检查数据库配置是否已更改
			newDBType := config.GetConfig().Database.Type
			newDBConn := config.GetConfig().Database.Connection
			if newDBType != oldDBType || newDBConn != oldDBConn {
				log.Info("数据库配置已更改，重新初始化数据库连接")

				dbMutex.Lock()
				defer dbMutex.Unlock()

				// 关闭旧连接
				if db != nil {
					if err := db.Close(); err != nil {
						log.Error("关闭数据库连接失败: %v", err)
					}
				}

				// 创建新连接
				var err error
				db, err = database.New(database.DBConfig{
					Type:       newDBType,
					Connection: newDBConn,
					MaxHistory: newConfig.System.MaxHistoryItems,
				})

				if err != nil {
					log.Error("创建数据库连接失败: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "重新初始化数据库失败"})
					return
				}

				// 初始化数据库
				if err := db.Initialize(); err != nil {
					log.Error("初始化数据库失败: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化数据库失败"})
					return
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
	r.StaticFile("/", filepath.Join(staticDirPath, "index.html")) // 根路径
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
			"./static",  // 开发环境
			"../static", // 打包后的相对路径
			path.Join(filepath.Dir(os.Args[0]), "static"),                    // 可执行文件同级目录
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
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, 0755)
	}

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

	// 初始化数据库连接
	dbConfig := database.DBConfig{
		Type:       config.GetConfig().Database.Type,
		Connection: config.GetConfig().Database.Connection,
		MaxHistory: config.GetConfig().System.MaxHistoryItems,
	}

	// 创建数据库实例
	db, err = database.New(dbConfig)
	if err != nil {
		log.Fatal("创建数据库连接失败: %v", err)
	}

	// 初始化数据库
	if err := db.Initialize(); err != nil {
		log.Fatal("初始化数据库失败: %v", err)
	}
	log.Info("数据库初始化成功: %s", dbConfig.Type)
	defer db.Close()

	// 初始化AI客户端
	apiConfig := config.GetConfig().API

	// 根据配置选择API密钥
	var apiKey string
	if apiConfig.UseEnvKey {
		apiKey = os.Getenv("AI_API_KEY")
	} else {
		apiKey = apiConfig.APIKey
	}

	// 创建AI客户端
	aiClientConfig := ai.AIConfig{
		Provider: apiConfig.Provider,
		APIKey:   apiKey,
		Model:    apiConfig.Model,
		BaseURL:  apiConfig.BaseURL,
	}

	aiClient, err = ai.NewAIClient(aiClientConfig)
	if err != nil {
		log.Fatal("AI客户端初始化失败: %v", err)
	}
	defer aiClient.Close()

	log.Info("AI客户端初始化成功: %s", aiClient.GetName())

	// 使用配置中的端口
	port := config.GetConfig().UI.Port

	// 启动热键监听
	go listenHotkey(context.Background())

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
