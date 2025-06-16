# Project configuration
PROJECT_NAME := clipboard-translate
VERSION := 1.0.0
MAIN_FILE := main.go
BUILD_DIR := build
STATIC_DIR := static
TARGET_EXE := $(BUILD_DIR)/$(PROJECT_NAME).exe

# Go compiler settings
GO := go
GOOS := windows
GOARCH := amd64
# GO_BUILD_FLAGS := -ldflags="-s -w -H windowsgui -X main.Version=$(VERSION)"
GO_BUILD_FLAGS := -ldflags="-s -w -X main.Version=$(VERSION)"

.PHONY: all clean build run debug release help resources

# Default target
all: build

# Clean build directory
clean:
	@echo Cleaning build directory...
	@if exist $(BUILD_DIR) rd /s /q $(BUILD_DIR)

# Create directory structure
init:
	@echo Creating directory structure...
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	@if not exist $(BUILD_DIR)\logs mkdir $(BUILD_DIR)\logs
	@if not exist $(BUILD_DIR)\$(STATIC_DIR) mkdir $(BUILD_DIR)\$(STATIC_DIR)
	@if not exist $(BUILD_DIR)\$(STATIC_DIR)\css mkdir $(BUILD_DIR)\$(STATIC_DIR)\css
	@if not exist $(BUILD_DIR)\$(STATIC_DIR)\js mkdir $(BUILD_DIR)\$(STATIC_DIR)\js

# Copy resource files
resources: init
	@echo Copying static resource files...
	@echo Copying HTML files...
	@copy /Y $(STATIC_DIR)\*.html $(BUILD_DIR)\$(STATIC_DIR)\ > nul
	@echo Copying CSS files...
	@copy /Y $(STATIC_DIR)\css\*.css $(BUILD_DIR)\$(STATIC_DIR)\css\ > nul
	@echo Copying JavaScript files...
	@copy /Y $(STATIC_DIR)\js\*.js $(BUILD_DIR)\$(STATIC_DIR)\js\ > nul
	@echo Copying configuration file...
	@copy /Y config.json $(BUILD_DIR)\ > nul
	@echo Static resource files copied successfully!

# Compile program
build: resources
	@echo Compiling application...
	@set GOOS=$(GOOS)& set GOARCH=$(GOARCH)& $(GO) build $(GO_BUILD_FLAGS) -o $(TARGET_EXE) $(MAIN_FILE)
	@if %ERRORLEVEL% EQU 0 (echo Compilation successful! Program generated at: $(TARGET_EXE)) else (echo Compilation failed! & exit /b 1)

# Run program
run: build
	@echo Running program...
	@cd $(BUILD_DIR) && $(PROJECT_NAME).exe

# Debug build version
debug: GO_BUILD_FLAGS := -gcflags="all=-N -l"
debug: clean build
	@echo Debug version built

# Release version
release: clean
	@echo Building release version...
	@$(MAKE) build
	@echo Release version build complete!

# Create a simple README
create-readme: set-utf8 init
	@echo # $(PROJECT_NAME) > $(BUILD_DIR)\README.txt
	@echo. >> $(BUILD_DIR)\README.txt
	@echo Version: $(VERSION) >> $(BUILD_DIR)\README.txt
	@echo. >> $(BUILD_DIR)\README.txt
	@echo Usage: >> $(BUILD_DIR)\README.txt
	@echo 1. Run $(PROJECT_NAME).exe >> $(BUILD_DIR)\README.txt
	@echo 2. Use Ctrl+Alt+T hotkey to trigger translation >> $(BUILD_DIR)\README.txt
	@echo 3. Visit http://localhost:8080 in browser to view translation history >> $(BUILD_DIR)\README.txt
	@echo. >> $(BUILD_DIR)\README.txt
	@echo Notes: >> $(BUILD_DIR)\README.txt
	@echo - Ensure config.json exists in the same directory as the program >> $(BUILD_DIR)\README.txt
	@echo - Log files are saved in the logs directory >> $(BUILD_DIR)\README.txt
	@echo README file created

# Package everything
package: set-utf8 resources build create-readme
	@echo Program packaging complete!

# Help information
help: set-utf8
	@echo Available commands:
	@echo   make          - Default command, builds the application
	@echo   make clean    - Clean build directory
	@echo   make build    - Compile application
	@echo   make run      - Compile and run application
	@echo   make debug    - Build debug version (no optimization)
	@echo   make release  - Build release version
	@echo   make package  - Build complete application package
	@echo   make help     - Display this help information