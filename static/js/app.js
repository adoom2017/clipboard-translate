let selectedId = null;

// 加载历史记录
function loadHistory() {
    fetch('/api/history')
        .then(response => response.json())
        .then(data => {
            const historyList = document.getElementById('historyList');
            historyList.innerHTML = '';

            data.reverse().forEach(item => {
                const div = document.createElement('div');
                div.className = 'history-item';
                if (item.id === selectedId) {
                    div.className += ' selected';
                }
                div.innerHTML = `
                    <div class="timestamp">${item.timestamp}</div>
                    <div>${item.original.substring(0, 30)}${item.original.length > 30 ? '...' : ''}</div>
                `;
                div.onclick = () => {
                    document.querySelectorAll('.history-item').forEach(el => {
                        el.classList.remove('selected');
                    });
                    div.classList.add('selected');
                    selectedId = item.id;
                    displayItem(item);
                };
                historyList.appendChild(div);
            });

            // 自动选择最新的项目
            if (data.length > 0 && !selectedId) {
                selectedId = data[0].id;
                displayItem(data[0]);
                const firstItem = document.querySelector('.history-item');
                if (firstItem) {
                    firstItem.classList.add('selected');
                }
            }
        });
}

// 显示选中的项目内容
function displayItem(item) {
    document.getElementById('originalText').textContent = item.original;
    document.getElementById('translatedText').textContent = item.translated;

    // 显示翻译方向
    const directionElem = document.getElementById('translationDirection');
    if (directionElem) {
        directionElem.textContent = item.direction || "自动检测";
    }
}

// 复制文本到剪贴板
function copyTextToClipboard(text, button) {
    navigator.clipboard.writeText(text).then(() => {
        button.textContent = "已复制";
        button.classList.add('copy-success');

        // 1.5秒后恢复按钮状态
        setTimeout(() => {
            button.textContent = "复制";
            button.classList.remove('copy-success');
        }, 1500);
    }).catch(err => {
        console.error('无法复制文本: ', err);
    });
}

// 清空历史记录
document.getElementById('clearBtn').addEventListener('click', () => {
    fetch('/api/clear', { method: 'POST' })
        .then(() => {
            loadHistory();
            document.getElementById('originalText').textContent = '';
            document.getElementById('translatedText').textContent = '';
            selectedId = null;
        });
});

// 刷新剪贴板
document.getElementById('refreshBtn').addEventListener('click', () => {
    fetch('/api/refresh', { method: 'POST' })
        .then(() => {
            setTimeout(loadHistory, 1000);
        });
});

// 添加到 app.js 末尾
function showToast(message) {
  const toast = document.getElementById('copyToast');
  toast.textContent = message;
  toast.classList.add('show');

  setTimeout(() => {
    toast.classList.remove('show');
  }, 2000);
}

// 修改复制按钮的事件处理
document.getElementById('copyOriginal').addEventListener('click', () => {
  const text = document.getElementById('originalText').textContent;
  navigator.clipboard.writeText(text).then(() => {
    showToast('已复制原文到剪贴板');
  });
});

document.getElementById('copyTranslated').addEventListener('click', () => {
  const text = document.getElementById('translatedText').textContent;
  navigator.clipboard.writeText(text).then(() => {
    showToast('已复制译文到剪贴板');
  });
});

// 初始加载
loadHistory();

// 定期刷新
setInterval(loadHistory, 5000);