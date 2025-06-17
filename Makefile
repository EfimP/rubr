# Значения по умолчанию для целевой ОС и архитектуры
TARGET_OS   ?= windows
TARGET_ARCH ?= amd64

# Директория для сборки
BIN_DIR := ./bin
# Базовое имя выходного файла
OUTPUT_NAME := rubr-gui

# Общие флаги линковщика
LDFLAGS_COMMON := -w -s

# Компилятор и суффикс в зависимости от ОС и архитектуры
ifeq ($(TARGET_OS),windows)
  ifeq ($(TARGET_ARCH),amd64)
    CC_COMPILER  := x86_64-w64-mingw32-gcc
    CXX_COMPILER := x86_64-w64-mingw32-g++
    OUTPUT_SUFFIX := .exe
  else
    $(error Неподдерживаемая архитектура $(TARGET_ARCH) для ОС $(TARGET_OS))
  endif
else ifeq ($(TARGET_OS),linux)
  ifeq ($(TARGET_ARCH),amd64)
    CC_COMPILER  := gcc
    CXX_COMPILER := g++
    OUTPUT_SUFFIX :=
  else ifeq ($(TARGET_ARCH),arm64)
    CC_COMPILER  := aarch64-linux-gnu-gcc
    CXX_COMPILER := aarch64-linux-gnu-g++
    OUTPUT_SUFFIX :=
  else
    $(error Неподдерживаемая архитектура $(TARGET_ARCH) для ОС $(TARGET_OS))
  endif
else
  $(error Неподдерживаемая ОС $(TARGET_OS))
endif

# Полный путь к выходному файлу
OUTPUT_PATH := $(BIN_DIR)/$(OUTPUT_NAME)-$(TARGET_OS)-$(TARGET_ARCH)$(OUTPUT_SUFFIX)

# Цель для генерации встроенных ресурсов
bundle:
	@echo "Генерация встроенных ресурсов..."
	@mkdir -p cmd/fyne-client
	fyne bundle -o cmd/fyne-client/bundled.go bin/logo/hse_logo.svg

# Цель для сборки без обфускации
build: bundle
	@echo "Сборка для OS=$(TARGET_OS) ARCH=$(TARGET_ARCH)"
	@echo "Компилятор CC: $(CC_COMPILER)"
	@echo "Компилятор CXX: $(CXX_COMPILER)"
	@echo "Выходной файл: $(OUTPUT_PATH)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=1 GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CC=$(CC_COMPILER) CXX=$(CXX_COMPILER) go build -ldflags "$(LDFLAGS_COMMON)" -o $(OUTPUT_PATH) ./cmd/fyne-client

# Цель для сборки с обфускацией через Garble
build-obfuscated: bundle
	@echo "Сборка с обфускацией для OS=$(TARGET_OS) ARCH=$(TARGET_ARCH)"
	@echo "Компилятор CC: $(CC_COMPILER)"
	@echo "Компилятор CXX: $(CXX_COMPILER)"
	@echo "Выходной файл: $(OUTPUT_PATH)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=1 GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CC=$(CC_COMPILER) CXX=$(CXX_COMPILER) garble -seed=random -literals -tiny build -ldflags "$(LDFLAGS_COMMON)" -o $(OUTPUT_PATH) ./cmd/fyne-client

# Цель для очистки
clean:
	@echo "Очистка..."
	@rm -rf $(BIN_DIR)/* rubr-gui$(OUTPUT_SUFFIX) cmd/fyne-client/bundled.go

# Цель по умолчанию
all: build
.PHONY: all build build-obfuscated bundle clean