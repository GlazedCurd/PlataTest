# Plata — Go Engineer — Test Assignment

### Требования

- Docker
- Docker Compose

### Запуск
Собираем и запускаем приложение с помощью Docker Compose:
Ключ надо взять из https://exchangeratesapi.io/documentation/
   ```
   EXCHANGERATESAPI_API_KEY={KEY} docker compose --env-file docker-compose.env -f docker-compose.yml up --build --force-recreate
   ```

### Использование
Доступ к приложению по адресу `http://localhost:8080`.

Пример запроса на обновление котировки:
```
$ curl localhost:8080/quotes/EUR_USD/task -d '{ "idempotency_key":"abcdefghij1324"}'
```
Ответ 
```
{
  "id": 22,
  "price": 1.170617,
  "code": "EUR_USD",
  "idempotency_key": "abcdefghij1324",
  "created_at": "2025-08-17T19:31:08.601595Z",
  "updated_at": "2025-08-17T19:31:18.201311Z",
  "status": "success"
}
```

Код пары записывается через нижнее подчёркивание, например `EUR_USD`. Ключ идемпотентности передаётся в теле запроса.
Если сходить два раза с одним и тем же ключём (для одной и той же пары) - нового апдейта добавлено не будет. Если же ключ будет другим, то будет возвращён конфликт. И сообщение вида:
```
{
  "error": "Conflict with different body"
}
```

Пример запроса последней котировки
```
curl localhost:8080/quotes/EUR_USD
```

Пример запроса конкретной котировки 
```
curl localhost:8080/quotes/EUR_USD/task/22
```

### Endpoints


### Примечание
`USD_MXN` не обрабатывается exchangeratesapi.io
с ошибкой 
```
{"error":{"code":"base_currency_access_restricted","message":"An unexpected error ocurred. [Technical Support: support@apilayer.com]"}}
```

### Что можно сделать лучше
- Работа с конфигами. Сейчас сделано через переменные окружения, но можно использовать более удобные решения, например, Viper.
- Сделать ретраи на уровне очереди (сейчас есть ретраи на уровне клиента, но возможно следует добавить возвращение заявок в очередь на обработку).
- Больше тестов 
- Генерация openapi
- Поднимать compose и готовить базу в автоматическом режиме
- Настроить пользователей в базе
- Больше юнит тестов