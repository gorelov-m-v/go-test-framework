## Проблема, которую мы решаем

**Go становится стандартом для разработки бэкенда.**
Команды массово переходят на микросервисы, но автоматизация тестирования часто отстает. В этом процессе возникают две фундаментальные проблемы:

1.  **Технологический разрыв.** QA-инженеры либо вынуждены использовать "зоопарк" технологий (Python/Java для тестов Go-сервисов), разрывая связь с разработкой, либо сталкиваются с высоким порогом входа в Go.
2.  **Кошмар поддержки.** E2E тесты имеют тенденцию превращаться в нечитаемую "стену кода". Суть бизнес-сценария тонет в технической реализации: настройке клиентов, парсинге JSON и бесконечных проверках ошибок. Такие тесты сложно писать, больно читать и очень дорого поддерживать.

Этот фреймворк создан, чтобы устранить эти барьеры. Он позволяет команде работать в **едином стеке**, превращая императивный технический код в чистый декларативный сценарий, доступный для понимания каждому участнику процесса.

## Идеология проекта

### 1. Низкий порог входа и "Быстрая автоматизация" (DSL)
Ручное регрессионное тестирование в современном темпе разработки становится неоправданно дорогим и медленным бутылочным горлышком.
Фреймворк предоставляет **Fluent Interface**, который позволяет писать тесты со скоростью написания ручных тест-кейсов. Вместо борьбы с синтаксисом кода (`if err != nil`), QA описывает сценарий: `Request... Expect... Fetch`.

**Результат:** Мы переходим от "догоняющей автоматизации" к выпуску тестов **одновременно с продуктом**. Это исключает рутину ручных проверок и сокращает время регресса с дней до минут.

### 2. Нативная поддержка асинхронности (Борьба с Flaky-тестами)
Современные системы асинхронны. Данные появляются в БД или Брокере сообщений не сразу после ответа API.
Механизм "умных ретраев" (Polling с Backoff & Jitter) встроен глубоко в ядро DSL.

**Результат:** Вы пишете линейный тест, а фреймворк сам берет на себя ожидание данных. Никаких `time.Sleep()` и нестабильных (мигающих) тестов.

### 3. Dependency Injection в стиле Spring
В классической разработке добавление нового микросервиса в тесты требует написания десятков строк настроечного кода, который часто дублируется и ломается.
Здесь реализован механизм **"Умной конфигурации"** (Tag-based DI): вы добавляете одну строку в настройки, и фреймворк сам подключает нужные клиенты и базы данных.

**Результат:** Подключение нового сервиса занимает **30 секунд**. Архитектура проекта остается чистой, даже когда количество тестов вырастает до тысяч.

### 4. Автоматическая отчетность (Allure)
Качественные отчеты критически важны, но их ручная поддержка - это рутина.
Здесь интеграция с **Allure Report** "зашита" в ядро DSL. Любое действие (HTTP запрос, SQL выборка, проверка поля в сообщении из Kafka) автоматически превращается в детализированный шаг отчета.

**Результат:** QA фокусируется только на сценарии теста. Вся "доказательная база" (Request/Response, Headers, SQL) собирается "под капотом".

### 5. Высокопроизводительная многопоточность
E2E тесты часто бывают медленными из-за IO-операций. Фреймворк использует нативные возможности Go на полную мощность.
Поддерживается не только параллельный запуск разных тестов, но и **параллельное выполнение асинхронных шагов внутри одного теста**.

**Результат:** Вы можете одновременно ждать событие в Kafka, проверять запись в БД и дергать API, кардинально сокращая время прогона пайплайна.

## Навигация

