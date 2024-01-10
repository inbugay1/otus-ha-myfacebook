# MyFacebook API

## Настройки окружения

* LOG_LEVEL - Уровень логирования в приложении. По умолчанию info
* SERVICE_NAME - Имя сервиса. По умолчанию myfacebook
* VERSION - Версия сервиса. По умолчанию version_not_set
* HTTP_INT_PORT - HTTP порт приложения. По умолчанию 9090
* REQUEST_HEADER_MAX_SIZE - максимальный размер header для входящих запросов. По умолчанию 10000 байт.
* REQUEST_READ_HEADER_TIMEOUT_MILLISECONDS - максимальное время отпущенное клиенту на чтение header в мс. По умолчанию
  2000мс.

* WRITE_DB_HOST - Адрес хоста для подключения к write БД. По умолчанию localhost
* WRITE_DB_PORT - Порт для подключения к write БД. По умолчанию 5432
* WRITE_DB_USERNAME - Имя пользователя write БД. По умолчанию postgres
* WRITE_DB_PASSWORD - Пароль к write БД. По умолчанию secret
* WRITE_DB_NAME - Название write БД. По умолчанию myfacebook
* WRITE_DB_DRIVER_NAME - Драйвер write БД. По умолчанию postgres
* WRITE_DB_SSL_MODE - Режим работы ssl для postgres. По умолчанию disable
* WRITE_DB_MAX_OPEN_CONNECTIONS - Число максимально одновременно открытых подключений к write БД. По умолчанию: 10

* READ_DB_HOST - Адрес хоста для подключения к read БД. По умолчанию localhost
* READ_DB_PORT - Порт для подключения к read БД. По умолчанию 5432
* READ_DB_USERNAME - Имя пользователя read БД. По умолчанию postgres
* READ_DB_PASSWORD - Пароль к read БД. По умолчанию secret
* READ_DB_NAME - Название read БД. По умолчанию myfacebook
* READ_DB_DRIVER_NAME - Драйвер read БД. По умолчанию postgres
* READ_DB_SSL_MODE - Режим работы ssl для postgres. По умолчанию disable
* READ_DB_MAX_OPEN_CONNECTIONS - Число максимально одновременно открытых подключений к read БД. По умолчанию: 10

* MYFACEBOOK_DIALOG_API_BASE_URL - Адрес сервиса диалогов. По умолчанию localhost:9091
* OTEL_EXPORTER_TYPE - Экспортер трассировок, доступны значения: otel_http,
  stdout. По умолчанию: stdout
* OTEL_EXPORTER_OTLP_ENDPOINT - адрес коллектора, работающего по протоколу OTLP over http. По умолчанию: localhost:4318

* RMQ_HOST - Хост для подключения к RabbitMQ. По умолчанию: localhost
* RMQ_HOST_PORT - Порт для подключения к RabbitMQ. По умолчанию: 5672
* RMQ_USERNAME - Имя пользователя. По умолчанию: guest
* RMQ_PASSWORD - Пароль. По умолчанию: guest

* REDIS_HOST - Хост для подключения к Redis. По умолчанию: localhost
* REDIS_PORT - Хост для подключения к Redis. По умолчанию: 6379
* REDIS_PASSWORD - Пароль. По умолчанию отсутствует
* REDIS_DB_NUM - Номер БД. По умолчанию: 0

* CONNECTION_WATCHER_PING_INTERVAL_SECONDS - Интервал в секундах, с которым пингуются сервисы, чтобы проверить состояние соединения. По умолчанию 5 сек.
* CONNECTION_WATCHER_PING_TIMEOUT_SECONDS - Таймаут пинга в секундах. По умолчанию 2 сек.
* CONNECTION_WATCHER_RECONNECT_TIMEOUT_SECONDS - Таймаут на переподключение к сервису в секундах. По умолчанию 2 сек.

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