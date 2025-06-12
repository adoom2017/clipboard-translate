// static/js/config.js
// 键码映射表
const KEYS = [
    { value: 'A', label: 'A' },
    { value: 'B', label: 'B' },
    // ...其他字母
    { value: 'T', label: 'T' },
    { value: 'U', label: 'U' },
    // ...其他字母和功能键
];

// 填充键码选择器
function populateKeySelectors() {
    const selectors = ['#translate-key', '#showhide-key'];

    selectors.forEach(selector => {
        const select = document.querySelector(selector);
        KEYS.forEach(key => {
            const option = document.createElement('option');
            option.value = key.value;
            option.textContent = key.label;
            select.appendChild(option);
        });
    });
}

// 加载配置
async function loadConfig() {
    try {
        const response = await fetch('/api/config');
        if (!response.ok) {
            throw new Error('Failed to load config');
        }

        const config = await response.json();

        // 填充热键设置
        const translateHotkey = config.hotkeys.translate;
        document.getElementById('translate-ctrl').checked = translateHotkey.modifiers.includes('control');
        document.getElementById('translate-alt').checked = translateHotkey.modifiers.includes('alt');
        document.getElementById('translate-shift').checked = translateHotkey.modifiers.includes('shift');
        document.getElementById('translate-win').checked = translateHotkey.modifiers.includes('win');
        document.getElementById('translate-key').value = translateHotkey.key.toUpperCase();

        const showHideHotkey = config.hotkeys.showHide;
        document.getElementById('showhide-ctrl').checked = showHideHotkey.modifiers.includes('control');
        document.getElementById('showhide-alt').checked = showHideHotkey.modifiers.includes('alt');
        document.getElementById('showhide-shift').checked = showHideHotkey.modifiers.includes('shift');
        document.getElementById('showhide-win').checked = showHideHotkey.modifiers.includes('win');
        document.getElementById('showhide-key').value = showHideHotkey.key.toUpperCase();

        // API设置
        document.getElementById('use-env-key').checked = config.api.use_env_key;
        document.getElementById('gemini-key').value = config.api.gemini_key;
        document.getElementById('api-key-group').style.display = config.api.use_env_key ? 'none' : 'block';

        // 翻译设置
        document.getElementById('target-language').value = config.translation.target_language;
        document.getElementById('auto-translate').checked = config.translation.auto_translate;
        document.getElementById('show-notification').checked = config.translation.show_notification;

        // UI设置
        document.getElementById('port').value = config.ui.port;
        document.getElementById('start-minimized').checked = config.ui.start_minimized;
        document.getElementById('theme').value = config.ui.theme;

        // 系统设置
        document.getElementById('auto-start').checked = config.system.auto_start;
        document.getElementById('max-history').value = config.system.max_history_items;

    } catch (error) {
        console.error('Error loading config:', error);
    }
}

// 保存配置
async function saveConfig() {
    try {
        // 构建配置对象
        const config = {
            hotkeys: {
                translate: {
                    modifiers: [],
                    key: document.getElementById('translate-key').value.toLowerCase()
                },
                showHide: {
                    modifiers: [],
                    key: document.getElementById('showhide-key').value.toLowerCase()
                }
            },
            api: {
                gemini_key: document.getElementById('gemini-key').value,
                use_env_key: document.getElementById('use-env-key').checked
            },
            translation: {
                target_language: document.getElementById('target-language').value,
                auto_translate: document.getElementById('auto-translate').checked,
                show_notification: document.getElementById('show-notification').checked
            },
            ui: {
                port: parseInt(document.getElementById('port').value),
                start_minimized: document.getElementById('start-minimized').checked,
                theme: document.getElementById('theme').value
            },
            system: {
                auto_start: document.getElementById('auto-start').checked,
                max_history_items: parseInt(document.getElementById('max-history').value)
            }
        };

        // 添加修饰符
        if (document.getElementById('translate-ctrl').checked) config.hotkeys.translate.modifiers.push('control');
        if (document.getElementById('translate-alt').checked) config.hotkeys.translate.modifiers.push('alt');
        if (document.getElementById('translate-shift').checked) config.hotkeys.translate.modifiers.push('shift');
        if (document.getElementById('translate-win').checked) config.hotkeys.translate.modifiers.push('win');

        if (document.getElementById('showhide-ctrl').checked) config.hotkeys.showHide.modifiers.push('control');
        if (document.getElementById('showhide-alt').checked) config.hotkeys.showHide.modifiers.push('alt');
        if (document.getElementById('showhide-shift').checked) config.hotkeys.showHide.modifiers.push('shift');
        if (document.getElementById('showhide-win').checked) config.hotkeys.showHide.modifiers.push('win');

        // 发送到服务器
        const response = await fetch('/api/config', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(config)
        });

        if (!response.ok) {
            throw new Error('Failed to save config');
        }

        alert('设置已保存！部分设置可能需要重启应用才能生效。');
        window.location.href = '/';

    } catch (error) {
        console.error('Error saving config:', error);
        alert('保存设置失败: ' + error.message);
    }
}

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    populateKeySelectors();
    loadConfig();

    // 事件监听
    document.getElementById('save-btn').addEventListener('click', saveConfig);
    document.getElementById('cancel-btn').addEventListener('click', () => {
        window.location.href = '/';
    });

    // API密钥显示/隐藏逻辑
    document.getElementById('use-env-key').addEventListener('change', (e) => {
        document.getElementById('api-key-group').style.display = e.target.checked ? 'none' : 'block';
    });
});