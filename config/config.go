package config

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	"clipboard-translate/constants"
)

// 配置文件路径
const (
	CONFIG_FILE = "./config.json"
)

// 配置结构
type HotkeyConfig struct {
	Modifiers []string `json:"modifiers"`
	Key      string    `json:"key"`
}

type ApiConfig struct {
	GeminiKey string `json:"gemini_key"`
	UseEnvKey bool   `json:"use_env_key"`
}

type TranslationConfig struct {
	TargetLanguage  string `json:"target_language"`
	AutoTranslate   bool   `json:"auto_translate"`
	ShowNotification bool  `json:"show_notification"`
}

type UIConfig struct {
	Port          int    `json:"port"`
	StartMinimized bool   `json:"start_minimized"`
	Theme         string `json:"theme"`
}

type SystemConfig struct {
	AutoStart      bool `json:"auto_start"`
	MaxHistoryItems int `json:"max_history_items"`
}

type Config struct {
	Hotkeys     map[string]HotkeyConfig `json:"hotkeys"`
	Api         ApiConfig              `json:"api"`
	Translation TranslationConfig      `json:"translation"`
	UI          UIConfig               `json:"ui"`
	System      SystemConfig           `json:"system"`
}

var (
	instance *Config
	once     sync.Once
	mu       sync.RWMutex
)

// GetConfig 获取配置实例(单例模式)
func GetConfig() *Config {
    once.Do(func() {

        // 设置默认配置
        config := Config{
            Hotkeys: map[string]HotkeyConfig{
                "translate": {
                    Modifiers: []string{"control", "alt"},
                    Key:       "t",
                },
                "showHide": {
                    Modifiers: []string{"control", "shift"},
                    Key:       "c",
                },
            },
            Api: ApiConfig{
                GeminiKey: "",
                UseEnvKey: true,
            },
            Translation: TranslationConfig{
                TargetLanguage:  "zh-CN",
                AutoTranslate:   false,
                ShowNotification: true,
            },
            UI: UIConfig{
                Port:          8080,
                StartMinimized: false,
                Theme:         "light",
            },
            System: SystemConfig{
                AutoStart:      true,
                MaxHistoryItems: 100,
            },
        }

        // 检查配置文件是否存在
        if _, err := os.Stat(CONFIG_FILE); !os.IsNotExist(err) {
            // 读取配置文件
            data, err := os.ReadFile(CONFIG_FILE)
            if err == nil {
                // 解析配置
                json.Unmarshal(data, &config)
            }
        } else {
            // 创建默认配置文件
            data, _ := json.MarshalIndent(config, "", "  ")
            os.WriteFile(CONFIG_FILE, data, 0644)
        }

        instance = &config
    })
    return instance
}

// LoadConfig 从配置文件加载配置
func LoadConfig() error {
    mu.Lock()
    defer mu.Unlock()

    // 读取配置文件
    data, err := os.ReadFile(CONFIG_FILE)
    if err != nil {
        return err
    }

    // 解析配置到临时变量
    var newConfig Config
    if err := json.Unmarshal(data, &newConfig); err != nil {
        return err
    }

    // 更新单例实例
    instance = &newConfig

    return nil
}

// SaveConfig 保存配置到文件
func SaveConfig(newConfig *Config) error {
    mu.Lock()
    defer mu.Unlock()

    // 更新单例实例
    *instance = *newConfig

    // 转换为JSON
    data, err := json.MarshalIndent(*instance, "", "  ")
    if err != nil {
        return err
    }

    // 写入文件
    return os.WriteFile(CONFIG_FILE, data, 0644)
}

// 将虚拟键码字符串转换为键码
func VirtualKeyFromString(key string) uint16 {
    key = strings.ToUpper(key)
    switch key {
    // 字母键
    case "A": return constants.VK_A
    case "B": return constants.VK_B
    case "C": return constants.VK_C
    case "D": return constants.VK_D
    case "E": return constants.VK_E
    case "F": return constants.VK_F
    case "G": return constants.VK_G
    case "H": return constants.VK_H
    case "I": return constants.VK_I
    case "J": return constants.VK_J
    case "K": return constants.VK_K
    case "L": return constants.VK_L
    case "M": return constants.VK_M
    case "N": return constants.VK_N
    case "O": return constants.VK_O
    case "P": return constants.VK_P
    case "Q": return constants.VK_Q
    case "R": return constants.VK_R
    case "S": return constants.VK_S
    case "T": return constants.VK_T
    case "U": return constants.VK_U
    case "V": return constants.VK_V
    case "W": return constants.VK_W
    case "X": return constants.VK_X
    case "Y": return constants.VK_Y
    case "Z": return constants.VK_Z

    // 数字键
    case "0": return constants.VK_0
    case "1": return constants.VK_1
    case "2": return constants.VK_2
    case "3": return constants.VK_3
    case "4": return constants.VK_4
    case "5": return constants.VK_5
    case "6": return constants.VK_6
    case "7": return constants.VK_7
    case "8": return constants.VK_8
    case "9": return constants.VK_9

    // 功能键
    case "INSERT": return constants.VK_INSERT
    case "DELETE": return constants.VK_DELETE

    default:
        // 如果是单个字符，尝试直接转换
        if len(key) == 1 {
            return uint16(key[0])
        }
        return 0
    }
}

// 将修饰符字符串转换为修饰符标志
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