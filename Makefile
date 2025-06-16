SHELL := cmd.exe
.SHELLFLAGS := /c

# 项目基本配置
PROJECT_NAME := clipboard-translate
VERSION := 1.0.0
MAIN_FILE := main.go
BUILD_DIR := build
STATIC_DIR := static
TARGET_EXE := $(BUILD_DIR)/$(PROJECT_NAME).exe

# Go编译器设置
GO := go
GOOS := windows
GOARCH := amd64
# GO_BUILD_FLAGS := -ldflags="-s -w -H windowsgui -X main.Version=$(VERSION)"
GO_BUILD_FLAGS := -ldflags="-s -w -X main.Version=$(VERSION)"

.PHONY: all clean build run debug release help resources

# 默认目标
all: build

# 清理构建目录
clean:
	@echo 正在清理构建目录...
	@if exist $(BUILD_DIR) rd /s /q $(BUILD_DIR)

# 创建构建目录结构
init:
	@echo Creating directory structure...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@if not exist $(BUILD_DIR)\logs mkdir $(BUILD_DIR)\logs
	@if not exist $(BUILD_DIR)\$(STATIC_DIR) mkdir $(BUILD_DIR)\$(STATIC_DIR)
	@if not exist $(BUILD_DIR)\$(STATIC_DIR)\css mkdir $(BUILD_DIR)\$(STATIC_DIR)\css
	@if not exist $(BUILD_DIR)\$(STATIC_DIR)\js mkdir $(BUILD_DIR)\$(STATIC_DIR)\js

# 复制资源文件
resources: init
	@echo 正在复制静态资源文件...
	@echo 正在复制 HTML 文件...
	@copy /Y $(STATIC_DIR)\*.html $(BUILD_DIR)\$(STATIC_DIR)\ > nul
	@echo 正在复制 CSS 文件...
	@copy /Y $(STATIC_DIR)\css\*.css $(BUILD_DIR)\$(STATIC_DIR)\css\ > nul
	@echo 正在复制 JavaScript 文件...
	@copy /Y $(STATIC_DIR)\js\*.js $(BUILD_DIR)\$(STATIC_DIR)\js\ > nul
	@echo 正在复制配置文件...
	@copy /Y config.json $(BUILD_DIR)\ > nul
	@echo 静态资源文件复制完成!

# 编译程序
build: resources
	@echo 正在编译应用程序...
	@set GOOS=$(GOOS)& set GOARCH=$(GOARCH)& $(GO) build $(GO_BUILD_FLAGS) -o $(TARGET_EXE) $(MAIN_FILE)
	@if %ERRORLEVEL% EQU 0 (echo 编译成功！程序已生成: $(TARGET_EXE)) else (echo 编译失败! & exit /b 1)

# 运行程序
run: build
	@echo 正在运行程序...
	@cd $(BUILD_DIR) && $(PROJECT_NAME).exe

# 调试构建版本
debug: GO_BUILD_FLAGS := -gcflags="all=-N -l"
debug: clean build
	@echo 已构建调试版本

# 发布版本
release: clean
	@echo 正在构建发布版本...
	@$(MAKE) build
	@echo 发布版本构建完成!

# 创建一个简单的README
create-readme: init
	@echo # $(PROJECT_NAME) > $(BUILD_DIR)\README.txt
	@echo. >> $(BUILD_DIR)\README.txt
	@echo Version: $(VERSION) >> $(BUILD_DIR)\README.txt
	@echo. >> $(BUILD_DIR)\README.txt
	@echo 使用方法: >> $(BUILD_DIR)\README.txt
	@echo 1. 运行 $(PROJECT_NAME).exe >> $(BUILD_DIR)\README.txt
	@echo 2. 使用 Ctrl+Alt+T 快捷键触发翻译 >> $(BUILD_DIR)\README.txt
	@echo 3. 通过浏览器访问 http://localhost:8080 查看翻译历史 >> $(BUILD_DIR)\README.txt
	@echo. >> $(BUILD_DIR)\README.txt
	@echo 注意事项: >> $(BUILD_DIR)\README.txt
	@echo - 确保 config.json 存在于程序同一目录下 >> $(BUILD_DIR)\README.txt
	@echo - 日志文件保存在 logs 目录中 >> $(BUILD_DIR)\README.txt
	@echo 已创建 README 文件

# 全部打包
package: resources build create-readme
	@echo 程序打包完成!

# 帮助信息
help:
	@echo 可用命令:
	@echo   make          - 默认命令，构建应用程序
	@echo   make clean    - 清理构建目录
	@echo   make build    - 编译应用程序
	@echo   make run      - 编译并运行应用程序
	@echo   make debug    - 构建调试版本(不优化)
	@echo   make release  - 构建发布版本
	@echo   make package  - 构建完整的应用包
	@echo   make help     - 显示此帮助信息