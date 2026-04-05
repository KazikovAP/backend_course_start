# Сервер для хранения цен криптовалют

---
## Описание задания:
**HTTP REST API** сервер на Go для хранения и отслеживания цен криптовалют в реальном времени с JWT-аутентификацией и интеграцией CoinGecko API.

---
## Технологии
* **Go** — основной язык реализации
* **Python + Make/Makefile** скрипты — тестирования и сборки
* **JWT** — аутентификация пользователей
* **bcrypt** — хеширование паролей
* **CoinGecko API** — получение актуальных цен криптовалют
* **In-memory storage** — хранение данных в оперативной памяти

---
## Архитектура

Проект построен по принципам чистой архитектуры с чётким разделением слоёв:
```make
hw2/
├── cmd/
│   └── server/
│       └── main.go                 # Точка входа, DI-сборка
├── internal/
│   ├── domain/
│   │   └── models.go               # Сущности и интерфейсы репозиториев
│   ├── repository/
│   │   └── memory.go               # In-memory реализация репозиториев
│   ├── coingecko/
│   │   └── client.go               # Клиент CoinGecko API с офлайн-фолбэком
│   ├── service/
│   │   ├── auth.go                 # Бизнес-логика аутентификации
│   │   ├── crypto.go               # Бизнес-логика криптовалют
│   │   └── scheduler.go            # Фоновый планировщик обновлений
│   └── handler/
│       ├── auth.go                 # HTTP-обработчики /auth/*
│       ├── crypto.go               # HTTP-обработчики /crypto/*
│       ├── schedule.go             # HTTP-обработчики /schedule/*
│       ├── middleware.go           # JWT middleware
│       └── helpers.go              # writeJSON, errorResponse
├── compile.sh
├── execute.sh
├── Makefile
└── go.mod
```

---
## API

### Аутентификация
- **POST /auth/register** - Регистрация нового пользователя
```json
// Request
{"username": "alice", "password": "secret123"}

// Response 201
{"token": "<jwt>"}

// Error 400 / 409
{"error": "invalid input" | "user exists"}
```

- **POST /auth/login** - Вход пользователя
```json
// Request
{"username": "alice", "password": "secret123"}

// Response 200
{"token": "<jwt>"}

// Error 400 / 401
{"error": "invalid input" | "invalid credentials"}
```

### Криптовалюты
Все операции требуют аутентификации через заголовок `Authorization: Bearer <token>`
- **GET /crypto** - Список всех отслеживаемых криптовалют
```json
{
  "cryptos": [
    {"symbol": "BTC", "name": "Bitcoin", "current_price": 85000.50, "last_updated": "2026-04-04T18:31:24Z"}
  ]
}
```

- **POST /crypto** - Добавить криптовалюту для отслеживания
```json
// Request
{"symbol": "BTC"}

// Response 201
{"crypto": {"symbol": "BTC", "name": "Bitcoin", "current_price": 85000.50, "last_updated": "..."}}

// Error 400 / 409 / 500
{"error": "symbol is required" | "already exists" | "..."}
```

- **GET /crypto/{symbol}** - Получить информацию о конкретной криптовалюте
```json
// Response 200
{"symbol": "BTC", "name": "Bitcoin", "current_price": 85000.50, "last_updated": "2026-04-04T18:31:47Z"}

// Error 404
{"error": "not found"}
```

- **PUT /crypto/{symbol}/refresh** - Принудительно обновить цену криптовалюты
```json
// Response 200
{"crypto": {"symbol": "BTC", "name": "Bitcoin", "current_price": 85200.00, "last_updated": "2026-04-04T18:31:59Z"}}

// Error 404 / 500
{"error": "not found" | "failed to fetch price"}
```

- **GET /crypto/{symbol}/history** - Получить историю цен криптовалюты
```json
{
  "symbol": "BTC",
  "history": [
    {"price": 85000.50, "timestamp": "2026-04-04T18:32:19Z"},
    {"price": 85100.00, "timestamp": "2026-04-04T18:32:19Z"}
  ]
}
```

- **GET /crypto/{symbol}/stats** - Получить статистику по ценам криптовалюты
```json
{
  "symbol": "BTC",
  "current_price": 85100.00,
  "stats": {
    "min_price": 84000.0,
    "max_price": 86000.0,
    "avg_price": 85000.0,
    "price_change": 1000.0,
    "price_change_percent": 1.19,
    "records_count": 48
  }
}
```

- **DELETE /crypto/{symbol}** - Удалить криптовалюту из отслеживания (включая историю)
```json
// Response 200
{}

// Error 404
{"error": "not found"}
```

---
## Расписание автообновления (дополнительная часть):

- **GET /schedule** - Получить текущие настройки автообновления
```json
{
  "enabled": true,
  "interval_seconds": 30,
  "last_update": "2026-04-04T18:32:39Z",
  "next_update": "2026-04-04T18:32:39ZZ"
}
```

- **PUT /schedule** - Изменить настройки автообновления
```json
// Request
{"enabled": true, "interval_seconds": 60}

// Response 200
{"enabled": true, "interval_seconds": 60}

// Error 400
{"error": "interval_seconds must be between 10 and 3600"}
```

- **POST /schedule/trigger** - Принудительно запустить обновление всех цен
```json
{
  "updated_count": 3,
  "timestamp": "2024-01-01T12:00:00Z"
}
```

---
## Запуск проекта

**1. Клонировать репозиторий:**
```
git clone https://github.com/KazikovAP/backend_course_start.git
```

**2. Перейти в директорию hw2:**
```
cd hw2/
```

**3. Выдать права на выполнение .sh файлам:**
```
chmod +x *.sh
```

**4. Собрать и запустить сервер:**
```
./compile.sh && ./execute.sh
```

либо

```
go run cmd/server/main.go
```

---
## Тесты

### Основные тесты
```bash
make test
```

---
## Ответ тестов
```
==================================================
📊 ИТОГОВЫЙ ОТЧЁТ
==================================================
✅ Регистрация пользователя
✅ Вход пользователя
✅ Добавление криптовалюты
✅ Получение списка криптовалют
✅ Получение конкретной криптовалюты
✅ Обновление цены криптовалюты
✅ История цен криптовалюты
✅ Статистика криптовалюты
✅ Удаление криптовалюты
✅ Требование аутентификации
--------------------------------------------------
📋 Режим: Основные тесты (используйте SCHEDULE=1 для полного тестирования)
✅ Все тесты пройдены: 10/10
🎉 Поздравляем! Домашнее задание выполнено корректно!
✅ Все тесты прошли успешно!
```

### Дополнительные тесты
```bash
make test SCHEDULE=1
```

---
## Ответ тестов
```
==================================================
📊 ИТОГОВЫЙ ОТЧЁТ
==================================================
✅ Регистрация пользователя
✅ Вход пользователя
✅ Добавление криптовалюты
✅ Получение списка криптовалют
✅ Получение конкретной криптовалюты
✅ Обновление цены криптовалюты
✅ История цен криптовалюты
✅ Статистика криптовалюты
✅ Получение настроек расписания
✅ Изменение настроек расписания
✅ Принудительное обновление цен
✅ Удаление криптовалюты
✅ Требование аутентификации
--------------------------------------------------
📋 Режим: Полные тесты (основные + дополнительные)
✅ Все тесты пройдены: 13/13
🎉 Поздравляем! Домашнее задание выполнено корректно с дополнительными функциями!
✅ Все тесты прошли успешно!
```

---
## Разработал:
[Aleksey Kazikov](https://github.com/KazikovAP)

---
## Лицензия:
[MIT](https://opensource.org/licenses/MIT)