- [Быстрый старт: ваш первый тест за 5 минут](#быстрый-старт-ваш-первый-тест-за-5-минут)
- [Рекомендуемая структура проекта](#рекомендуемая-структура-проекта)
- [HTTP DSL](#http-dsl)
    - [Сквозной E2E пример: Шаг 1 - HTTP запрос](#сквозной-e2e-пример-шаг-1---http-запрос)
    - [Справочник возможностей DSL](#справочник-возможностей-dsl)
- [Database DSL](#database-dsl-sql)
    - [Сквозной E2E пример: Шаг 2 - Проверка в БД](#сквозной-e2e-пример-шаг-2---проверка-в-бд)
    - [Справочник методов DB DSL](#справочник-методов-db-dsl)
- [Kafka DSL](#kafka-dsl)
    - [Сквозной E2E пример: Шаг 3 - Проверка события в Kafka](#сквозной-e2e-пример-шаг-3---проверка-события-в-kafka)
    - [Справочник методов Kafka DSL](#справочник-методов-kafka-dsl)
- [Асинхронные шаги (AsyncStep)](#асинхронные-шаги-asyncstep)
    - [Проблема: Flaky-тесты в асинхронных системах](#проблема-flaky-тесты-в-асинхронных-системах)
    - [Решение: AsyncStep с автоматическими retry](#решение-asyncstep-с-автоматическими-retry)
    - [Параллельное выполнение](#параллельное-выполнение)
    - [Разница между Step и AsyncStep](#разница-между-step-и-asyncstep)
    - [Конфигурация](#конфигурация-1)
    - [Примеры использования](#примеры-использования)
    - [Рекомендации](#рекомендации)
- [Параметризованные тесты (Table-Driven Tests)](#параметризованные-тесты-table-driven-tests)
    - [Зачем нужны параметризованные тесты](#зачем-нужны-параметризованные-тесты)
    - [Сквозной пример: Негативное тестирование регистрации](#сквозной-пример-негативное-тестирование-регистрации)
    - [Использование RequestBodyMap для негативных тестов](#использование-requestbodymap-для-негативных-тестов)
    - [Полный пример с RequestBodyMap](#полный-пример-с-requestbodymap)
    - [Рекомендации](#рекомендации-1)
- [Маскировка чувствительных данных](#маскировка-чувствительных-данных)
    - [Зачем нужна маскировка](#зачем-нужна-маскировка)
    - [Принцип работы](#принцип-работы)
    - [Конфигурация маскировки](#конфигурация-маскировки)
    - [Примеры маскировки](#примеры-маскировки)
    - [Важные замечания](#важные-замечания)
    - [Рекомендации](#рекомендации-2)
- [Генерация тестовых данных](#генерация-тестовых-данных)
    - [Проблема: Хардкод данных в тестах](#проблема-хардкод-данных-в-тестах)
    - [Решение: Генератор данных](#решение-генератор-данных)
    - [Справочник функций](#справочник-функций)
    - [Примеры использования](#примеры-использования-1)

---

### Быстрый старт: ваш первый тест за 5 минут

Лучший способ освоить инструмент - это увидеть его в действии. Мы подготовили шаблонный проект, который содержит рекомендуемую структуру, примеры конфигурации и готовые демонстрационные тесты.
Этот подход позволяет пропустить этап начальной настройки и сразу перейти к написанию сценариев.

#### Шаг 1: Предварительные требования

Убедитесь, что в вашей системе установлены:

*   **Go** (версия 1.25 или выше)
*   **Git**
*   **Allure Commandline** (для генерации отчетов). [Инструкция по установке](https://docs.qameta.io/allure/#_installing_a_commandline).

#### Шаг 2: Клонируйте шаблонный проект

Вам не нужно клонировать репозиторий самого фреймворка. Достаточно клонировать проект-шаблон, где фреймворк уже подключен как зависимость.

```bash
# ЗАГЛУШКА: в разработке
git clone https://github.com/your-org/go-e2e-framework-example.git my-api-tests

cd my-api-tests
```

#### Шаг 3: Адаптируйте конфигурацию

В проекте уже есть готовый файл `configs/config.local.yaml`. Он содержит все необходимые секции с примерами значений.

Откройте этот файл в вашем редакторе и замените placeholder-значения на актуальные данные вашего тестового стенда (например, `baseURL` для вашего сервиса и `dsn` для подключения к базе данных).

#### Шаг 4: Запустите тесты

Теперь окружение полностью готово к запуску. Выполните в терминале стандартную команду для запуска тестов:

```bash
go test -v ./...
```

Успешное выполнение завершится выводом `--- PASS` в консоли. Это подтверждает, что фреймворк работает корректно, а в корне проекта создана директория `allure-results` с результатами прогона.

#### Шаг 5: Изучите детализированный отчет

На этом шаге вы увидите, как фреймворк автоматически собирает всю необходимую информацию для анализа прогона. Для генерации и просмотра отчета Allure выполните:

```bash
allure serve allure-results
```

В вашем браузере откроется страница с интерактивным отчетом. Изучите шаги теста, раскройте вложения с HTTP-запросами и убедитесь, что вся "доказательная база" собрана без дополнительных усилий с вашей стороны.

**Результат:** Вы успешно запустили полноценный E2E-тест.

Теперь вы готовы к решению реальных задач:
1.  **Проанализируйте** код в файлах `*_test.go`, чтобы понять логику построения тестов.
2.  **Модифицируйте** существующие тесты, используя эндпоинты и модели вашего API.
3.  **Создайте** новый тест, опираясь на готовые примеры как на образец.

Дальнейшая документация служит справочником по всем возможностям DSL, которые вы можете использовать для разработки надежных и читаемых E2E-сценариев.

---

# Рекомендуемая структура проекта

Фреймворк не навязывает жесткую структуру, но следование этой организации файлов упрощает поддержку и масштабирование тестов:

```
your-api-tests/
├── configs/                      # Файлы конфигурации
│   └── config.local.yaml         # Конфигурация локального окружения
│
├── internal/                     # Внутренние компоненты
│   ├── client/                   # HTTP клиенты (DSL методы)
│   │   └── [service_name]/       # Отдельная папка для каждого сервиса
│   │       └── client.go         # DSL методы для API сервиса
│   │
│   ├── db/                       # Database репозитории
│   │   └── [table_name]/         # Отдельная папка для каждой таблицы/домена
│   │       └── repo.go           # DSL методы для запросов к БД
│   │
│   ├── models/                   # Модели данных
│   │   ├── http/                 # HTTP модели (организованы по сервисам)
│   │   │   └── [service_name]/
│   │   │       └── *.go          # Request/Response модели
│   │   └── db/                   # Database модели (организованы по БД)
│   │       └── [database_name]/
│   │           └── *.go          # Модели таблиц с тегами `db`
│   │
│   └── kafka/                    # Kafka топики и сообщения
│       └── topics.go             # Определения топиков и моделей сообщений
│
├── tests/                        # Тестовые сьюты
│   ├── env.go                    # Тестовое окружение (DI контейнер)
│   └── *_test.go                 # Файлы с тестами
│
├── allure-results/               # Результаты Allure (генерируются автоматически)
├── .gitignore
├── go.mod
└── go.sum
```

## Ключевые принципы организации

### 1. Разделение по слоям
- **`internal/client/`** — HTTP DSL методы, группируются по микросервисам
- **`internal/db/`** — Database DSL методы, группируются по таблицам/доменам
- **`internal/models/`** — Модели данных, разделены на `http/` и `db/`
- **`internal/kafka/`** — Kafka топики и модели сообщений

### 2. Группировка по доменам
Каждый микросервис/таблица/топик имеет свою папку:
```
internal/client/
├── game/client.go      # DSL для Game API
├── auth/client.go      # DSL для Auth API
└── payment/client.go   # DSL для Payment API
```

### 3. Модели рядом с их использованием
```
internal/models/
├── http/
│   ├── game/
│   │   ├── player.go         # CreatePlayerReq, CreatePlayerResp
│   │   └── game_session.go   # StartSessionReq, StartSessionResp
│   └── auth/
│       └── token.go          # TokenRequest, TokenResponse
└── db/
    └── core/
        ├── player.go         # PlayerDB (с тегами `db`)
        └── game_category.go  # GameCategoryDB
```

### 4. Единая точка входа
`tests/env.go` содержит все зависимости:
```go
type TestEnv struct {
    // HTTP клиенты
    GameService    game.Link    `config:"gameService"`
    AuthService    auth.Link    `config:"authService"`

    // Database репозитории
    PlayersRepo    players.Link `db_config:"coreDatabase"`

    // Kafka
    Kafka          kafka.Link   `kafka_config:"kafka"`
}
```

## Преимущества такой структуры

- **Предсказуемость:** Любой разработчик знает, где искать DSL метод или модель
- **Масштабируемость:** Легко добавлять новые сервисы/таблицы без реструктуризации
- **Изоляция:** Изменения в одном сервисе не влияют на другие
- **Повторное использование:** DSL методы можно использовать в разных тестах

---

# HTTP DSL

Модуль предназначен для написания функциональных тестов REST API.
Он построен на **Generics**, что гарантирует строгую типизацию запросов и ответов на этапе компиляции. Вы не сможете отправить неверную структуру или ошибиться в типе ожидаемого ответа.

## Сквозной E2E пример: Шаг 1 - HTTP запрос

Начнём полноценный E2E сценарий, который пройдёт через все три DSL.
**Сценарий:** Создаём игрока через API, проверяем запись в БД, ждём событие в Kafka.

### 0. Конфигурация (`config.local.yaml`)

```yaml
http:
  gameService:
    baseURL: "https://game-api.example.com"
    timeout: 30s
```

### 1. Спецификация (Контракт)

```bash
curl -X POST https://game-api.example.com/api/v1/players \
  -H "Content-Type: application/json" \
  -d '{
    "username": "pro_gamer_2024",
    "region": "EU"
  }'

# Response: 201 Created
{
  "id": "uuid-12345",
  "username": "pro_gamer_2024",
  "status": "active"
}
```

### 2. Описание Моделей
Переносим JSON-структуру в Go (`internal/models/player.go`). Используем стандартные `json` теги.

```go
package models

type CreatePlayerReq struct {
    Username string `json:"username"`
    Region   string `json:"region"`
}

type CreatePlayerResp struct {
    ID        string `json:"id"`
    Username  string `json:"username"`
    Status    string `json:"status"`
}
```

### 3. Реализация Клиента
**Файл:** `internal/client/game/client.go`

```go
package game

import (
    "go-test-framework/pkg/http/client"
    "go-test-framework/pkg/http/dsl"
    "github.com/ozontech/allure-go/pkg/framework/provider"
    "my-project/internal/models"
)

// 1. Приватная переменная пакета (хранит http.Client с базовым URL и хедерами)
var httpClient *client.Client

// 2. Структура Link для Auto-Wiring.
// Билдер найдет её в TestEnv и автоматически внедрит клиент.
type Link struct{}

func (l *Link) SetHTTP(c *client.Client) {
    httpClient = c
}

// 3. DSL Метод
// Возвращает типизированный Call: [RequestModel, ResponseModel]
func CreatePlayer(sCtx provider.StepCtx) *dsl.Call[models.CreatePlayerReq, models.CreatePlayerResp] {
    // dsl.NewCall связывает шаг теста (sCtx) и транспорт (httpClient).
    // Далее мы сразу указываем HTTP метод и путь.
    return dsl.NewCall[models.CreatePlayerReq, models.CreatePlayerResp](sCtx, httpClient).
        POST("/api/v1/players")
}
```

### 4. Подключение в Env
Добавляем связь в `tests/env.go`. Билдер увидит `game.Link` и прокинет зависимости.

**Файл:** `tests/env.go`

```go
package tests

import (
    "go-test-framework/pkg/builder"
    "log"
    "my-project/internal/client/game"
)

type TestEnv struct {
    // HTTP клиенты - связываем конфиг "gameService" с пакетом "game"
    GameService game.Link `config:"gameService"`
}

var env *TestEnv

func init() {
    env = &TestEnv{}

    if err := builder.BuildEnv(env); err != nil {
        log.Fatalf("Failed to build test environment: %v", err)
    }
}

func Env() *TestEnv {
    return env
}
```

### 5. Тест: Создаём игрока и сохраняем ID
Это начало нашего сквозного сценария. Мы отправляем HTTP-запрос и **сохраняем `playerID`** для использования в следующих шагах (Database и Kafka).

```go
func (s *PlayerSuite) TestCreatePlayerE2E(t provider.T) {
    t.Title("E2E: Создание игрока (HTTP → DB → Kafka)")

    var playerID string
    var username = "pro_gamer_2024"

    // ШАГ 1: HTTP - Создаём игрока через API
    s.Step(t, "HTTP: Создание игрока через API", func(sCtx provider.StepCtx) {
        resp := game.CreatePlayer(sCtx).
            // 1. Настройка запроса (строгая типизация)
            RequestBody(models.CreatePlayerReq{
                Username: username,
                Region:   "EU",
            }).
            // 2. Проверки ответа
            ExpectResponseStatus(201).
            ExpectResponseBodyFieldNotEmpty("id").
            ExpectResponseBodyFieldValue("username", username).
            ExpectResponseBodyFieldValue("status", "active").
            // 3. Выполнение
            Send()

        // Сохраняем ID для проверки в БД и Kafka
        playerID = resp.Body.ID
    })

    // ШАГ 2: Database - см. раздел "Database DSL" ниже
    // ШАГ 3: Kafka - см. раздел "Kafka DSL" ниже
}
```

---

## Справочник возможностей DSL

Объект `dsl.Call[TReq, TResp]` предоставляет богатый набор методов для настройки запроса и валидации ответа.

### 1. Настройка запроса (Request Configuration)

Эти методы определяют, *что* мы отправляем.

| Метод | Описание | Пример |
| :--- | :--- | :--- |
| `.Header(k, v)` | Добавление заголовка. | `.Header("Authorization", "Bearer ...")` |
| `.QueryParam(k, v)` | Добавление GET-параметра. | `.QueryParam("page", "1")` -> `?page=1` |
| `.PathParam(k, v)` | Подстановка переменной в путь. | `.PathParam("id", "123")` -> `/users/123` |
| `.RequestBody(val)` | Установка тела (структура). | `.RequestBody(models.User{...})` |
| `.RequestBodyMap(map)` | Установка тела (map). Для негативных тестов. | `.RequestBodyMap(map[string]interface{}{"password": "123"})` |

### 2. Ожидания (Expectations)

Проверки добавляются в цепочку **до** отправки запроса. Они работают по принципу "Silent Success, Loud Failure": если всё хорошо, тест идет дальше. Если проверка не прошла, тест падает с детальным описанием в Allure.

#### Статус и Тело
*   `.ExpectResponseStatus(code int)` — Проверяет HTTP Status Code.
*   `.ExpectResponseBodyNotEmpty()` — Проверяет, что тело ответа пришло и не пустое.

#### Проверка полей (JSON Path)
Для навигации по JSON-структуре используется синтаксис [GJSON](https://github.com/tidwall/gjson).

Предположим, API вернул такой ответ:
```json
{
  "token": "eyJhbGciOiJIUz...",
  "meta": { "server": "auth-01" },
  "items": [
    { "id": 101, "code": "read" },
    { "id": 102, "code": "write" }
  ]
}
```

**Примеры проверок:**

| Путь (Path) | Значение | DSL Метод |
| :--- | :--- | :--- |
| `"token"` | `"eyJ..."` | `.ExpectResponseBodyFieldNotEmpty("token")` |
| `"meta.server"` | `"auth-01"` | `.ExpectResponseBodyFieldValue("meta.server", "auth-01")` |
| `"items.0.code"` | `"read"` | `.ExpectResponseBodyFieldValue("items.0.code", "read")` |
| `"items.#"` | `2` | `.ExpectResponseBodyFieldValue("items.#", 2)` |

**Поддерживаемый синтаксис путей:**
- Простые поля: `"name"`
- Вложенные поля: `"user.email"`
- Элементы массива: `"items.0"`, `"items.1"`
- Подсчёт элементов: `"items.#"`
- Вложенные поля в массиве: `"users.0.name"`

**Поддерживаемые типы сравнения:**
*   `string`: `"active"`
*   `int`, `float`: `100`, `99.99`
*   `bool`: `true`, `false`
*   `nil`: Проверяет, что поле в JSON равно `null` или отсутствует.

### 3. Выполнение и Результат

*   `.Send()` — **Финализирующий метод.**
    1.  Собирает HTTP запрос.
    2.  Создает шаг в Allure.
    3.  Прикрепляет `curl`, Headers, Body запроса и ответа.
    4.  Выполняет все `Expect` проверки.
    5.  Возвращает типизированный ответ `*client.Response[TResp]`.

    ```go
    // Пример chain-requests: Создали -> Забрали ID
    var playerID string

    s.Step(t, "Create", func(sCtx provider.StepCtx) {
        resp := game.CreatePlayer(sCtx).
            RequestBody(...).
            ExpectResponseStatus(201).
            Send() // <-- Возвращает *Response[CreatePlayerResp]

        playerID = resp.Body.ID // Строго типизированный доступ
    })
    ```

---

# Database DSL (SQL)

Модуль `pkg/database/dsl` предназначен для **верификации состояния** базы данных после выполнения бизнес-операций.
Он поддерживает PostgreSQL и MySQL.

---

## Сквозной E2E пример: Шаг 2 - Проверка в БД

После создания игрока через API проверим, что запись попала в базу данных.

### 0. Конфигурация (`config.local.yaml`)

Расширяем конфигурацию — добавляем секцию `database`:

```yaml
http:
  gameService:
    baseURL: "https://game-api.example.com"
    timeout: 30s

database:
  coreDatabase:
    driver: "postgres"  # или "mysql"
    dsn: "postgres://user:password@localhost:5432/game_db?sslmode=disable"
    maxOpenConns: 10
    maxIdleConns: 5
```

**Ожидаемое состояние таблицы `players`:**

| id | username | status | region | is_vip | created_at |
|:---|:---------|:-------|:-------|:-------|:-----------|
| uuid-12345 | pro_gamer_2024 | active | EU | false | 2024-01-09 10:30:00 |

### 1. Описание Модели БД
Опишите Go-структуру, соответствующую таблице. Используйте тег `db` для маппинга колонок.

**Файл:** `internal/models/db_player.go`

```go
package models

import (
    "database/sql"
    "time"
)

type PlayerDB struct {
    ID        string         `db:"id"`
    Username  string         `db:"username"`
    Status    string         `db:"status"`
    Region    sql.NullString `db:"region"`
    IsVip     bool           `db:"is_vip"`
    CreatedAt time.Time      `db:"created_at"`
}
```

### 2. Реализация Репозитория (Auto-Wiring)
**Файл:** `internal/db/players/repo.go`

```go
package players

import (
    "go-test-framework/pkg/database/client"
    "go-test-framework/pkg/database/dsl"
    "github.com/ozontech/allure-go/pkg/framework/provider"
    "my-project/internal/models"
)

var dbClient *client.Client

// Link для Auto-Wiring
type Link struct{}

func (l *Link) SetDB(c *client.Client) {
    dbClient = c
}

// DSL Метод: Поиск по ID
func FindByID(sCtx provider.StepCtx, id string) *dsl.Query[models.PlayerDB] {
    return dsl.NewQuery[models.PlayerDB](sCtx, dbClient).
        SQL("SELECT * FROM players WHERE id = ?", id)
}
```

### 3. Подключение в Env
Расширяем `tests/env.go` — добавляем Database к уже существующему HTTP клиенту.

**Файл:** `tests/env.go`

```go
package tests

import (
    "go-test-framework/pkg/builder"
    "log"
    "my-project/internal/client/game"
    "my-project/internal/db/players"
)

type TestEnv struct {
    // HTTP клиенты
    GameService game.Link `config:"gameService"`

    // Database - связываем конфиг "coreDatabase" с репозиторием "players"
    PlayersRepo players.Link `db_config:"coreDatabase"`
}

var env *TestEnv

func init() {
    env = &TestEnv{}

    if err := builder.BuildEnv(env); err != nil {
        log.Fatalf("Failed to build test environment: %v", err)
    }
}

func Env() *TestEnv {
    return env
}
```

### 4. Продолжение теста: Проверяем БД
Добавляем второй шаг к нашему E2E-тесту. Используем `playerID`, который получили из HTTP-ответа.

```go
func (s *PlayerSuite) TestCreatePlayerE2E(t provider.T) {
    t.Title("E2E: Создание игрока (HTTP → DB → Kafka)")

    var playerID string
    var username = "pro_gamer_2024"

    // ШАГ 1: HTTP - Создаём игрока через API
    s.Step(t, "HTTP: Создание игрока через API", func(sCtx provider.StepCtx) {
        resp := game.CreatePlayer(sCtx).
            // 1. Настройка запроса (строгая типизация)
            RequestBody(models.CreatePlayerReq{
                Username: username,
                Region:   "EU",
            }).
            // 2. Проверки ответа
            ExpectResponseStatus(201).
            ExpectResponseBodyFieldNotEmpty("id").
            ExpectResponseBodyFieldValue("username", username).
            ExpectResponseBodyFieldValue("status", "active").
            // 3. Выполнение
            Send()

        // Сохраняем ID для проверки в БД и Kafka
        playerID = resp.Body.ID
    })

    // ШАГ 2: Database - Проверяем запись по полученному ID
    s.Step(t, "Database: Проверка записи в БД", func(sCtx provider.StepCtx) {
        players.FindByID(sCtx, playerID). // Используем ID из HTTP-шага

            // Ожидаем, что запись существует
            ExpectFound().

            // Проверяем значения колонок
            ExpectColumnEquals("username", username).
            ExpectColumnEquals("status", "active").
            ExpectColumnEquals("region", "EU").
            ExpectColumnFalse("is_vip").         // Проверка bool
            ExpectColumnIsNotNull("created_at"). // Дата должна быть заполнена

            // Выполнение (SELECT и скан в структуру)
            Send()
    })

    // ШАГ 3: Kafka - см. раздел "Kafka DSL" ниже
}
```

---

## Справочник методов DB DSL

Объект `dsl.Query[Model]` предоставляет методы для валидации данных в БД.

### 1. Настройка запроса
*   `.SQL(query, args...)` — Устанавливает SQL-запрос и аргументы.
    *   Для MySQL используйте синтаксис `?`: `WHERE id = ?`
    *   Для PostgreSQL (pgx/sqlx) используйте `$`: `WHERE id = $1`

### 2. Ожидания (Expectations)
Проверки выполняются **после** получения данных от БД.

**Проверки наличия:**
*   `.ExpectFound()` — Ожидает, что запрос вернет хотя бы 1 строку.
*   `.ExpectNotFound()` — Ожидает, что запрос вернет `sql.ErrNoRows`.

**Проверки значений колонок:**
*   `.ExpectColumnEquals("col", val)` — Сравнивает значение.
*   `.ExpectColumnTrue("col")` / `.ExpectColumnFalse("col")` — Для boolean полей.
*   `.ExpectColumnIsNull("col")` / `.ExpectColumnIsNotNull("col")` — Для `sql.Null*` типов.

**Примечание:** Имена колонок (`"col"`) должны совпадать с тегом `db` в вашей модели.

### 3. Выполнение

*   `.Send()` — **Финализирующий метод.**
    1.  Выполняет SQL запрос (`GetContext`).
    2.  Сканирует первую строку результата в структуру `Model`.
    3.  Запускает все проверки.
    4.  Создает шаг в Allure с Query и Result.
    5.  Возвращает заполненную структуру `Model`.

---

# Kafka DSL

Модуль `pkg/kafka/dsl` предназначен для **верификации событий** в Apache Kafka.
Он построен на фоновом consumer'е, который непрерывно читает сообщения в буфер, позволяя тестам искать события с retry-логикой.

---

## Сквозной E2E пример: Шаг 3 - Проверка события в Kafka

Финальный шаг: проверяем, что система отправила событие `PLAYER_CREATED` в Kafka.

### 1. Конфигурация (`config.local.yaml`)

Добавляем последнюю секцию — `kafka`:

```yaml
http:
  gameService:
    baseURL: "https://game-api.example.com"
    timeout: 30s

database:
  coreDatabase:
    driver: "postgres"
    dsn: "postgres://user:password@localhost:5432/game_db?sslmode=disable"
    maxOpenConns: 10
    maxIdleConns: 5

kafka:
  bootstrapServers:
    - "kafka.example.com:9092"
  groupId: "qa-test-group"
  topics:
    - "game-player-events"
  bufferSize: 1000
  uniqueDuplicateWindowMs: 5000
```

### 2. Определите топики и модели сообщений

**Файл:** `internal/kafka/topics.go`

```go
package kafka

import (
    kafkaClient "go-test-framework/pkg/kafka/client"
    "go-test-framework/pkg/kafka/dsl"
)

// Приватная переменная для хранения клиента
var client *kafkaClient.Client

// Link для Auto-Wiring (аналогично HTTP и DB)
type Link struct{}

func (l *Link) SetKafka(c *kafkaClient.Client) {
    client = c

    // Регистрируем связь топика с моделью
    dsl.Register[PlayerEventMessage](c, "player-events")
}

func Client() *kafkaClient.Client {
    return client
}

// Определите тип топика (используется как generic параметр)
type PlayerEventsTopic string

const PlayerEventsTopicName PlayerEventsTopic = "game-player-events"

// TopicName реализует интерфейс topic.TopicName
func (PlayerEventsTopic) TopicName() string {
    return string(PlayerEventsTopicName)
}

// Модель сообщения
type PlayerEventMessage struct {
    PlayerID   string `json:"playerId"`
    EventType  string `json:"eventType"`
    PlayerName string `json:"playerName"`
}
```

### 3. Подключите в `test_env.go`

```go
package tests

import (
    "my-project/internal/kafka"
    "my-project/internal/client/game"
    "my-project/internal/db/players"
    "go-test-framework/pkg/builder"
    "log"
)

type TestEnv struct {
    // HTTP клиенты
    GameService game.Link `config:"gameService"`

    // Database
    PlayersRepo players.Link `db_config:"coreDatabase"`

    // Kafka
    Kafka kafka.Link `kafka_config:"kafka"`
}

var env *TestEnv

func init() {
    env = &TestEnv{}

    if err := builder.BuildEnv(env); err != nil {
        log.Fatalf("Failed to build environment: %v", err)
    }
}
```

---

## Полный E2E тест: HTTP → Database → Kafka

Теперь соберём все три шага вместе. Это **полноценный E2E сценарий**:

```go
import (
    "my-project/internal/kafka"
    "my-project/internal/client/game"
    "my-project/internal/db/players"
    kafkaDSL "go-test-framework/pkg/kafka/dsl"
)

func (s *PlayerSuite) TestCreatePlayerE2E(t provider.T) {
    t.Title("E2E: Создание игрока (HTTP → DB → Kafka)")

    var playerID string
    var username = "pro_gamer_2024"

    // ШАГ 1: HTTP - Создаём игрока через API
    s.Step(t, "HTTP: Создание игрока через API", func(sCtx provider.StepCtx) {
        resp := game.CreatePlayer(sCtx).
            // 1. Настройка запроса (строгая типизация)
            RequestBody(models.CreatePlayerReq{
                Username: username,
                Region:   "EU",
            }).
            // 2. Проверки ответа
            ExpectResponseStatus(201).
            ExpectResponseBodyFieldNotEmpty("id").
            ExpectResponseBodyFieldValue("username", username).
            ExpectResponseBodyFieldValue("status", "active").
            // 3. Выполнение
            Send()

        // Сохраняем ID для проверки в БД и Kafka
        playerID = resp.Body.ID
    })

    // ШАГ 2: Database - Проверяем запись в таблице players
    s.Step(t, "Database: Проверка записи в БД", func(sCtx provider.StepCtx) {
        players.FindByID(sCtx, playerID). // Используем ID из HTTP
            ExpectFound().
            ExpectColumnEquals("username", username).
            ExpectColumnEquals("status", "active").
            ExpectColumnEquals("region", "EU").
            ExpectColumnFalse("is_vip").
            ExpectColumnIsNotNull("created_at").
            Send()
    })

    // ШАГ 3: Kafka - Ждём событие PLAYER_CREATED
    s.Step(t, "Kafka: Ожидание события PLAYER_CREATED", func(sCtx provider.StepCtx) {
        kafkaDSL.Expect[kafka.PlayerEventsTopic](sCtx, kafka.Client()).
            // Фильтры для поиска нужного события
            With("playerId", playerID).           // Используем ID из HTTP
            With("eventType", "PLAYER_CREATED").  // Тип события

            // Проверка уникальности (нет дубликатов)
            Unique().

            // Проверки полей события
            ExpectField("playerName", username).  // Имя совпадает
            ExpectFieldNotEmpty("timestamp").     // Дата заполнена
            ExpectFieldTrue("isActive").          // Флаг активности

            // Выполнение (поиск)
            Send()
    })
}
```

**Результат E2E теста:** Один тест проверил три слоя системы — HTTP API, базу данных и брокер сообщений.

---

## Справочник методов Kafka DSL

### Создание ожидания

| Метод | Описание |
|:---|:---|
| `Expect[TTopic](sCtx, client)` | Создает ожидание сообщения из топика |

**Параметр:**
- `TTopic` - тип топика (string-based type с именем топика)

### Фильтры (для поиска)

| Метод | Описание |
|:---|:---|
| `.With(key, value)` | Добавляет фильтр для поиска сообщения |

**Примеры:**
- `.With("playerId", "123")` - простое поле
- `.With("player.id", "123")` - вложенное поле
- `.With("status", "ACTIVE")` - строка
- `.With("amount", 100)` - число

**Логика:** AND (все фильтры должны совпасть)

### Уникальность

| Метод | Описание |
|:---|:---|
| `.Unique()` | Проверяет уникальность в окне (5 сек из конфига) |
| `.UniqueWithWindow(duration)` | Кастомное окно уникальности |

**Как работает:** Если найдено >1 сообщение в окне → тест падает

### Проверки полей (Expectations)

| Метод | Описание |
|:---|:---|
| `.ExpectField(field, value)` | Поле равно значению |
| `.ExpectFieldNotEmpty(field)` | Поле не пустое |
| `.ExpectFieldIsNull(field)` | Поле = null |
| `.ExpectFieldIsNotNull(field)` | Поле ≠ null |
| `.ExpectFieldTrue(field)` | Поле = true |
| `.ExpectFieldFalse(field)` | Поле = false |

**Синтаксис путей (GJSON):**
- Простое поле: `"playerName"`
- Вложенное поле: `"player.name"`
- Элемент массива: `"items.0"`
- Подсчёт элементов: `"items.#"`
- Вложенное в массиве: `"users.0.email"`

### Выполнение

| Метод | Описание |
|:---|:---|
| `.Send()` | Выполняет поиск и проверки (ничего не возвращает) |

**Что происходит:**
1. Ищет сообщение по фильтрам (`With`)
2. Если async режим → retry с интервалами
3. Проверяет уникальность (если `Unique()`)
4. Выполняет все expectations (`ExpectField*`)
5. Прикрепляет результат в Allure
6. Падает если что-то не сошлось

---

# Асинхронные шаги (AsyncStep)

## Проблема: Flaky-тесты в асинхронных системах

Современные микросервисы работают асинхронно: события приходят в Kafka с задержкой, базы данных обновляются не мгновенно, API возвращает статус `PENDING`, который через время меняется на `COMPLETED`. Классические синхронные тесты в таких условиях становятся нестабильными (flaky), они падают, потому что данные ещё не готовы.

Типичная ситуация без AsyncStep:

```go
// Проблема: тест упадёт, если запись в БД ещё не появилась
s.Step(t, "Проверка в БД", func(sCtx provider.StepCtx) {
    players.FindByID(sCtx, playerID).
        ExpectFound().  // Упадёт с ошибкой "record not found"
        Send()
})
```

Разработчик добавляет `time.Sleep(5 * time.Second)` — но это:
- Замедляет тесты (даже если данные готовы за 100ms)
- Не гарантирует успех (в нагруженной системе может потребоваться больше времени)
- Создаёт нестабильность (flaky tests)

---

## Решение: AsyncStep с автоматическими retry

AsyncStep **автоматически повторяет проверки** до успеха или таймаута:

```go
// Решение: AsyncStep будет повторять запрос, пока запись не появится
s.AsyncStep(t, "Проверка в БД", func(sCtx provider.StepCtx) {
    players.FindByID(sCtx, playerID).
        ExpectFound().  // Если не найдено → retry через interval
        ExpectColumnEquals("status", "active").
        Send()
})
```

Разработчик добавляет `time.Sleep(5 * time.Second)`, но это:
- Замедляет тесты (даже если данные готовы за 100ms)
- Не гарантирует успех (в нагруженной системе может потребоваться больше времени)
- Создаёт нестабильность (flaky tests)

## Решение: AsyncStep с автоматическими retry

AsyncStep автоматически повторяет проверки до успеха или таймаута:

```go
// Решение: AsyncStep будет повторять запрос, пока запись не появится
s.AsyncStep(t, "Проверка в БД", func(sCtx provider.StepCtx) {
    players.FindByID(sCtx, playerID).
        ExpectFound().  // Если не найдено, retry через interval
        ExpectColumnEquals("status", "active").
        Send()
})
```

**Как это работает:**
1. Выполняет запрос, проверки не прошли, ждёт `interval` (например, 200ms)
2. Повторяет запрос, проверки всё ещё не прошли, ждёт с увеличенным интервалом (backoff)
3. Повторяет... пока проверки не пройдут или не истечёт `timeout`

## Параллельное выполнение

AsyncStep предоставляет два ключевых преимущества:
- **Автоматические retry с умными стратегиями ожидания** (backoff & jitter)
- **Параллельное выполнение независимых операций**

Рядом стоящие `AsyncStep` автоматически запускаются параллельно в отдельных goroutines, аналогично [allure-go](https://github.com/ozontech/allure-go). Это дает огромный выигрыш по времени для множественных операций.

**Примеры реального ускорения:**
- 3 Kafka события с polling 30s: **90s -> 30s** (3x быстрее)
- 3 HTTP-запроса по 200ms: **600ms -> 200ms** (3x быстрее)
- 5 DB-запросов по 100ms: **500ms -> 100ms** (5x быстрее)

### Пример: Множественные Kafka events

** Последовательное выполнение (старый подход):**
```go
// Каждый шаг с polling 30 секунд
s.Step(t, "Wait event 1", func(sCtx provider.StepCtx) {
    kafkaDSL.Expect[Topic](sCtx, client).With("id", "1").Send() // 30s
})
s.Step(t, "Wait event 2", func(sCtx provider.StepCtx) {
    kafkaDSL.Expect[Topic](sCtx, client).With("id", "2").Send() // 30s
})
s.Step(t, "Wait event 3", func(sCtx provider.StepCtx) {
    kafkaDSL.Expect[Topic](sCtx, client).With("id", "3").Send() // 30s
})
// Итого: 90 секунд
```

** Параллельное выполнение (AsyncStep):**
```go
// Все три шага запускаются ОДНОВРЕМЕННО
s.AsyncStep(t, "Wait event 1", func(sCtx provider.StepCtx) {
    kafkaDSL.Expect[Topic](sCtx, client).With("id", "1").Send() // polling 30s
})
s.AsyncStep(t, "Wait event 2", func(sCtx provider.StepCtx) {
    kafkaDSL.Expect[Topic](sCtx, client).With("id", "2").Send() // polling 30s параллельно!
})
s.AsyncStep(t, "Wait event 3", func(sCtx provider.StepCtx) {
    kafkaDSL.Expect[Topic](sCtx, client).With("id", "3").Send() // polling 30s параллельно!
})

// Sync шаг автоматически ЖДЁТ завершения всех 3 async шагов
s.Step(t, "Verify events received", func(sCtx provider.StepCtx) {
    // Все 3 события уже получены!
})
// Итого: ~30 секунд (время самого долгого)
// Выигрыш: 3x быстрее
```

### Как это работает под капотом

```go
// Внутренняя реализация (автоматическая)
s.AsyncStep(t, "Task 1", fn1) // asyncWg.Add(1) → go func() { defer wg.Done(); fn1() }()
s.AsyncStep(t, "Task 2", fn2) // asyncWg.Add(1) → go func() { defer wg.Done(); fn2() }()
s.AsyncStep(t, "Task 3", fn3) // asyncWg.Add(1) → go func() { defer wg.Done(); fn3() }()

// Все 3 goroutines работают ОДНОВРЕМЕННО

s.Step(t, "Verify", fn4)      // asyncWg.Wait() → ждёт все 3 → выполняется fn4
```

### Реальный выигрыш по времени

#### Сценарий 1: Множественные HTTP-запросы
```
3 HTTP-запроса по 200ms каждый:
  Последовательно: 200ms + 200ms + 200ms = 600ms
  Параллельно:     max(200ms, 200ms, 200ms) = ~200ms
  Выигрыш: 3x быстрее
```

#### Сценарий 2: Kafka polling с timeout 30s
```
3 события с polling 30s каждое:
  Последовательно: 30s + 30s + 30s = 90 секунд
  Параллельно:     max(30s, 30s, 30s) = ~30 секунд
  Выигрыш: 3x быстрее
```

#### Сценарий 3: Смешанные операции (HTTP + DB + Kafka)
```
HTTP (500ms) + DB (300ms) + Kafka (5s):
  Последовательно: 500ms + 300ms + 5000ms = 5.8 секунд
  Параллельно:     max(500ms, 300ms, 5000ms) = ~5 секунд
  Выигрыш: ~15% быстрее, но главное - масштабируется
```

### Автоматическая синхронизация

**Sync шаги автоматически ждут все async шаги:**

```go
// Batch 1: 3 параллельных HTTP-запроса
s.AsyncStep(t, "Fetch users", ...)
s.AsyncStep(t, "Fetch products", ...)
s.AsyncStep(t, "Fetch orders", ...)

// Этот sync шаг дождётся завершения всех 3 запросов выше
s.Step(t, "Verify data loaded", func(sCtx provider.StepCtx) {
    // Все 3 запроса гарантированно завершены
})

// Batch 2: 2 параллельных DB-запроса
s.AsyncStep(t, "Query table A", ...)
s.AsyncStep(t, "Query table B", ...)

// Снова автоматическая синхронизация
s.Step(t, "Verify DB state", func(sCtx provider.StepCtx) {
    // Оба DB-запроса завершены
})
```

**Ключевые моменты:**
- Не нужно вручную управлять `sync.WaitGroup`
- Не нужно вызывать `.Wait()` явно
- `Step()` автоматически ждёт все предыдущие `AsyncStep`
- `AfterEach()` гарантирует завершение всех goroutines в конце теста

**Результат:**
```
Step 1 started at: 14:23:10.100
Step 2 started at: 14:23:10.101  <- Все стартуют одновременно!
Step 3 started at: 14:23:10.102  <- Разница <2ms!

Total execution time: 2.01s
PROOF: Steps ran in PARALLEL (would be 6s if sequential)
```

Подробнее см. файл [`parallel_proof_test.go`](./parallel_proof_test.go) с полными примерами и доказательствами.

## Разница между Step и AsyncStep

| Аспект | Step | AsyncStep |
|:---|:---|:---|
| **Выполнение** | Последовательно, синхронно | **Параллельно** в goroutines |
| **Ожидание** | Выполняется сразу | Ждёт перед собой все async шаги |
| **Поведение при неудаче** | Падает сразу (`Require`) | Повторяет попытки (`Assert`) |
| **Retry** | Нет | Есть (по конфигу, если есть expectations) |
| **Когда использовать** | API вернул ошибку 4xx/5xx | Ждём события Kafka, запись в БД, изменение статуса |
| **Главное преимущество** | Точные проверки | **Параллелизм + Retry** |

**Технически:**
- `Step` использует `Require` → остановка теста при первой ошибке
- `AsyncStep` использует `Assert` → накопление ошибок и повторные попытки
- `Step` вызывает `asyncWg.Wait()` → ждёт завершения всех async шагов перед выполнением
- `AfterEach` вызывает `asyncWg.Wait()` → гарантирует завершение всех goroutines

**Пример:**
```go
// Эти 3 шага выполнятся ПАРАЛЛЕЛЬНО
s.AsyncStep(t, "Task 1", ...) // Запустился
s.AsyncStep(t, "Task 2", ...) // Запустился параллельно с Task 1
s.AsyncStep(t, "Task 3", ...) // Запустился параллельно с Task 1 и 2

// Этот шаг дождётся ВСЕХ трёх async шагов
s.Step(t, "Verify", ...) // asyncWg.Wait() -> выполняется после завершения всех
```

## Конфигурация

Настройки async режима задаются **отдельно для каждого DSL** в `config.local.yaml`:

```yaml
# HTTP клиенты
http:
  gameService:
    baseURL: "https://game-api.example.com"
    timeout: 30s

# База данных
database:
  coreDatabase:
    driver: "postgres"
    dsn: "postgres://user:password@localhost:5432/game_db?sslmode=disable"
    maxOpenConns: 10
    maxIdleConns: 5

# Kafka
kafka:
  bootstrapServers:
    - "kafka.example.com:9092"
  groupId: "qa-test-group"
  topics:
    - "game-player-events"
  bufferSize: 1000
  uniqueDuplicateWindowMs: 5000

# HTTP DSL - для асинхронных API запросов
http_dsl:
  async:
    enabled: true          # Включить async режим
    timeout: 10s           # Максимальное время ожидания
    interval: 200ms        # Начальный интервал между попытками
    backoff:
      enabled: true        # Включить экспоненциальное увеличение интервала
      factor: 1.5          # Множитель (200ms → 300ms → 450ms → ...)
      max_interval: 1s     # Максимальный интервал (не больше 1 секунды)
    jitter: 0.2            # Случайное отклонение ±20% (против синхронизации)

# Database DSL - для проверок в БД
db_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
    backoff:
      enabled: true
      factor: 1.5
      max_interval: 1s
    jitter: 0.2

# Kafka DSL - для ожидания событий
kafka_dsl:
  async:
    enabled: true
    timeout: 30s           # Kafka может требовать больше времени
    interval: 200ms
    backoff:
      enabled: true
      factor: 1.5
      max_interval: 1s
    jitter: 0.2
```

### Параметры конфигурации

| Параметр | Описание | Пример |
|:---|:---|:---|
| `enabled` | Включить/выключить async режим | `true` |
| `timeout` | Максимальное время ожидания | `10s`, `30s` |
| `interval` | Начальный интервал между попытками | `200ms`, `500ms` |
| `backoff.enabled` | Включить экспоненциальное увеличение интервала | `true` |
| `backoff.factor` | Множитель для каждой попытки (1.5 = +50%) | `1.5`, `2.0` |
| `backoff.max_interval` | Максимальный интервал (ограничитель роста) | `1s`, `2s` |
| `jitter` | Случайное отклонение (0.2 = ±20% от интервала) | `0.2` |

### Как работает Backoff

Без backoff (fixed interval):
```
Попытка 1, ждём 200ms, Попытка 2, ждём 200ms, Попытка 3...
```

С backoff (factor = 1.5):
```
Попытка 1, ждём 200ms, Попытка 2, ждём 300ms, Попытка 3, ждём 450ms... max 1s
```

### Зачем нужен Jitter

Jitter добавляет случайность, чтобы избежать "эффекта стада" (thundering herd):
- Без jitter: 100 тестов запрашивают БД одновременно каждые 200ms, пиковая нагрузка
- С jitter 0.2: запросы распределены от 160ms до 240ms, плавная нагрузка

## Примеры использования

### HTTP DSL: Ожидание изменения статуса

API возвращает `202 Accepted` и статус `PENDING`. Нужно дождаться статуса `COMPLETED`:

```go
s.AsyncStep(t, "Wait for order completion", func(sCtx provider.StepCtx) {
    orders.GetOrder(sCtx, orderID).
        ExpectResponseStatus(200).
        ExpectResponseBodyFieldValue("status", "COMPLETED"). // Retry пока не станет COMPLETED
        ExpectResponseBodyFieldNotEmpty("completedAt").
        Send()
})
```

### Database DSL: Ожидание появления записи

После HTTP запроса запись в БД может появиться с задержкой (триггеры, очереди):

```go
// Создали заказ через API
orderID := createOrderViaAPI()

// Ждём, пока запись появится в БД
s.AsyncStep(t, "Wait for order in DB", func(sCtx provider.StepCtx) {
    orders.FindByID(sCtx, orderID).
        ExpectFound().                              // Retry пока не найдётся
        ExpectColumnEquals("status", "pending").
        Send()
})
```

### Kafka DSL: Ожидание события

Событие в Kafka может прийти через несколько секунд после API-запроса:

```go
// Создали игрока через API
playerID := createPlayerViaAPI()

// Ждём событие в Kafka
s.AsyncStep(t, "Wait for PLAYER_CREATED event", func(sCtx provider.StepCtx) {
    kafkaDSL.Expect[kafka.PlayerEventsTopic](sCtx, kafka.Client()).
        With("playerId", playerID).
        With("eventType", "PLAYER_CREATED").
        ExpectField("playerName", "test_user").
        Send()
})
```

### Комбинирование sync и async шагов

В одном тесте можно смешивать оба типа:

```go
func (s *OrderSuite) TestCreateOrder(t provider.T) {
    var orderID string

    // Синхронный шаг: HTTP запрос должен вернуть 201 немедленно
    s.Step(t, "Create order via API", func(sCtx provider.StepCtx) {
        resp := orders.CreateOrder(sCtx).
            RequestBody(models.OrderRequest{...}).
            ExpectResponseStatus(201).  // Падаем сразу, если не 201
            Send()
        orderID = resp.Body.ID
    })

    // Асинхронный шаг: ждём появления в БД
    s.AsyncStep(t, "Wait for order in DB", func(sCtx provider.StepCtx) {
        orders.FindByID(sCtx, orderID).
            ExpectFound().  // Retry
            Send()
    })

    // Асинхронный шаг: ждём событие в Kafka
    s.AsyncStep(t, "Wait for ORDER_CREATED event", func(sCtx provider.StepCtx) {
        kafkaDSL.Expect[kafka.OrdersTopic](sCtx, kafka.Client()).
            With("orderId", orderID).
            Send()
    })
}
```

## Polling Summary в Allure

Каждый `AsyncStep` автоматически прикрепляет **Polling Summary** к отчёту Allure:

```json
{
  "attempts": 5,
  "elapsed_time": "1.234s",
  "success": true
}
```

Если проверки не прошли:

```json
{
  "attempts": 15,
  "elapsed_time": "10.001s",
  "success": false,
  "failed_checks": [
    "Expected status 'COMPLETED', got 'PENDING'",
    "Field 'completedAt' is empty"
  ],
  "timeout_reason": "Timeout or context cancelled",
  "last_error": "context deadline exceeded"
}
```

Это помогает анализировать, почему тест упал:
- Было недостаточно времени?
- Данные вообще не пришли?
- Пришли, но с неправильными значениями?

## Рекомендации

### Когда использовать AsyncStep

**Используйте AsyncStep:**
- **Параллельное выполнение независимых операций** (множественные HTTP/DB/Kafka запросы)
- Ожидание событий в Kafka
- Проверка появления записей в БД после асинхронных операций
- Ожидание изменения статуса через API (PENDING, COMPLETED)
- Проверка побочных эффектов (триггеры, очереди, фоновые задачи)

**Не используйте AsyncStep:**
- Зависимые операции (результат первого нужен второму, будет race condition)
- Синхронные API запросы, которые должны вернуть ошибку сразу (используйте `Step` с `Require`)
- Negative-тесты (проверка, что запись НЕ появилась)

### Лучшие практики параллельности

#### Хорошо: Независимые операции параллельно
```go
// Эти запросы независимы — запускаем параллельно
s.AsyncStep(t, "Fetch user", func(sCtx provider.StepCtx) {
    users.GetUser(sCtx, "123").Send()
})
s.AsyncStep(t, "Fetch product", func(sCtx provider.StepCtx) {
    products.GetProduct(sCtx, "456").Send()
})
s.AsyncStep(t, "Fetch settings", func(sCtx provider.StepCtx) {
    settings.GetSettings(sCtx).Send()
})

// Sync шаг дождётся всех трёх запросов
s.Step(t, "Verify all data loaded", func(sCtx provider.StepCtx) {
    // Все данные уже загружены
})
```

#### Плохо: Зависимые операции с AsyncStep
```go
// НЕ ДЕЛАЙТЕ ТАК! Будет race condition
var userID string

s.AsyncStep(t, "Create user", func(sCtx provider.StepCtx) {
    resp := users.CreateUser(sCtx).Send()
    userID = resp.Body.ID
})

//  ОШИБКА: userID может быть пустым или гонка данных
s.AsyncStep(t, "Update user", func(sCtx provider.StepCtx) {
    users.UpdateUser(sCtx, userID).Send() // Race condition!
})
```

#### Правильно: Зависимые операции с Step
```go
var userID string

// Синхронный шаг: дождёмся создания
s.Step(t, "Create user", func(sCtx provider.StepCtx) {
    resp := users.CreateUser(sCtx).Send()
    userID = resp.Body.ID
})

// Теперь безопасно: userID гарантированно заполнен
s.Step(t, "Update user", func(sCtx provider.StepCtx) {
    users.UpdateUser(sCtx, userID).Send()
})
```

#### Хорошо: Батчи async + sync checkpoints
```go
// Batch 1: Параллельная подготовка данных
s.AsyncStep(t, "Setup DB", ...)
s.AsyncStep(t, "Setup Kafka", ...)
s.AsyncStep(t, "Setup Cache", ...)

// Checkpoint: дождёмся завершения подготовки
s.Step(t, "Verify setup complete", ...)

// Batch 2: Параллельные тесты разных фич
s.AsyncStep(t, "Test feature A", ...)
s.AsyncStep(t, "Test feature B", ...)
s.AsyncStep(t, "Test feature C", ...)

// Final checkpoint
s.Step(t, "Verify all tests passed", ...)
```

### Настройка timeout

- **HTTP DSL:** 10-15 секунд (большинство API отвечают быстро)
- **Database DSL:** 10-15 секунд (триггеры обычно быстрые)
- **Kafka DSL:** 30-60 секунд (брокер может иметь задержку доставки)

### Оптимизация скорости тестов

1. **Начинайте с малого `interval`:** 100-200ms достаточно для большинства случаев
2. **Включайте backoff:** экономит ресурсы при длительном ожидании
3. **Ограничивайте max_interval:** не больше 1-2 секунд
4. **Используйте jitter:** снижает пиковую нагрузку на тестовую среду

---

# Параметризованные тесты (Table-Driven Tests)

Параметризованные тесты позволяют запустить один и тот же тест с разными наборами данных. Это особенно полезно для негативных тестов, когда нужно проверить множество граничных случаев и валидаций.

## Зачем нужны параметризованные тесты

### Проблема: Дублирование кода

Без параметризации приходится писать отдельный тест для каждого случая:

```go
func (s *RegisterSuite) TestEmptyEmail(t provider.T) {
    auth.Register(sCtx).
        RequestBody(models.RegisterRequest{Email: "", Password: "P@ss"}).
        ExpectResponseStatus(422).
        ExpectResponseBodyFieldValue("detail.code", "EMAIL_IS_EMPTY").
        Send()
}

func (s *RegisterSuite) TestInvalidEmail(t provider.T) {
    auth.Register(sCtx).
        RequestBody(models.RegisterRequest{Email: "invalid", Password: "P@ss"}).
        ExpectResponseStatus(422).
        ExpectResponseBodyFieldValue("detail.code", "INVALID_EMAIL").
        Send()
}

// ... еще 10 похожих тестов
```

### Решение: Table-Driven подход с Allure-Go

Фреймворк использует встроенную поддержку параметризации из `allure-go`. Allure автоматически обнаруживает параметры и запускает тест для каждого набора данных.

---

## Сквозной пример: Негативное тестирование регистрации

Проверим валидацию email при регистрации пользователя с множеством невалидных входных данных.

### Шаг 1: Создайте структуру для test case

```go
type EmailTestCase struct {
    Name           string
    Email          string
    Password       string
    ExpectedStatus int
    ExpectedCode   string
    ExpectedDetail string
    ExpectedField  string
}
```

### Шаг 2: Добавьте параметры в Suite

```go
type RegisterNegativeSuite struct {
    extension.BaseSuite
    ParamEmailValidation []EmailTestCase  // Имя ОБЯЗАТЕЛЬНО должно быть Param + <название из метода>
}
```

**ВАЖНО:** Имя поля должно соответствовать паттерну: `Param` + название из метода `TableTest<Название>`.

Если метод называется `TableTestEmailValidation`, то поле должно быть `ParamEmailValidation`.

### Шаг 3: Инициализируйте данные в BeforeAll

```go
func (s *RegisterNegativeSuite) BeforeAll(t provider.T) {
    s.ParamEmailValidation = []EmailTestCase{
        {
            Name:           "Email field is empty",
            Email:          "",
            Password:       "P@ssw0rd",
            ExpectedStatus: http.StatusUnprocessableEntity,
            ExpectedCode:   "EMAIL_IS_EMPTY",
            ExpectedDetail: "Данное поле обязательно.",
            ExpectedField:  "email",
        },
        {
            Name:           "Email without @",
            Email:          "invalid-email.com",
            Password:       "P@ssw0rd",
            ExpectedStatus: http.StatusUnprocessableEntity,
            ExpectedCode:   "INVALID_EMAIL_FORMAT",
            ExpectedDetail: "Email должен содержать @ и точку в доменной части",
            ExpectedField:  "email",
        },
        {
            Name:           "Email without domain",
            Email:          "test@",
            Password:       "P@ssw0rd",
            ExpectedStatus: http.StatusUnprocessableEntity,
            ExpectedCode:   "INVALID_EMAIL_FORMAT",
            ExpectedDetail: "Email должен содержать @ и точку в доменной части",
            ExpectedField:  "email",
        },
    }
}
```

### Шаг 4: Создайте Table метод

```go
func (s *RegisterNegativeSuite) TableTestEmailValidation(t provider.T, tc EmailTestCase) {
    t.Title(tc.Name)

    s.Step(t, "Send registration request with invalid email", func(sCtx provider.StepCtx) {
        auth_service.Register(sCtx).
            RequestBody(authModels.RegisterRequest{
                Email:    tc.Email,
                Password: tc.Password,
            }).
            ExpectResponseStatus(tc.ExpectedStatus).
            ExpectResponseBodyFieldValue("detail.code", tc.ExpectedCode).
            ExpectResponseBodyFieldValue("detail.detail", tc.ExpectedDetail).
            ExpectResponseBodyFieldValue("detail.field", tc.ExpectedField).
            Send()
    })
}
```

**Важные детали:**
- Имя метода **ОБЯЗАТЕЛЬНО** должно начинаться с `TableTest`
- Второй параметр - это структура test case
- Allure автоматически запустит метод для каждого элемента в `ParamEmailValidation`

### Шаг 5: Запустите Suite

```go
func TestRegisterNegativeSuite(t *testing.T) {
    suite.RunSuite(t, new(RegisterNegativeSuite))  // Используйте RunSuite, не RunNamedSuite
}
```

**ВАЖНО:** Для параметризованных тестов используйте `suite.RunSuite`, а не `suite.RunNamedSuite`.

---

## Использование RequestBodyMap для негативных тестов

Часто в негативных тестах нужно отправить запрос **без определенных полей** (например, проверить валидацию отсутствующего поля).

### Проблема с типизированными структурами

```go
type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// Даже если передать пустую строку, поле всё равно будет в JSON
auth.Register(sCtx).
    RequestBody(RegisterRequest{Password: "P@ss"})
// Отправит: {"email": "", "password": "P@ss"}
// Поле email присутствует, просто пустое!
```

### Решение: RequestBodyMap

Используйте `.RequestBodyMap()` вместо `.RequestBody()`:

```go
s.Step(t, "Register without email field", func(sCtx provider.StepCtx) {
    auth_service.Register(sCtx).
        RequestBodyMap(map[string]interface{}{
            "password": "P@ssw0rd",
            // email полностью отсутствует в JSON
        }).
        ExpectResponseStatus(http.StatusBadRequest).
        ExpectResponseBodyFieldValue("detail.code", "EMAIL_REQUIRED").
        Send()
})
```

Отправленный JSON:
```json
{
  "password": "P@ssw0rd"
}
```

### Другие сценарии использования RequestBodyMap

**1. Дополнительные поля (проверка на лишние поля):**
```go
.RequestBodyMap(map[string]interface{}{
    "email":    "test@test.com",
    "password": "P@ssw0rd",
    "extra":    "unexpected_field",
})
```

**2. Невалидные типы данных:**
```go
.RequestBodyMap(map[string]interface{}{
    "email":    123,  // number вместо string
    "password": true, // boolean вместо string
})
```

**3. Null значения:**
```go
.RequestBodyMap(map[string]interface{}{
    "email":    nil,
    "password": "P@ssw0rd",
})
```

---

## Полный пример с RequestBodyMap

```go
type FieldAbsenceTestCase struct {
    Name           string
    Body           map[string]interface{}
    ExpectedStatus int
    ExpectedCode   string
}

type RegisterNegativeSuite struct {
    extension.BaseSuite
    ParamFieldAbsence []FieldAbsenceTestCase
}

func (s *RegisterNegativeSuite) BeforeAll(t provider.T) {
    s.ParamFieldAbsence = []FieldAbsenceTestCase{
        {
            Name:           "Email field is missing",
            Body:           map[string]interface{}{"password": "P@ssw0rd"},
            ExpectedStatus: http.StatusBadRequest,
            ExpectedCode:   "EMAIL_REQUIRED",
        },
        {
            Name:           "Password field is missing",
            Body:           map[string]interface{}{"email": "test@test.com"},
            ExpectedStatus: http.StatusBadRequest,
            ExpectedCode:   "PASSWORD_REQUIRED",
        },
        {
            Name:           "Both fields missing",
            Body:           map[string]interface{}{},
            ExpectedStatus: http.StatusBadRequest,
            ExpectedCode:   "VALIDATION_ERROR",
        },
    }
}

func (s *RegisterNegativeSuite) TableTestFieldAbsence(t provider.T, tc FieldAbsenceTestCase) {
    t.Title(tc.Name)

    s.Step(t, "Send request with missing fields", func(sCtx provider.StepCtx) {
        auth_service.Register(sCtx).
            RequestBodyMap(tc.Body).
            ExpectResponseStatus(tc.ExpectedStatus).
            ExpectResponseBodyFieldValue("detail.code", tc.ExpectedCode).
            Send()
    })
}
```

---

## Рекомендации

1. **Именование:** Используйте описательные имена для test cases (`Name` поле)
2. **Группировка:** Группируйте похожие проверки в один параметризованный тест
3. **RequestBodyMap vs RequestBody:**
   - `RequestBody` - для позитивных тестов и случаев с полными данными
   - `RequestBodyMap` - для негативных тестов с отсутствующими/лишними полями
4. **Не смешивайте:** Не используйте одновременно `RequestBody` и `RequestBodyMap` - только один из них
5. **Allure отчеты:** Каждый test case появится как отдельный тест в Allure с названием из поля `Name`

---

# Маскировка чувствительных данных

Тесты часто работают с конфиденциальной информацией: токенами авторизации, паролями, API ключами. Эти данные попадают в Allure отчёты, которые могут быть доступны широкому кругу лиц. Фреймворк предоставляет механизм **настраиваемой маскировки** чувствительных данных в HTTP запросах, SQL запросах и результатах из БД.

---

## Зачем нужна маскировка

### Проблема: Утечка credentials в отчётах

Без маскировки Allure отчёт содержит полные значения:

```
HTTP Request
Headers:
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  
  Cookie: session_id=abc123def456...                          

SQL Query:
  UPDATE users SET password = ? WHERE id = ?
Arguments:
  [1] MySecretPassword123  
  [2] 42
```

Эти данные могут быть использованы для несанкционированного доступа.

### Решение: Настраиваемая маскировка

С настроенной маскировкой отчёт безопасен:

```
HTTP Request
Headers:
  Authorization: Bearer ***MASKED***  
  Cookie: ***MASKED***                

SQL Query:
  UPDATE users SET password = ? WHERE id = ?
Arguments:
  [1] ***MASKED***  
  [2] 42
```

---

## Принцип работы

**По умолчанию ничего не маскируется.** Вы должны явно указать в конфигурации, какие заголовки и колонки нужно скрыть.

**Маскировка работает по точному совпадению имён:**
- HTTP заголовок `Authorization` будет маскирован только если в конфиге указано именно `Authorization`
- Колонка БД `user_password` будет маскирована только если в конфиге указано `user_password`
- Частичные совпадения НЕ работают: конфиг `password` не замаскирует колонку `user_password_hash`

---

## Конфигурация маскировки

### HTTP: Маскировка заголовков

Укажите заголовки, которые нужно маскировать:

```yaml
http:
  gameService:
    baseURL: "https://game-api.example.com"
    timeout: 30s
    maskHeaders: "Authorization,Cookie,X-API-Key"  # Через запятую
```

**Как это работает:**
- Список заголовков разделяется запятыми
- Регистр не важен (case-insensitive): `Authorization` = `authorization`
- Совпадение **точное**: `Authorization` не замаскирует `X-Authorization`
- Для заголовка `Authorization` сохраняется префикс (Bearer/Basic), маскируется только токен
- Остальные заголовки заменяются полностью на `***MASKED***`

### Database: Маскировка колонок

Укажите колонки, которые нужно маскировать в SQL результатах:

```yaml
database:
  coreDatabase:
    driver: "postgres"
    dsn: "postgres://user:password@localhost:5432/game_db?sslmode=disable"
    maxOpenConns: 10
    maxIdleConns: 5
    maskColumns: "password,api_token,secret_key"  # Через запятую
```

**Как это работает:**
- Список колонок разделяется запятыми
- Регистр не важен (case-insensitive): `password` = `PASSWORD`
- Совпадение **точное**: `password` не замаскирует `user_password` или `password_hash`
- Если нужно замаскировать `user_password`, укажите его явно: `maskColumns: "password,user_password"`

---

## Примеры маскировки

### HTTP: Заголовки

**Конфиг:**
```yaml
http:
  authService:
    baseURL: "https://auth.example.com"
    maskHeaders: "Authorization,X-Admin-Secret"
```

**Allure отчёт (HTTP Request):**
```
Headers:
  Content-Type: application/json
  Authorization: Bearer ***MASKED***       # Из конфига (сохранён префикс Bearer)
  X-Admin-Secret: ***MASKED***             # Из конфига
  X-Request-ID: abc-123-def-456            # Не маскируется
```

### Database: Колонки в результатах

**Конфиг:**
```yaml
database:
  coreDatabase:
    driver: "mysql"
    dsn: "user:pass@tcp(localhost:3306)/db"
    maskColumns: "password,api_token"  # Точные имена колонок
```

**SQL запрос и результат:**
```sql
SELECT id, username, password, email, api_token FROM users WHERE id = 1
```

**Allure отчёт (SQL Result):**
```json
{
  "id": 1,
  "username": "john_doe",
  "password": "***MASKED***",         // Маскировано (точное совпадение с конфигом)
  "email": "john@example.com",        // Не маскируется
  "api_token": "***MASKED***"         // Маскировано (точное совпадение с конфигом)
}
```

**Что НЕ будет замаскировано при `maskColumns: "password"`:**
- Колонка `user_password` (другое имя)
- Колонка `password_hash` (другое имя)
- Колонка `reset_password_token` (другое имя)

Для маскировки этих колонок нужно указать их явно:
```yaml
maskColumns: "password,user_password,password_hash,reset_password_token"
```

---

## Полный пример конфигурации

```yaml
# HTTP клиенты
http:
  gameService:
    baseURL: "https://game-api.example.com"
    timeout: 30s
    maskHeaders: "Authorization,Cookie,X-API-Key"  # Чувствительные заголовки

# База данных
database:
  coreDatabase:
    driver: "postgres"
    dsn: "postgres://user:password@localhost:5432/game_db?sslmode=disable"
    maxOpenConns: 10
    maxIdleConns: 5
    maskColumns: "password,user_password,api_token,secret_key,private_key"  # Точные имена колонок

# Kafka (маскировка не требуется - события публичны)
kafka:
  bootstrapServers:
    - "kafka.example.com:9092"
  groupId: "qa-test-group"
  topics:
    - "game-player-events"
  bufferSize: 1000
```

---

## Важные замечания

### ️ Маскировка только в отчётах

Маскировка применяется **только к данным в Allure отчётах**. Реальные HTTP запросы и SQL запросы выполняются с оригинальными значениями.

```go
// В тесте используются реальные значения
httpClient.Do(req.Header("Authorization", "Bearer real-token-12345"))

// В Allure отчёте будет: Authorization: Bearer ***MASKED***
// Но реальный HTTP запрос уйдёт с: Authorization: Bearer real-token-12345
```

### Что маскируется

- HTTP заголовки (request и response)
- SQL аргументы в отчёте запросов
- Значения колонок в SQL результатах

### Что НЕ маскируется

- Тела JSON запросов/ответов (если содержат sensitive поля, их нужно не логировать)
- Kafka сообщения (предполагается, что события не содержат credentials)
- Логи приложения (фреймворк не контролирует логирование самого приложения)

---

## Рекомендации

### Указывайте точные имена колонок

Маскировка работает только по точному совпадению:

```yaml
#  Неправильно: password не замаскирует user_password или password_hash
database:
  coreDatabase:
    maskColumns: "password"

#  Правильно: укажите все нужные колонки
database:
  coreDatabase:
    maskColumns: "password,user_password,password_hash,reset_password_token"
```

### Настраивайте maskColumns для production-like БД

Если тесты работают с чувствительными данными:

```yaml
database:
  productionDB:
    maskColumns: "password,user_password,secret,api_token,access_token,refresh_token,ssn,credit_card_number"
```

### Не храните credentials в коде

Используйте переменные окружения:

```yaml
http:
  apiService:
    baseURL: "https://api.example.com"
    defaultHeaders:
      Authorization: "Bearer ${API_TOKEN}"  # Читается из env переменной
```

### Проверяйте отчёты перед публикацией

Даже с маскировкой:
- Проверяйте, что в отчётах нет случайных credentials
- Не публикуйте отчёты в открытый доступ без проверки
- Ограничивайте доступ к Allure серверу

---
# Генерация тестовых данных

Тесты часто требуют уникальных данных для каждого запуска: email адреса, пароли, случайные строки. Модуль `pkg/datagen` предоставляет простой API для генерации тестовых данных с гарантией соответствия требованиям валидации.

---

## Проблема: Хардкод данных в тестах

### Что происходит без генератора

Классический подход - использовать хардкод значения или timestamp:

```go
func TestRegisterUser(t provider.T) {
    // Проблема 1: Хардкод - при повторном запуске упадёт (email exists)
    email := "test@example.com"
    password := "P@ssw0rd"

    // Проблема 2: Timestamp - работает, но нечитаемо
    email := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
    password := "P@ssw0rd"  // Всегда один и тот же
}
```

**Проблемы:**
- Хардкод приводит к падениям при повторных запусках (дубликаты)
- Timestamp создаёт длинные нечитаемые значения в отчётах
- Пароли одинаковые - не тестируют реальную вариативность
- Невозможно генерировать специфичные невалидные данные для негативных тестов

---

## Решение: Генератор данных

Модуль `pkg/datagen` автоматически создаёт уникальные валидные данные:

```go
import "go-test-framework/pkg/datagen"

func TestRegisterUser(t provider.T) {
    // Решение: Генерация валидных данных
    email := datagen.Email(10)      // "xK7mP2nQaB@generated.com"
    password := datagen.Password(8) // "A3!kL9@z" (гарантия: digit + upper + lower + special)

    auth.Register(sCtx).
        RequestBody(models.RegisterRequest{
            Email:    email,
            Password: password,
        }).
        Send()
}
```

**Преимущества:**
- **Уникальность:** Каждый запуск создаёт новые данные
- **Валидность:** Пароли гарантированно содержат все требуемые типы символов
- **Читаемость:** Короткие значения в отчётах Allure
- **Гибкость:** Настройка через charset для негативных тестов

---

## Справочник функций

### Email(length int) string

Генерирует случайный email адрес.

**Параметры:**
- `length` - длина локальной части (до @). Если ≤ 0, используется 10.

**Примеры:**
```go
datagen.Email(10)  // "xK7mP2nQaB@generated.com"
datagen.Email(5)   // "aB3Xz@generated.com"
datagen.Email(0)   // "kL9pQw2MnV@generated.com" (default 10)
```

**Формат:** `<random_alphanumeric>@generated.com`

---

### Password(length int, charsets ...string) string

Генерирует случайный пароль с гарантией включения символов из КАЖДОГО переданного charset.

**Параметры:**
- `length` - длина пароля. Если меньше количества charsets, автоматически увеличивается.
- `charsets` (variadic) - наборы символов. Если не указаны, используются: `Digits`, `LatinUpper`, `LatinLower`, `SpecialChars`.

**Логика:**
1. Берётся минимум 1 символ из каждого переданного charset
2. Остальные позиции заполняются случайными символами из всех charsets
3. Результат перемешивается (shuffle)

**Примеры:**

```go
// Дефолтный пароль (Digits + LatinUpper + LatinLower + SpecialChars)
datagen.Password(8)  // "A3!kL9@z" - гарантия всех типов

// Пароль без uppercase (для негативного теста)
datagen.Password(8, datagen.LatinLower, datagen.Digits, datagen.SpecialChars)
// "a7b@2k!9" - только lowercase, digits, special

// Пароль без цифр (для негативного теста)
datagen.Password(8, datagen.LatinUpper, datagen.LatinLower, datagen.SpecialChars)
// "AbK@zLp!" - только буквы и спецсимволы

// Пароль только из букв (без special chars - негативный тест)
datagen.Password(8, datagen.LatinUpper, datagen.LatinLower, datagen.Digits)
// "A3kL9pZx" - буквы + цифры, без спецсимволов

// Слишком короткий пароль (негативный тест)
datagen.Password(4)  // "A3!k" - длина 4 (минимальная для 4 charsets)
```

---

### String(length int, charsets ...string) string

Генерирует случайную строку из указанных наборов символов.

**Параметры:**
- `length` - длина строки. Если ≤ 0, используется 10.
- `charsets` (variadic) - наборы символов. Если не указаны, используется `Alphanumeric`.

**Примеры:**

```go
// Только цифры
datagen.String(5, datagen.Digits)  // "72849"

// Только буквы (lowercase)
datagen.String(8, datagen.LatinLower)  // "abkzpmqw"

// Буквы + цифры
datagen.String(10, datagen.LatinLower, datagen.Digits)  // "a7bx2kp9mq"

// Специальные символы
datagen.String(6, datagen.SpecialChars)  // "!@#$%^"

// Дефолт (alphanumeric)
datagen.String(10)  // "xK7mP2nQaB"
```

---

## Доступные константы (Charsets)

Используйте эти константы для настройки генерации:

| Константа | Значение | Описание |
|:----------|:---------|:---------|
| `Digits` | `"0123456789"` | Цифры |
| `LatinLower` | `"abcdefghijklmnopqrstuvwxyz"` | Строчные латинские буквы |
| `LatinUpper` | `"ABCDEFGHIJKLMNOPQRSTUVWXYZ"` | Заглавные латинские буквы |
| `LatinLetters` | `LatinLower + LatinUpper` | Все латинские буквы |
| `Alphanumeric` | `LatinLetters + Digits` | Буквы + цифры |
| `SpecialChars` | <code>\`~!@#$%^&*()-_=+[]{}\|;:'",<.>/?</code> | Все спецсимволы английской раскладки |

---

## Примеры использования

### Позитивные тесты: Валидные данные

```go
func (s *RegisterSuite) TestRegisterNewUser(t provider.T) {
    email := datagen.Email(10)
    password := datagen.Password(8)

    s.Step(t, "Register user", func(sCtx provider.StepCtx) {
        auth.Register(sCtx).
            RequestBody(models.RegisterRequest{
                Email:    email,
                Password: password,
            }).
            ExpectResponseStatus(201).
            Send()
    })
}
```

### Негативные тесты: Невалидные пароли

Используйте комбинации charsets для создания специфичных невалидных данных:

```go
func (s *RegisterNegativeSuite) BeforeAll(t provider.T) {
    validEmail := datagen.Email(10)

    s.ParamPasswordValidation = []PasswordTestCase{
        {
            Name:     "Password too short",
            Email:    validEmail,
            Password: datagen.Password(4),  // Слишком короткий
            ExpectedCode: "INVALID_PASSWORD",
        },
        {
            Name:     "Password without uppercase",
            Email:    validEmail,
            Password: datagen.Password(8, datagen.LatinLower, datagen.Digits, datagen.SpecialChars),
            ExpectedCode: "INVALID_PASSWORD",
        },
        {
            Name:     "Password without digits",
            Email:    validEmail,
            Password: datagen.Password(8, datagen.LatinUpper, datagen.LatinLower, datagen.SpecialChars),
            ExpectedCode: "INVALID_PASSWORD",
        },
        {
            Name:     "Password without special characters",
            Email:    validEmail,
            Password: datagen.Password(8, datagen.LatinUpper, datagen.LatinLower, datagen.Digits),
            ExpectedCode: "INVALID_PASSWORD",
        },
        {
            Name:     "Password only lowercase",
            Email:    validEmail,
            Password: datagen.Password(8, datagen.LatinLower),
            ExpectedCode: "INVALID_PASSWORD",
        },
    }
}
```

### Генерация уникальных строк

```go
func (s *GameSuite) TestCreatePlayer(t provider.T) {
    // Уникальное имя игрока
    playerName := datagen.String(12, datagen.LatinLetters)

    // Уникальный ID (только цифры)
    customID := datagen.String(8, datagen.Digits)

    s.Step(t, "Create player", func(sCtx provider.StepCtx) {
        game.CreatePlayer(sCtx).
            RequestBody(models.CreatePlayerReq{
                Name: playerName,
                CustomID: customID,
            }).
            Send()
    })
}
```

### BeforeAll для переиспользования

Если один пароль нужен в нескольких тестах - используйте **локальные переменные**:

```go
type RegisterNegativeSuite struct {
    extension.BaseSuite
    ParamEmailValidation []EmailTestCase
}

func (s *RegisterNegativeSuite) BeforeAll(t provider.T) {
    // Генерируем локально - данные копируются в массив
    validPassword := datagen.Password(8)
    validEmail := datagen.Email(10)

    s.ParamEmailValidation = []EmailTestCase{
        {
            Name:     "Empty email",
            Email:    "",
            Password: validPassword,  // Копируется в тест-кейс
            ExpectedCode: "EMAIL_IS_EMPTY",
        },
        {
            Name:     "Invalid email format",
            Email:    "invalid-email",
            Password: validPassword,  // Тот же пароль, скопирован в тест-кейс
            ExpectedCode: "INVALID_EMAIL_FORMAT",
        },
    }
}
```

**Почему локальные переменные, а не поля структуры?**
- Thread-safe: данные "замораживаются" при создании массива
- Нет глобального состояния между тестами
- Безопасно для параллельного выполнения
---
