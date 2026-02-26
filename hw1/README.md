## Технологии

[![Go](https://img.shields.io/badge/-Go-464646?style=flat-square&logo=Go)](https://go.dev/)
[![Python](https://img.shields.io/badge/-Python-464646?style=flat-square&logo=Python)](https://www.python.org/)
[![Linux](https://img.shields.io/badge/-Linux-464646?style=flat-square&logo=Linux)](https://www.linux.org/)
[![Bash](https://img.shields.io/badge/-Bash-464646?style=flat-square&logo=gnubash)](https://www.gnu.org/software/bash/)
[![Make](https://img.shields.io/badge/-Make-464646?style=flat-square&logo=cmake)](https://www.gnu.org/software/make/)
[![HTTP](https://img.shields.io/badge/-HTTP-464646?style=flat-square&logo=http)](https://developer.mozilla.org/en-US/docs/Web/HTTP)
[![TCP/IP](https://img.shields.io/badge/-TCP%2FIP-464646?style=flat-square&logo=cisco)](https://en.wikipedia.org/wiki/Internet_protocol_suite)

# CLI утилита многопоточного curl'a с хеджированием

---
## Описание задания:
Необходимо реализовать **CLI утилиту** `hedgedcurl` (наш ответ `curl`'у) для выполнения хеджированных HTTP запросов на языке программирования Go.

---
## Справка:
**Хеджирование запросов (request hedging)** - это техника повышения надежности и скорости получения ответа путем отправки одинакового запроса сразу к нескольким серверам и использования первого полученного ответа.

---
## Требования к утилите:
- Принимает произвольное количество URL в качестве аргументов командной строки
- Выполняет HTTP GET запросы ко всем указанным URL параллельно (асинхронно)
- Выводит только первый полученный ответ (включая заголовки и тело)
- Игнорирует все остальные ответы после получения первого
- Завершает работу сразу после получения первого ответа
- Корректно обрабатывает ошибки сети и неверные URL
- Если все запросы завершились ошибкой, выводит сообщение об ошибке

---
## Спец. коды возврата:
- **228** - ошибка таймаута

---
## Обязательные флаги командной строки:

#### `-t, --timeout SECONDS`
**Устанавливает таймаут** для всех HTTP запросов в секундах:
- По умолчанию: 15 секунд
- Пример: `hedgedcurl -t 30 url1.com url2.com`
- Пример: `hedgedcurl --timeout 5 url1.com url2.com`

#### `-h, --help`
**Выводит справку** по использованию утилиты:
- Пример: `hedgedcurl -h`
- Пример: `hedgedcurl --help`

---
## Пример использования:

```bash
# Запрос к нескольким серверам, вернет первый полученный ответ
./hedgedcurl https://example.com https://example.org

HTTP/1.1 200 OK
Expires: Thu, 26 Feb 2026 19:28:28 GMT
Cache-Control: public, max-age=14400
Cf-Cache-Status: HIT
Vary: Accept-Encoding
Content-Type: text/html
Last-Modified: Wed, 25 Feb 2026 07:20:54 GMT
Allow: GET, HEAD
Server: cloudflare
Cf-Ray: 9d4078f24e395b66-VIE
Date: Thu, 26 Feb 2026 15:28:28 GMT
Age: 7120

# С таймаутом 5 секунд
./hedgedcurl -t 5 https://httpbin.org/delay/1 https://httpbin.org/delay/10

HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Date: Thu, 26 Feb 2026 15:29:00 GMT
Content-Type: application/json
Content-Length: 322
Server: gunicorn/19.9.0
Access-Control-Allow-Origin: *

{
  "args": {},
  "data": "",
  "files": {},
  "form": {},
  "headers": {
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "Go-http-client/2.0",
    "X-Amzn-Trace-Id": "Root=1-69a066bb-5c7c77c172da778e6a611bfe"
  },
  "origin": "93.91.124.210",
  "url": "https://httpbin.org/delay/1"
}

# Помощь
./hedgedcurl --help

Usage: hedgedcurl [OPTIONS] URL [URL...]

CLI утилита многопоточного curl'а с хеджированием запросов.

Options:
  -t, --timeout SECONDS  Таймаут для HTTP запросов в секундах (по умолчанию: 15)
  -h, --help             Показать эту справку

Examples:
  hedgedcurl https://example.com https://example.org
  hedgedcurl -t 5 https://httpbin.org/delay/1 https://httpbin.org/delay/10
  hedgedcurl --help
```

---
## Ожидаемый вывод:

Утилита должна выводить:
- HTTP статус-код и заголовки ответа
- Тело ответа

---
## Технологии
* Go
* Python
* Linux
* Bash
* Make/Makefile
* HTTP
* TCP

## Запуск проекта

**1. Клонировать репозиторий:**
```
git clone https://github.com/KazikovAP/backend_course_start.git
```

**2. Перейти в директорию hw1:**
```
cd hw1/
```

**3. Выдать права .sh файлам на выполнение:**
```
chmod +x *.sh
```

**4. Запустить тесты:**
```
make test
```

---
## Ответ тестов
```
🔧 Установка зависимостей...
⚠️  pip3 не найден, пропускаем установку зависимостей
✅ Подготовка завершена
🧪 Запуск тестов...
python3 tests/tests.py
🧪 Начало тестирования домашнего задания №1 - hedgedcurl
ℹ️  Найден файл: hedgedcurl.go
ℹ️  Запуск компиляции...
✅ Компиляция завершена успешно
ℹ️  Запущен тестовый сервер на порту 37401
ℹ️  Запущен тестовый сервер на порту 42895
ℹ️  Запущен тестовый сервер на порту 40809
ℹ️  Выполнение: Тест с одним URL
ℹ️  Тестирование с одним URL...
✅ Тест с одним URL прошел успешно

ℹ️  Выполнение: Тест формата вывода
ℹ️  Тестирование формата вывода...
✅ HTTP статус-код найден в выводе
✅ HTTP заголовки найдены в выводе
✅ Тело ответа найдено в выводе
✅ Тест формата вывода прошел успешно

ℹ️  Выполнение: Тест флажка --help
ℹ️  Проверка работы флажков -h и --help...
✅ Флажок -h работает корректно
✅ Флажок --help работает корректно
✅ Проверка флажков -h и --help прошла успешно

ℹ️  Выполнение: Тест флажка --timeout
ℹ️  Тестирование флажка --timeout...
✅ Тест таймаута прошел успешно

ℹ️  Выполнение: Тест хеджирования с задержками
✅ OK: hedgedcurl выполнился за 0.01s
✅ Получен ответ от быстрого сервера (delay=0)

ℹ️  Выполнение: Тест обработки ошибок
ℹ️  Тестирование обработки ошибок...
✅ hedgedcurl корректно обработал ошибки

ℹ️  Выполнение: Тест смешанных URL
ℹ️  Тестирование со смешанными валидными/невалидными URL...
✅ hedgedcurl корректно обработал смешанные URL
✅ hedgedcurl вернул ответ от валидного сервера

============================================================
📊 ИТОГОВЫЙ ОТЧЁТ
============================================================
✅ Тест с одним URL
✅ Тест формата вывода
✅ Тест флажка --help
✅ Тест флажка --timeout
✅ Тест хеджирования с задержками
✅ Тест обработки ошибок
✅ Тест смешанных URL
------------------------------------------------------------
✅ Все тесты пройдены: 7/7
🎉 Поздравляем! hedgedcurl работает корректно!
✅ Все тесты прошли успешно!
```

---
## Разработал:
[Aleksey Kazikov](https://github.com/KazikovAP)

---
## Лицензия:
[MIT](https://opensource.org/licenses/MIT)
