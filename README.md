# MyFacebook API

## Настройки окружения

* LOG_LEVEL - Уровень логирования в приложении. По умолчанию info
* SERVICE_NAME - Имя сервиса. По умолчанию myfacebook
* VERSION - Версия сервиса. По умолчанию version_not_set
* HTTP_INT_PORT - HTTP порт приложения. По умолчанию 9090
* REQUEST_HEADER_MAX_SIZE - максимальный размер header для входящих запросов. По умолчанию 10000 байт.
* REQUEST_READ_HEADER_TIMEOUT_MILLISECONDS - максимальное время отпущенное клиенту на чтение header в мс. По умолчанию
  2000мс.
* DB_HOST - Адрес хоста для подключения к БД. По умолчанию localhost
* DB_PORT - Порт для подключения к БД. По умолчанию 5432
* DB_USERNAME - Имя пользователя БД. По умолчанию postgres
* DB_PASSWORD - Пароль к БД. По умолчанию secret
* DB_NAME - Название БД. По умолчанию myfacebook
* DB_DRIVER_NAME - Драйвер БД. По умолчанию postgres
* DB_SSL_MODE - Режим работы ssl для postgres. По умолчанию disable
* DB_MAX_OPEN_CONNECTIONS - Число максимально одновременно открытых подключений. По умолчанию: 10
* MYFACEBOOK_DIALOG_API_BASE_URL - Адрес сервиса диалогов. По умолчанию localhost:9091
* OTEL_EXPORTER_TYPE - Экспортер трассировок, доступны значения: otel_http,
  stdout. По умолчанию: stdout
* OTEL_EXPORTER_OTLP_ENDPOINT - адрес коллектора, работающего по протоколу OTLP over http. По умолчанию: localhost:4318

## Локальный запуск приложения

Для запуска приложения необходим установленный docker

Version:           24.0.5
API version:       1.43

- Скопируйте .env.example в .env файл.
- Запустите следующие команды по порядку.

```
docker network create myfacebook
make build
make run
```