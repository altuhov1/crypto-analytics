# Определяем переменные для удобства
BINARY_NAME = httpBack
CMD_PATH = ./cmd/httpBack
BUILD_DIR = ./bin
GO = go

# Цель по умолчанию
.DEFAULT_GOAL := help

## help: Выводит список всех доступных команд
help:
	@echo "Доступные команды:"
	@echo ""
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  \033[32m%-15s\033[0m %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
	@echo ""

## run: Запускает приложение в режиме разработки (с поддержкой .env)
run:
	@if command -v godotenv >/dev/null 2>&1; then \
		godotenv -f .env $(GO) run $(CMD_PATH); \
	else \
		echo "godotenv не установлен. Запуск без переменных .env"; \
		echo "Установите: go install github.com/joho/godotenv/cmd/godotenv@latest"; \
		$(GO) run $(CMD_PATH); \
	fi

## build: Собирает бинарный файл
build:
	mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Собранный файл: $(BUILD_DIR)/$(BINARY_NAME)"

## clean: Очищает скомпилированные файлы
clean:
	rm -rf $(BUILD_DIR)
	@echo "Директория $(BUILD_DIR) очищена"

## test: Запускает все тесты
test:
	$(GO) test -v ./...

## test-cover: Запускает тесты с отчетом о покрытии
test-cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Отчет о покрытии: coverage.html"

## deps: Устанавливает/обновляет все зависимости
deps:
	$(GO) mod tidy && $(GO) mod download
	@echo "Зависимости обновлены"

## install-tools: Устанавливает необходимые инструменты (godotenv)
install-tools:
	go install github.com/joho/godotenv/cmd/godotenv@latest
	@echo "godotenv установлен"

## lint: Запускает линтеры
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint не установлен. Установите: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
	fi

.PHONY: help run build clean test test-cover deps install-tools lint