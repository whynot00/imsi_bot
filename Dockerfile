# ЭТАП СБОРКИ
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник. Указываем путь к нашему main.go в папке cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

# ЭТАП ЗАПУСКА
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из билдера
COPY --from=builder /app/main .

# КРИТИЧНО: Копируем папку static, иначе фронтенд не загрузится
COPY --from=builder /app/static ./static

# Если приложение ищет миграции или данные в этих папках, их тоже стоит скопировать:
# COPY --from=builder /app/migrations ./migrations

EXPOSE 8080 
# (или тот порт, который у вас прописан в config.go)

CMD ["./main"]