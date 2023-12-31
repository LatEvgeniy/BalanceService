# Используем официальный образ Golang в качестве базового образа
FROM golang:latest

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для эффективного кэширования зависимостей
COPY go.mod .
COPY go.sum .

# Загружаем зависимости
RUN go mod download

# Копируем файлы проекта в текущую директорию
COPY . .

# Собираем Go-приложение
RUN go build -o main .

# Указываем команду для запуска приложения при старте контейнера
CMD ["./main"]