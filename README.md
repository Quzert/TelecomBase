# TelecomBase

Десктопное приложение (Qt/C++) + Go API + PostgreSQL для ведения сведений о телекоммуникационном оборудовании.

## Быстрый старт (backend + DB)

1) Скопируйте переменные окружения:

`cp .env.example .env`

2) Запустите сервисы:

`docker compose up --build`

3) Проверка доступности API:

`curl -s http://localhost:8080/health`

Ожидаемый ответ: `{"status":"ok"}`

PostgreSQL доступен на хосте по порту `5433`.

## Сборка Qt клиента (Linux)

Зависимости: Qt6 (Widgets, Network), CMake, компилятор C++20.

Команды:

`cmake -S . -B build`

`cmake --build build -j`

Запуск клиента:

`./build/client/TelecomBaseClient`

## Структура репозитория

- `server/` — Go API.
- `db/` — SQL инициализация (миграции/seed).
- `client/` — Qt desktop (будет добавлено следующим этапом).
- `docs/` — документация (включая план разработки).
