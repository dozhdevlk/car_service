# Используем официальный образ Go 1.24.2
FROM golang:1.24.2-alpine
# Устанавливаем рабочую директорию в контейнере
WORKDIR /app
# Копируем файлы Go модуля и зависимости
COPY go.mod go.sum ./
# Загружаем зависимости
RUN go mod tidy
# Копируем весь исходный код в контейнер
COPY . .
# Компилируем приложение
RUN go build -o main .
# Открываем порт, на котором будет работать ваше приложение
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
