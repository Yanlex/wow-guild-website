# Официальный образ Go
FROM golang:1.22.1-alpine

# Рабочая директория в контейнере
WORKDIR /app

# Копировать go.mod и go.sum
COPY go.mod go.sum ./

# Загрузить зависимости
RUN go mod download

# Копируем проект
COPY assets/class /root/assets/class
COPY cmd ./cmd
COPY configs ./configs
COPY deployments ./deployments
COPY internal ./internal

# Сборка приложения
RUN go build -o ./ ./cmd/app/main.go
# RUN go build -o ./deployments/db/ ./deployments/db/deploy.go

# Запуск приложения
CMD ["./main"]