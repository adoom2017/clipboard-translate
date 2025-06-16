package config

import (
	"clipboard-translate/constants"
	"encoding/json"
	"os"
	"strings"
	"sync"
)

var (
	configInstance *Config
	configMutex    sync.RWMutex
	configFile     = "config.json"
)

// Config 应用配置
type Config struct {
	Hotkeys     map[string]HotkeyConfig `json:"hotkeys"`
	API         APIConfig                `json:"api"`
	Translation TranslationConfig       `json:"translation"`
	UI          UIConfig                 `json:"ui"`
	System      SystemConfig             `json:"system"`
	Database    DatabaseConfig           `json:"database"`
}

// HotkeyConfig 热键配置
type HotkeyConfig struct {
	Modifiers []string `json:"modifiers"`
	Key       string   `json:"key"`
}

// APIConfig API相关配置
type APIConfig struct {
	GeminiKey string `json:"gemini_key"`
	UseEnvKey bool   `json:"use_env_key"`
}

// TranslationConfig 翻译相关配置
type TranslationConfig struct {
	TargetLanguage   string `json:"target_language"`
	AutoTranslate    bool   `json:"auto_translate"`
	ShowNotification bool   `json:"show_notification"`
}

// UIConfig UI相关配置
type UIConfig struct {
	Port  int    `json:"port"`
	Theme string `json:"theme"`
}

