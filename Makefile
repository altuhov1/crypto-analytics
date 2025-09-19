# Определяем переменные для удобства
BINARY_NAME=httpBack
CMD_PATH=./cmd/httpBack
BUILD_DIR=./bin
GO=go

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

## run: Запускает приложение в режиме разработки
run:
	$(GO) run $(CMD_PATH)

## build: Собирает бинарный файл
build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

## clean: Очищает скомпилированные файлы
clean:
	rm -rf $(BUILD_DIR)

## test: Запускает все тесты
test:
	$(GO) test ./...

## deps: Устанавливает/обновляет все зависимости
deps:
	$(GO) mod tidy && $(GO) mod download

.PHONY: help run build clean test deps