# MyFacebook API

## Настройки окружения

* LOG_LEVEL - Уровень логирования в приложении. По умолчанию info
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

## Локальный запуск приложения

Для запуска приложения необходим установленный docker

Version:           24.0.5
API version:       1.43

- Скопируйте .env.example в .env файл.
- Запустите следующие команды по порядку.

```
make build
make run
```