// SystemConfig 系统相关配置
type SystemConfig struct {
	AutoStart       bool `json:"auto_start"`
	MaxHistoryItems int  `json:"max_history_items"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type        string `json:"type"`
	Connection  string `json:"connection"`
}

// LoadConfig 加载配置文件
func LoadConfig() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// 创建默认配置
		configInstance = &Config{
			Hotkeys: map[string]HotkeyConfig{
				"showHide": {
					Modifiers: []string{"control", "shift"},
					Key:       "",
				},
				"translate": {
					Modifiers: []string{"control", "alt"},
					Key:       "t",
				},
			},
			API: APIConfig{
				GeminiKey: "",
				UseEnvKey: true,
			},
			Translation: TranslationConfig{
				TargetLanguage:   "zh-CN",
				AutoTranslate:    false,
				ShowNotification: true,
			},
			UI: UIConfig{
				Port:  8080,
				Theme: "light",
			},
			System: SystemConfig{
				AutoStart:       true,
				MaxHistoryItems: 100,
			},
			Database: DatabaseConfig{
				Type:       "sqlite",
				Connection: "clipboard-translate.db",
			},
		}

		// 保存默认配置
		return SaveConfig(configInstance)
	}

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	// 解析JSON
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	// 设置默认值
	// UI配置
	if config.UI.Port == 0 {
		config.UI.Port = 8080
	}
	if config.UI.Theme == "" {
		config.UI.Theme = "light"
	}

	// 热键配置
	if config.Hotkeys == nil {
		config.Hotkeys = make(map[string]HotkeyConfig)
	}
	if _, ok := config.Hotkeys["translate"]; !ok {
		config.Hotkeys["translate"] = HotkeyConfig{
			Modifiers: []string{"control", "alt"},
			Key:       "t",
		}
	}
	if _, ok := config.Hotkeys["showHide"]; !ok {
		config.Hotkeys["showHide"] = HotkeyConfig{
			Modifiers: []string{"control", "shift"},
			Key:       "",
		}
	}

	// API配置
	if config.API.GeminiKey == "" && !config.API.UseEnvKey {
		config.API.UseEnvKey = true
	}

	// 翻译配置
	if config.Translation.TargetLanguage == "" {
		config.Translation.TargetLanguage = "zh-CN"
	}

	// 系统配置
	if config.System.MaxHistoryItems == 0 {
		config.System.MaxHistoryItems = 100
	}

	// 数据库配置
	if config.Database.Type == "" {
		config.Database.Type = "sqlite"
	}
	if config.Database.Connection == "" {
		config.Database.Connection = "clipboard-translate.db"
	}

	configInstance = &config
	return nil
}

// GetConfig 获取配置
func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if configInstance == nil {
		// 如果配置未加载，尝试加载
		configMutex.RUnlock()
		err := LoadConfig()
		if err != nil {
			// 加载失败，返回默认配置
			return &Config{
				Hotkeys: map[string]HotkeyConfig{
					"showHide": {
						Modifiers: []string{"control", "shift"},
						Key:       "",
					},
					"translate": {
						Modifiers: []string{"control", "alt"},
						Key:       "t",
					},
				},
				API: APIConfig{
					GeminiKey: "",
					UseEnvKey: true,
				},
				Translation: TranslationConfig{
					TargetLanguage:   "zh-CN",
					AutoTranslate:    false,
					ShowNotification: true,
				},
				UI: UIConfig{
					Port:  8080,
					Theme: "light",
				},
				System: SystemConfig{
					AutoStart:       true,
					MaxHistoryItems: 100,
				},
				Database: DatabaseConfig{
					Type:        "sqlite",
					Connection:  "clipboard-translate.db",
				},
			}
		}
		configMutex.RLock()
	}

	return configInstance
}

// SaveConfig 保存配置
func SaveConfig(config *Config) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 序列化为JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		return err
	}

	configInstance = config
	return nil
}

// 将虚拟键码字符串转换为键码
func VirtualKeyFromString(key string) uint16 {
	key = strings.ToUpper(key)
	switch key {
	// 字母键
	case "A":
		return constants.VK_A
	case "B":
		return constants.VK_B
	case "C":
		return constants.VK_C
	case "D":
		return constants.VK_D
	case "E":
		return constants.VK_E
	case "F":
		return constants.VK_F
	case "G":
		return constants.VK_G
	case "H":
		return constants.VK_H
	case "I":
		return constants.VK_I
	case "J":
		return constants.VK_J
	case "K":
		return constants.VK_K
	case "L":
		return constants.VK_L
	case "M":
		return constants.VK_M
	case "N":
		return constants.VK_N
	case "O":
		return constants.VK_O
	case "P":
		return constants.VK_P
	case "Q":
		return constants.VK_Q
	case "R":
		return constants.VK_R
	case "S":
		return constants.VK_S
	case "T":
		return constants.VK_T
	case "U":
		return constants.VK_U
	case "V":
		return constants.VK_V
	case "W":
		return constants.VK_W
	case "X":
		return constants.VK_X
	case "Y":
		return constants.VK_Y
	case "Z":
		return constants.VK_Z

	// 数字键
	case "0":
		return constants.VK_0
	case "1":
		return constants.VK_1
	case "2":
		return constants.VK_2
	case "3":
		return constants.VK_3
	case "4":
		return constants.VK_4
	case "5":
		return constants.VK_5
	case "6":
		return constants.VK_6
	case "7":
		return constants.VK_7
	case "8":
		return constants.VK_8
	case "9":
		return constants.VK_9

	// 功能键
	case "INSERT":
		return constants.VK_INSERT
	case "DELETE":
		return constants.VK_DELETE

	default:
		// 如果是单个字符，尝试直接转换
		if len(key) == 1 {
			return uint16(key[0])
		}
		return 0
	}
}

// 将修饰符字符串数组转换为修饰符标志
func ModifierFromString(modifiers []string) uint16 {
	var result uint16
	for _, mod := range modifiers {
		mod = strings.ToLower(mod)
		switch mod {
		case "alt":
			result |= constants.MOD_ALT
		case "control", "ctrl":
			result |= constants.MOD_CONTROL
		case "shift":
			result |= constants.MOD_SHIFT
		case "win":
			result |= constants.MOD_WIN
		}
	}
	return result
}

// Equals 判断两个 HotkeyConfig 实例是否相等
func (h HotkeyConfig) Equals(other HotkeyConfig) bool {
	// 比较键名
	if h.Key != other.Key {
		return false
	}

	// 比较修饰符数量
	if len(h.Modifiers) != len(other.Modifiers) {
		return false
	}

	// 创建修饰符集合以便进行无序比较
	selfMods := make(map[string]struct{}, len(h.Modifiers))
	for _, mod := range h.Modifiers {
		selfMods[strings.ToLower(mod)] = struct{}{}
	}

	// 检查对方的所有修饰符是否都存在
	for _, mod := range other.Modifiers {
		if _, exists := selfMods[strings.ToLower(mod)]; !exists {
			return false
		}
	}

	// 所有检查都通过，认为相等
	return true
}