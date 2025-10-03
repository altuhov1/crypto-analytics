FROM golang:1.25-alpine

WORKDIR /app

# Копируем всё что нужно для зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Создаем папку bin (как в твоем Makefile)
RUN mkdir -p bin

# Собираем точно так же как твой Makefile
RUN go build -o bin/httpBack ./cmd/httpBack

# Копируем необходимые файлы
COPY static ./static
COPY storage ./storage
COPY bd ./bd

EXPOSE 8080

# Запускаем из правильного места
CMD ["./bin/httpBack"]