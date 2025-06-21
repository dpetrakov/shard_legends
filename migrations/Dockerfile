# Dockerfile для выполнения миграций БД
# Использует golang-migrate/migrate для управления миграциями PostgreSQL

FROM alpine:3.19

# Установка migrate CLI и PostgreSQL client
RUN apk add --no-cache \
    curl \
    postgresql-client && \
    # Определение архитектуры системы
    ARCH=$(case $(uname -m) in \
        x86_64) echo "amd64" ;; \
        aarch64) echo "arm64" ;; \
        armv7l) echo "armv7" ;; \
        *) echo "amd64" ;; \
    esac) && \
    echo "Detected architecture: $ARCH" && \
    # Скачивание правильной версии migrate
    curl -L "https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-${ARCH}.tar.gz" | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Создание рабочей директории
WORKDIR /migrations

# Копирование всех миграций
COPY . /migrations/

# Создание wrapper скрипта для работы с переменными окружения
RUN echo '#!/bin/sh' > /migrate.sh && \
    echo '# Wrapper script для migrate с поддержкой переменных окружения' >> /migrate.sh && \
    echo '' >> /migrate.sh && \
    echo '# Проверка наличия DATABASE_URL' >> /migrate.sh && \
    echo 'if [ -z "$DATABASE_URL" ]; then' >> /migrate.sh && \
    echo '    echo "Error: DATABASE_URL environment variable is not set"' >> /migrate.sh && \
    echo '    exit 1' >> /migrate.sh && \
    echo 'fi' >> /migrate.sh && \
    echo '' >> /migrate.sh && \
    echo '# Выполнение migrate с переданными аргументами' >> /migrate.sh && \
    echo 'exec migrate -path /migrations -database "$DATABASE_URL" "$@"' >> /migrate.sh && \
    chmod +x /migrate.sh

# Точка входа через wrapper скрипт
ENTRYPOINT ["/migrate.sh"]

# Команда по умолчанию
CMD ["up"]