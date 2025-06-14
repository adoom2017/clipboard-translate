SHELL := cmd.exe
.SHELLFLAGS := /c

# ��Ŀ��������
PROJECT_NAME := clipboard-translate
VERSION := 1.0.0
MAIN_FILE := main.go
BUILD_DIR := build
STATIC_DIR := static
TARGET_EXE := $(BUILD_DIR)/$(PROJECT_NAME).exe

# Go����������
GO := go
GOOS := windows
GOARCH := amd64
# GO_BUILD_FLAGS := -ldflags="-s -w -H windowsgui -X main.Version=$(VERSION)"
GO_BUILD_FLAGS := -ldflags="-s -w -X main.Version=$(VERSION)"

.PHONY: all clean build run debug release help resources

# Ĭ��Ŀ��
all: build

# ������Ŀ¼
clean:
	@echo ����������Ŀ¼...
	@if exist $(BUILD_DIR) rd /s /q $(BUILD_DIR)

# ��������Ŀ¼�ṹ
init:
	@echo Creating directory structure...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@if not exist $(BUILD_DIR)\logs mkdir $(BUILD_DIR)\logs
	@if not exist $(BUILD_DIR)\$(STATIC_DIR) mkdir $(BUILD_DIR)\$(STATIC_DIR)
	@if not exist $(BUILD_DIR)\$(STATIC_DIR)\css mkdir $(BUILD_DIR)\$(STATIC_DIR)\css
	@if not exist $(BUILD_DIR)\$(STATIC_DIR)\js mkdir $(BUILD_DIR)\$(STATIC_DIR)\js

# ������Դ�ļ�
resources: init
	@echo ���ڸ��ƾ�̬��Դ�ļ�...
	@echo ���ڸ��� HTML �ļ�...
	@copy /Y $(STATIC_DIR)\*.html $(BUILD_DIR)\$(STATIC_DIR)\ > nul
	@echo ���ڸ��� CSS �ļ�...
	@copy /Y $(STATIC_DIR)\css\*.css $(BUILD_DIR)\$(STATIC_DIR)\css\ > nul
	@echo ���ڸ��� JavaScript �ļ�...
	@copy /Y $(STATIC_DIR)\js\*.js $(BUILD_DIR)\$(STATIC_DIR)\js\ > nul
	@echo ���ڸ��������ļ�...
	@copy /Y config.json $(BUILD_DIR)\ > nul
	@echo ��̬��Դ�ļ��������!

# �������
build: resources
	@echo ���ڱ���Ӧ�ó���...
	@set GOOS=$(GOOS)& set GOARCH=$(GOARCH)& $(GO) build $(GO_BUILD_FLAGS) -o $(TARGET_EXE) $(MAIN_FILE)
	@if %ERRORLEVEL% EQU 0 (echo ����ɹ�������������: $(TARGET_EXE)) else (echo ����ʧ��! & exit /b 1)

# ���г���
run: build
	@echo �������г���...
	@cd $(BUILD_DIR) && $(PROJECT_NAME).exe

# ���Թ����汾
debug: GO_BUILD_FLAGS := -gcflags="all=-N -l"
debug: clean build
	@echo �ѹ������԰汾

# �����汾
release: clean
	@echo ���ڹ��������汾...
	@$(MAKE) build
	@echo �����汾�������!

# ����һ���򵥵�README
create-readme: init
	@echo # $(PROJECT_NAME) > $(BUILD_DIR)\README.txt
	@echo. >> $(BUILD_DIR)\README.txt
	@echo Version: $(VERSION) >> $(BUILD_DIR)\README.txt
	@echo. >> $(BUILD_DIR)\README.txt
	@echo ʹ�÷���: >> $(BUILD_DIR)\README.txt
	@echo 1. ���� $(PROJECT_NAME).exe >> $(BUILD_DIR)\README.txt
	@echo 2. ʹ�� Ctrl+Alt+T ��ݼ��������� >> $(BUILD_DIR)\README.txt
	@echo 3. ͨ����������� http://localhost:8080 �鿴������ʷ >> $(BUILD_DIR)\README.txt
	@echo. >> $(BUILD_DIR)\README.txt
	@echo ע������: >> $(BUILD_DIR)\README.txt
	@echo - ȷ�� config.json �����ڳ���ͬһĿ¼�� >> $(BUILD_DIR)\README.txt
	@echo - ��־�ļ������� logs Ŀ¼�� >> $(BUILD_DIR)\README.txt
	@echo �Ѵ��� README �ļ�

# ȫ�����
package: resources build create-readme
	@echo ���������!

# ������Ϣ
help:
	@echo ��������:
	@echo   make          - Ĭ���������Ӧ�ó���
	@echo   make clean    - ������Ŀ¼
	@echo   make build    - ����Ӧ�ó���
	@echo   make run      - ���벢����Ӧ�ó���
	@echo   make debug    - �������԰汾(���Ż�)
	@echo   make release  - ���������汾
	@echo   make package  - ����������Ӧ�ð�
	@echo   make help     - ��ʾ�˰�����Ϣ