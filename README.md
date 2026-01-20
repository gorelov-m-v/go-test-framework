# Go Test Framework

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?logo=postgresql&logoColor=white)
![MySQL](https://img.shields.io/badge/MySQL-4479A1?logo=mysql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?logo=redis&logoColor=white)
![Kafka](https://img.shields.io/badge/Kafka-231F20?logo=apachekafka&logoColor=white)
![HTTP](https://img.shields.io/badge/HTTP-REST-green)
![gRPC](https://img.shields.io/badge/gRPC-244c5a)
![Allure](https://img.shields.io/badge/Allure-Report-yellow)

Декларативный DSL-фреймворк для E2E тестирования на Go, построенный поверх [allure-go](https://github.com/ozontech/allure-go) от OzonTech.

## Пример

```go
func (s *PlayerSuite) TestCreatePlayer(t provider.T) {
    var resp *client.Response[models.CreatePlayerResp]

    s.Step(t, "Create player via API", func(sCtx provider.StepCtx) {
        resp = httpGame.CreatePlayer(sCtx).
            RequestBody(models.CreateReq{Username: "test_user"}).
            ExpectStatus(201).
            Send()
    })

    s.AsyncStep(t, "Verify in database", func(sCtx provider.StepCtx) {
        dbPlayers.FindByID(sCtx, resp.Body.ID).
            ExpectFound().
            ExpectColumnEquals("status", "active").
            Send()
    })

    s.AsyncStep(t, "Verify Kafka event", func(sCtx provider.StepCtx) {
        kafka.Expect[events.PlayerCreated](sCtx).
            With("playerId", resp.Body.ID).
            Send()
    })
}
```

## Оглавление

- [Проблемы и решения](#проблемы-и-решения)
- [DSL](#dsl)
    - [HTTP](#http)
        - [Сквозной E2E пример](#сквозной-e2e-пример-шаг-1---создание-игрока)
        - [Справочник](#справочник-методов-http-dsl)
    - [Database](#database)
        - [Сквозной E2E пример](#сквозной-e2e-пример-шаг-21---проверка-в-бд)
        - [Справочник](#справочник-методов-db-dsl)
    - [Kafka](#kafka)
        - [Сквозной E2E пример](#сквозной-e2e-пример-шаг-22---проверка-события)
        - [Справочник](#справочник-методов-kafka-dsl)
    - [Redis](#redis)
        - [Сквозной E2E пример](#сквозной-e2e-пример-шаг-23---проверка-кэша)
        - [Справочник](#справочник-методов-redis-dsl)
    - [gRPC](#grpc)
        - [Сквозной E2E пример](#сквозной-e2e-пример-шаг-3---верификация-через-grpc)
        - [Справочник](#справочник-методов-grpc-dsl)
    - [Полный E2E тест](#полный-e2e-тест)
    - [Кодогенерация](#кодогенерация)
        - [OpenAPI Generator](#openapi-generator-openapi-gen)
        - [gRPC Generator](#grpc-generator-grpc-gen)
        - [Автоматизация генерации](#автоматизация-генерации)
- [Расширенные возможности](#расширенные-возможности)
    - [Асинхронные шаги (AsyncStep)](#асинхронные-шаги-asyncstep)
        - [Зачем это нужно](#зачем-это-нужно)
        - [Параметры конфигурации](#параметры-конфигурации)
        - [Когда использовать AsyncStep](#когда-использовать-asyncstep)
        - [Лучшие практики параллельности](#лучшие-практики-параллельности)
    - [Параметризованные тесты (Table-Driven Tests)](#параметризованные-тесты-table-driven-tests)
        - [Зачем это нужно](#зачем-это-нужно-1)
        - [Сквозной пример](#сквозной-пример-негативное-тестирование-регистрации)
    - [Маскировка чувствительных данных](#маскировка-чувствительных-данных)
        - [Зачем это нужно](#зачем-это-нужно-2)
        - [HTTP: Маскировка заголовков](#http-маскировка-заголовков)
        - [Database: Маскировка колонок](#database-маскировка-колонок)
    - [Контрактное тестирование (Contract Testing)](#контрактное-тестирование-contract-testing)
        - [Зачем нужно контрактное тестирование](#зачем-нужно-контрактное-тестирование)
        - [Конфигурация контрактного тестирования](#конфигурация-контрактного-тестирования)
        - [Использование контрактного тестирования](#использование-контрактного-тестирования)
    - [Генерация тестовых данных](#генерация-тестовых-данных)
        - [Зачем это нужно](#зачем-это-нужно-3)
        - [Email](#emaillength-int-string)
        - [Password](#passwordlength-int-charsets-string-string)
        - [String](#stringlength-int-charsets-string-string)
    - [BaseSuite и Cleanup](#basesuite-и-cleanup)
        - [Структура BaseSuite](#структура-basesuite)
        - [Методы Step и AsyncStep](#методы-step-и-asyncstep)
        - [Автоматический Cleanup](#автоматический-cleanup)
    - [Настройка Allure](#настройка-allure)
- [Быстрый старт](#быстрый-старт)
- [Рекомендуемая структура проекта](#рекомендуемая-структура-проекта)

---

## Проблемы и решения

Фреймворк решает типичные проблемы E2E автоматизации на Go.

### Проблемы

1. **Отставание автоматизации от разработки** — Автотесты появляются после релиза фич. Инженеры тратят время на написание клиентов, DTO, мапперов вместо покрытия сценариев.

2. **Сложность подключения инфраструктуры** — Добавление нового сервиса или БД требует десятков строк настроечного кода, который дублируется и ломается.

3. **Высокий порог входа** — Написание E2E тестов требует глубоких знаний HTTP-клиентов, работы с БД, Kafka, обработки ошибок. Новым инженерам сложно начать.

4. **Сложность написания и поддержки тестов** — Суть тестовых сценариев теряется за технической реализацией — обработка ошибок, парсинг данных, настройка транспортов.

5. **Технологический разрыв** — Использование разных стеков (Go для разработки, Python/Java для тестов) приводит к дублированию моделей и рассинхронизации логики.

6. **Медленные пайплайны** — Последовательное выполнение IO-операций блокирует CI/CD.

7. **Нестабильные тесты (Flaky)** — Данные появляются в БД или Kafka не сразу. Тесты с `time.Sleep()` нестабильны.

### Решения

1. **Contract-Based Code Gen** — Генераторы создают типизированные клиенты из контрактов (OpenAPI, Protobuf). Shift-Left подход.

2. **Spring-like DI** — Подключение сервисов, баз данных или брокеров — одна строка в конфиге.

3. **Низкий порог входа** — Декларативный DSL скрывает сложность. Инженер описывает *что* проверить, а не *как*.

4. **Declarative Fluent DSL** — Сценарии описываются декларативно (`Request → Expect → Verify`), превращая код теста в исполняемую спецификацию.

5. **Единый стек** — Тесты пишутся на том же языке, что и сервисы. Разработчики и QA работают в одной кодовой базе.

6. **Multi-Level Concurrency** — Параллельное выполнение на уровне тестовых наборов и шагов внутри теста.

7. **Smart Async & Polling** — Retry с Backoff & Jitter интегрированы в ядро. Автоматическое ожидание данных.

## DSL

Фреймворк предоставляет три типизированных DSL для работы с основными компонентами системы.

## HTTP

Модуль предназначен для написания функциональных тестов REST API.
Он построен на **Generics**, что гарантирует строгую типизацию запросов и ответов на этапе компиляции. Вы не сможете отправить неверную структуру или ошибиться в типе ожидаемого ответа.

### Сквозной E2E пример: Шаг 1 - Создание игрока

Начнём полноценный E2E сценарий, который пройдёт через **все 5 DSL**.

**Сценарий:** Создаём игрока через HTTP API → асинхронно проверяем в БД, Kafka и Redis → финально верифицируем через gRPC.

```
┌─────────────────────────────────────────────────────────────────────┐
│  Шаг 1: HTTP POST /players                                          │
│  → Создаём игрока, получаем ID                                      │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Шаг 2: Параллельные AsyncStep (запускаются ОДНОВРЕМЕННО)           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │ 2.1 Database │  │ 2.2 Kafka    │  │ 2.3 Redis    │               │
│  │  (retry)     │  │  (retry)     │  │  (retry)     │               │
│  └──────────────┘  └──────────────┘  └──────────────┘               │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Шаг 3: gRPC GetPlayer                                              │
│  → Финальная проверка через другой протокол                         │
└─────────────────────────────────────────────────────────────────────┘
```

### 0. Конфигурация (`config.local.yaml`)

```yaml
http:
  gameService:
    baseURL: "https://game-api.example.com"
    timeout: 30s
```

### 1. Спецификация (Контракт)

Представим, что нам нужно зарегистрировать игрока. Вот описание эндпоинта:

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
**Файл:** `internal/http_client/game/client.go`

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
    "my-project/internal/http_client/game"
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
```

### 5. Тест

Это **Шаг 1** нашего сквозного сценария. Мы отправляем HTTP-запрос и **сохраняем ответ целиком** для использования в следующих шагах.

```go
func (s *PlayerSuite) TestCreatePlayerFullE2E(t provider.T) {
    t.Title("Full E2E: HTTP → [DB + Kafka + Redis] → gRPC")

    var (
        httpResp *client.Response[models.CreatePlayerResp]
        username = "pro_gamer_2024"
    )

    // ═══════════════════════════════════════════════════════════════
    // ШАГ 1: HTTP - Создаём игрока через API
    // ═══════════════════════════════════════════════════════════════
    s.Step(t, "HTTP: Создание игрока", func(sCtx provider.StepCtx) {
        httpResp = game.CreatePlayer(sCtx).
            RequestBody(models.CreatePlayerReq{
                Username: username,
                Region:   "EU",
            }).
            ExpectResponseStatus(201).
            ExpectResponseBodyFieldNotEmpty("id").
            ExpectResponseBodyFieldValue("username", username).
            ExpectResponseBodyFieldValue("status", "active").
            Send()
    })

    // Шаг 2.1: Database - см. раздел "Database" ниже
    // Шаг 2.2: Kafka - см. раздел "Kafka" ниже
    // Шаг 2.3: Redis - см. раздел "Redis" ниже
    // Шаг 3: gRPC - см. раздел "gRPC" ниже
}
```

---

### Справочник методов HTTP DSL

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

#### Проверка null/not null
*   `.ExpectResponseBodyFieldIsNull(path string)` — Проверяет, что поле существует и равно `null`.
*   `.ExpectResponseBodyFieldIsNotNull(path string)` — Проверяет, что поле существует и НЕ равно `null`.

#### Проверка boolean true/false
*   `.ExpectResponseBodyFieldTrue(path string)` — Проверяет, что поле существует и равно `true`.
*   `.ExpectResponseBodyFieldFalse(path string)` — Проверяет, что поле существует и равно `false`.

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
| `"created_at"` | не `null` | `.ExpectResponseBodyFieldIsNotNull("created_at")` |
| `"updated_at"` | `null` | `.ExpectResponseBodyFieldIsNull("updated_at")` |
| `"is_active"` | `true` | `.ExpectResponseBodyFieldTrue("is_active")` |
| `"is_verified"` | `false` | `.ExpectResponseBodyFieldFalse("is_verified")` |

**Поддерживаемый синтаксис путей:**
- Простые поля: `"name"`
- Вложенные поля: `"user.email"`
- Элементы массива: `"items.0"`, `"items.1"`
- Подсчёт элементов: `"items.#"`
- Вложенные поля в массиве: `"users.0.name"`

**Поддерживаемые типы сравнения:**
*   `string`: `"active"`
*   `int`, `float`: `100`, `99.99` (авто-конвертация между `int`, `int16`, `int32`, `int64`, `float32`, `float64`)
*   `bool`: `true`, `false`
*   `nil`: Проверяет, что поле в JSON равно `null` или отсутствует.

#### Проверка объектов в массиве (struct matching)

Для проверки наличия объекта в JSON-массиве используйте методы с Go-структурами:

*   `.ExpectArrayContains(path string, expected any)` — **Partial match**: проверяет только non-zero поля структуры.
*   `.ExpectArrayContainsExact(path string, expected any)` — **Exact match**: проверяет ВСЕ поля включая zero values (`""`, `0`, `false`).

**Пример ответа API:**
```json
{
  "items": [
    {"id": "123", "name": "Sports", "gamesCount": 0, "isDefault": false},
    {"id": "456", "name": "Casino", "gamesCount": 10, "isDefault": true}
  ]
}
```

**Partial match** — проверяет только заполненные поля:
```go
// Найдёт объект с id="123", остальные поля игнорируются
resp.ExpectArrayContains("items", Category{
    Id: "123",
})
```

**Exact match** — проверяет все поля включая zero values:
```go
// Найдёт объект где ВСЕ поля совпадают, включая gamesCount=0 и isDefault=false
resp.ExpectArrayContainsExact("items", Category{
    Id:         "123",
    Name:       "Sports",
    GamesCount: 0,      // zero value проверяется!
    IsDefault:  false,  // zero value проверяется!
})
```

**Поддерживаемые типы полей:**
- Примитивы: `string`, `int`, `bool`, `float`
- Указатели: `*string`, `*int` (автоматическое разыменование)
- Maps: `map[string]interface{}`, `map[string]string`
- Вложенные структуры
- Slices

**Обработка несоответствий контракта:**
Если поле есть в Go-структуре, но отсутствует в JSON — оно пропускается (не вызывает ошибку).

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

### 4. RequestBodyMap (для негативных тестов)

Часто в негативных тестах нужно отправить запрос **без определенных полей** (например, проверить валидацию отсутствующего поля).

**Проблема с типизированными структурами:**

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

**Решение:** Используйте `.RequestBodyMap()` вместо `.RequestBody()`:

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

**Сценарии использования:**

*   **Отсутствующие поля:**
    ```go
    .RequestBodyMap(map[string]interface{}{
        "password": "P@ssw0rd",
        // email отсутствует
    })
    ```

*   **Дополнительные поля:**
    ```go
    .RequestBodyMap(map[string]interface{}{
        "email":    "test@test.com",
        "password": "P@ssw0rd",
        "extra":    "unexpected_field",
    })
    ```

*   **Невалидные типы данных:**
    ```go
    .RequestBodyMap(map[string]interface{}{
        "email":    123,  // number вместо string
        "password": true, // boolean вместо string
    })
    ```

*   **Null значения:**
    ```go
    .RequestBodyMap(map[string]interface{}{
        "email":    nil,
        "password": "P@ssw0rd",
    })
    ```

---

## Database

Модуль `pkg/database/dsl` предназначен для **верификации состояния** базы данных после выполнения бизнес-операций.
Он поддерживает PostgreSQL и MySQL.

---

### Сквозной E2E пример: Шаг 2.1 - Проверка в БД

**Продолжаем E2E сценарий.** После создания игрока (Шаг 1) проверяем, что запись попала в базу данных.
Это первый из трёх **параллельных AsyncStep**, которые запустятся одновременно.

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
    schemas:                    # Маппинг алиасов схем
      core: "production_core"   # alias -> реальное имя схемы
```

#### Schemas (маппинг схем БД)

Параметр `schemas` позволяет использовать в коде алиасы вместо реальных имён схем. Это полезно когда:
- На разных стендах схемы называются по-разному (`beta-10_core`, `prod_core`)
- Вы хотите чтобы имена в коде соответствовали MCP/документации

**Использование в репозитории:**
```go
func FindByID(sCtx provider.StepCtx, id string) *dsl.Query[PlayerRow] {
    return dsl.NewQuery[PlayerRow](sCtx, dbClient).
        SQL(fmt.Sprintf("SELECT * FROM `%s`.players WHERE id = ?", dbClient.Schema("core")), id)
}
```

При конфиге `schemas.core: "beta-10_core"` запрос будет: `SELECT * FROM beta-10_core.players WHERE id = ?`

**Ожидаемое состояние таблицы `players`:**

| id | username | status | region | is_vip | created_at |
|:---|:---------|:-------|:-------|:-------|:-----------|
| uuid-12345 | pro_gamer_2024 | active | EU | false | 2024-01-09 10:30:00 |

### 1. Описание Моделей
Опишите Go-структуру, соответствующую таблице. Используйте тег `db` для маппинга колонок.

**Файл:** `internal/models/db_player.go`

```go
package models

import (
    "database/sql"
    "time"
)

type PlayerRow struct {
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
func FindByID(sCtx provider.StepCtx, id string) *dsl.Query[models.PlayerRow] {
    return dsl.NewQuery[models.PlayerRow](sCtx, dbClient).
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
    "my-project/internal/http_client/game"
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
```

### 4. Тест

Добавляем **Шаг 2.1** к нашему E2E-тесту. Используем `httpResp.Body.ID` из Шага 1.

> **Важно:** Используем `AsyncStep` — это позволяет запустить проверку параллельно с Kafka и Redis.

```go
func (s *PlayerSuite) TestCreatePlayerFullE2E(t provider.T) {
    // ... Шаг 1: HTTP (см. выше) ...

    // ═══════════════════════════════════════════════════════════════
    // ШАГ 2.1: Database - Проверяем запись (AsyncStep с retry)
    // ═══════════════════════════════════════════════════════════════
    s.AsyncStep(t, "Database: Проверка записи", func(sCtx provider.StepCtx) {
        players.FindByID(sCtx, httpResp.Body.ID).
            ExpectFound().
            ExpectColumnEquals("username", username).
            ExpectColumnEquals("status", "active").
            ExpectColumnEquals("region", "EU").
            ExpectColumnFalse("is_vip").
            ExpectColumnIsNotNull("created_at").
            Send()
    })

    // Шаг 2.2: Kafka - см. раздел "Kafka" ниже
    // Шаг 2.3: Redis - см. раздел "Redis" ниже
    // Шаг 3: gRPC - см. раздел "gRPC" ниже
}
```

---

### Справочник методов DB DSL

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
*   `.ExpectColumnEquals("col", val)` — Сравнивает значение (поддерживает `*string`, `sql.Null*` типы).
*   `.ExpectColumnTrue("col")` / `.ExpectColumnFalse("col")` — Для boolean полей.
*   `.ExpectColumnIsNull("col")` / `.ExpectColumnIsNotNull("col")` — Для `sql.Null*` типов.
*   `.ExpectColumnEmpty("col")` — Проверяет, что колонка NULL или пустая строка.
*   `.ExpectColumnJsonEquals("col", map[string]interface{})` — Сравнивает JSON-поле с ожидаемым map.

**Примечание:** Имена колонок (`"col"`) должны совпадать с тегом `db` в вашей модели.

**Авто-конвертация числовых типов:** DSL автоматически сравнивает числа разных типов (`int`, `int16`, `int32`, `int64`, `float64` и т.д.). Используйте простые числа в константах:

```go
// ✅ Правильно
const StatusEnabled = 1
ExpectColumnEquals("status_id", StatusEnabled)

// ❌ Избыточно
const StatusEnabled = int16(1)
```

### 3. Выполнение

*   `.Send()` — **Финализирующий метод.**
    1.  Выполняет SQL запрос (`GetContext`).
    2.  Сканирует первую строку результата в структуру `Model`.
    3.  Запускает все проверки.
    4.  Создает шаг в Allure с Query и Result.
    5.  Возвращает заполненную структуру `Model`.

---

## Kafka

Модуль `pkg/kafka/dsl` предназначен для **верификации событий** в Apache Kafka.
Он построен на фоновом consumer'е, который непрерывно читает сообщения в буфер, позволяя тестам искать события с retry-логикой.

---

### Сквозной E2E пример: Шаг 2.2 - Проверка события

**Продолжаем E2E сценарий.** Проверяем, что система отправила событие `PLAYER_CREATED` в Kafka.
Это второй **параллельный AsyncStep**, который запустится одновременно с Database и Redis.

### 0. Конфигурация (`config.local.yaml`)

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
  topicPrefix: "beta-10_"        # Префикс для всех топиков
  topics:
    - "core.player-events"       # Без префикса - он добавится автоматически
  bufferSize: 1000
  uniqueDuplicateWindowMs: 5000
```

#### TopicPrefix (префикс топиков)

Параметр `topicPrefix` автоматически добавляется ко всем топикам. Это полезно когда:
- На разных стендах топики имеют разные префиксы (`beta-10_`, `prod_`)
- Вы хотите чтобы имена топиков в коде соответствовали MCP/документации

С конфигом выше:
- В `topics` указан `core.player-events`
- Реальный топик в Kafka будет `beta-10_core.player-events`
- В коде `TopicName()` возвращает `core.player-events` (без префикса)
- Фреймворк автоматически добавит prefix при поиске сообщений

### 1. Описание Моделей

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
}

func Client() *kafkaClient.Client {
    return client
}

// Определите тип топика (используется как generic параметр)
type PlayerEventsTopic string

// TopicName реализует интерфейс topic.TopicName
// Возвращает имя БЕЗ префикса - он добавится автоматически из конфига
func (PlayerEventsTopic) TopicName() string {
    return "core.player-events"
}

// Модель сообщения
type PlayerEventMessage struct {
    PlayerID   string `json:"playerId"`
    EventType  string `json:"eventType"`
    PlayerName string `json:"playerName"`
}
```

### 2. Подключение в Env

```go
package tests

import (
    "my-project/internal/kafka"
    "my-project/internal/http_client/game"
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

### 3. Тест

Добавляем **Шаг 2.2** к нашему E2E-тесту.

```go
func (s *PlayerSuite) TestCreatePlayerFullE2E(t provider.T) {
    // ... Шаг 1: HTTP (см. выше) ...
    // ... Шаг 2.1: Database (см. выше) ...

    // ═══════════════════════════════════════════════════════════════
    // ШАГ 2.2: Kafka - Ждём событие PLAYER_CREATED (AsyncStep с retry)
    // ═══════════════════════════════════════════════════════════════
    s.AsyncStep(t, "Kafka: Ожидание события", func(sCtx provider.StepCtx) {
        kafkaDSL.Expect[kafka.PlayerEventsTopic](sCtx, kafka.Client()).
            With("playerId", httpResp.Body.ID).
            With("eventType", "PLAYER_CREATED").
            Unique().
            ExpectField("playerName", username).
            ExpectFieldNotEmpty("timestamp").
            Send()
    })

    // Шаг 2.3: Redis - см. раздел "Redis" ниже
    // Шаг 3: gRPC - см. раздел "gRPC" ниже
}
```

---

### Справочник методов Kafka DSL

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

### Количество сообщений

| Метод | Описание |
|:---|:---|
| `.ExpectCount(n)` | Ожидает ровно N сообщений по фильтрам |

**Как работает:**
1. Ищет все сообщения по фильтрам `.With()`
2. В async режиме → retry пока не найдёт минимум N сообщений
3. Проверяет что найдено **ровно N** (не больше, не меньше)
4. Аттачит **все найденные сообщения** как JSON массив в Allure
5. Применяет `ExpectField*` проверки к **первому** (самому свежему) сообщению

**Пример использования:**
```go
// Ожидаем 2 сообщения: первое от Create, второе от Delete
s.AsyncStep(t, "Verify Kafka Messages", func(sCtx provider.StepCtx) {
    kafkaDSL.Expect[kafka.GameTopic](sCtx, kafka.Client()).
        With("category.uuid", categoryId).
        With("category.status", "disabled").
        ExpectCount(2).
        ExpectField("message.eventType", "category").
        Send()
})
```

### Проверки полей (Expectations)

| Метод | Описание |
|:---|:---|
| `.ExpectField(field, value)` | Поле равно значению |
| `.ExpectFieldNotEmpty(field)` | Поле не пустое |
| `.ExpectFieldIsNull(field)` | Поле = null |
| `.ExpectFieldIsNotNull(field)` | Поле ≠ null |
| `.ExpectFieldTrue(field)` | Поле = true |
| `.ExpectFieldFalse(field)` | Поле = false |

**Авто-конвертация числовых типов:** При сравнении чисел DSL автоматически конвертирует типы (`int`, `int16`, `int64`, `float64` и т.д.).

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

## Redis

Модуль `pkg/redis/dsl` предназначен для **верификации состояния кэша** после выполнения бизнес-операций.
Он позволяет проверять наличие ключей, их значения (включая JSON) и TTL.

---

### Сквозной E2E пример: Шаг 2.3 - Проверка кэша

**Продолжаем E2E сценарий.** Проверяем, что данные игрока попали в Redis-кэш.
Это третий **параллельный AsyncStep**, который запустится одновременно с Database и Kafka.

### 0. Конфигурация (`config.local.yaml`)

Расширяем конфигурацию — добавляем секцию `redis`:

```yaml
redis:
  cache:
    addr: "localhost:6379"
    password: ""
    db: 0

redis_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
    backoff: { enabled: true, factor: 1.5, max_interval: 1s }
```

### 1. Реализация Клиента

**Файл:** `internal/cache/player/client.go`

```go
package player

import (
    "github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
    "github.com/gorelov-m-v/go-test-framework/pkg/redis/dsl"
    "github.com/ozontech/allure-go/pkg/framework/provider"
)

var redisClient *client.Client

type Link struct{}

func (l *Link) SetRedis(c *client.Client) {
    redisClient = c
}

func Client() *client.Client {
    return redisClient
}

func GetByID(sCtx provider.StepCtx, playerID string) *dsl.Query {
    return dsl.NewQuery(sCtx, redisClient).
        Key("player:" + playerID)
}
```

### 2. Подключение в Env

Расширяем `tests/env.go` — добавляем Redis к уже существующим клиентам.

```go
type TestEnv struct {
    // HTTP
    GameService game.Link `config:"gameService"`
    // Database
    PlayersRepo players.Link `db_config:"coreDatabase"`
    // Kafka
    Kafka kafka.Link `kafka_config:"kafka"`
    // Redis - связываем конфиг "redis.cache" с пакетом "player"
    PlayerCache player.Link `redis_config:"redis.cache"`
}
```

### 3. Тест

Добавляем **Шаг 2.3** к нашему E2E-тесту.

```go
func (s *PlayerSuite) TestCreatePlayerFullE2E(t provider.T) {
    // ... Шаг 1: HTTP (см. выше) ...
    // ... Шаг 2.1: Database (см. выше) ...
    // ... Шаг 2.2: Kafka (см. выше) ...

    // ═══════════════════════════════════════════════════════════════
    // ШАГ 2.3: Redis - Проверяем кэш (AsyncStep с retry)
    // ═══════════════════════════════════════════════════════════════
    s.AsyncStep(t, "Redis: Проверка кэша", func(sCtx provider.StepCtx) {
        playerCache.GetByID(sCtx, httpResp.Body.ID).
            ExpectExists().
            ExpectJSONField("username", username).
            ExpectJSONField("status", "active").
            ExpectTTL(4*time.Minute, 6*time.Minute).
            Send()
    })

    // Шаг 3: gRPC - см. раздел "gRPC" ниже
}
```

---

### Справочник методов Redis DSL

Объект `dsl.Query` предоставляет методы для проверки состояния Redis-ключей.

### 1. Настройка запроса

| Метод | Описание | Пример |
|:---|:---|:---|
| `.Key(name)` | Ключ для проверки | `.Key("player:123")` |
| `.Context(ctx)` | Установить контекст | `.Context(ctx)` |
| `.StepName(name)` | Кастомное имя шага | `.StepName("Check cache")` |

### 2. Ожидания (Expectations)

**Проверки наличия:**
*   `.ExpectExists()` — Ключ существует.
*   `.ExpectNotExists()` — Ключ не существует.

**Проверки значения:**
*   `.ExpectValue("value")` — Точное строковое значение.
*   `.ExpectValueNotEmpty()` — Непустое значение.

**Проверки JSON-полей (GJSON Path):**
*   `.ExpectJSONField("path", value)` — Значение поля в JSON.
*   `.ExpectJSONFieldNotEmpty("path")` — Непустое JSON-поле.

**Проверки TTL:**
*   `.ExpectTTL(min, max)` — TTL в диапазоне.
*   `.ExpectNoTTL()` — Без TTL (persistent key).

### 3. Выполнение

*   `.Send()` — Выполняет GET, проверяет expectations, возвращает `*client.Result`.

### 4. Утилиты для setup/cleanup

```go
redisClient := playerCache.Client()
err := redisClient.Set(ctx, "key", "value", 5*time.Minute)
err := redisClient.Del(ctx, "key1", "key2")
```

---

## gRPC

Модуль `pkg/grpc/dsl` предназначен для тестирования gRPC сервисов.
Он построен на **Generics**, что гарантирует строгую типизацию protobuf-сообщений на этапе компиляции.

---

### Сквозной E2E пример: Шаг 3 - Верификация через gRPC

**Финальный шаг E2E сценария.** После создания игрока (Шаг 1) и параллельных проверок (Шаг 2.x),
верифицируем данные через **другой протокол** — gRPC.

> **Важно:** `s.Step` **ожидает завершения всех предыдущих AsyncStep** перед выполнением.

### 0. Конфигурация (`config.local.yaml`)

Расширяем конфигурацию — добавляем секцию `grpc`:

```yaml
grpc:
  playerService:
    target: "localhost:9090"
    insecure: true

grpc_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
```

### 1. Описание Моделей

Модели генерируются из `.proto` файлов:

```protobuf
service PlayerService {
    rpc GetPlayer (GetPlayerRequest) returns (GetPlayerResponse);
}

message GetPlayerRequest {
    string id = 1;
}

message GetPlayerResponse {
    string id = 1;
    string username = 2;
    string status = 3;
    string region = 4;
}
```

### 2. Реализация Клиента

**Файл:** `internal/grpc_client/player/client.go`

```go
package player

import (
    "github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
    "github.com/gorelov-m-v/go-test-framework/pkg/grpc/dsl"
    "github.com/ozontech/allure-go/pkg/framework/provider"
    pb "my-project/internal/models/grpc/player"
)

var grpcClient *client.Client

type Link struct{}

func (l *Link) SetGRPC(c *client.Client) {
    grpcClient = c
}

func GetPlayer(sCtx provider.StepCtx) *dsl.Call[pb.GetPlayerRequest, pb.GetPlayerResponse] {
    return dsl.NewCall[pb.GetPlayerRequest, pb.GetPlayerResponse](sCtx, grpcClient).
        Method("/player.PlayerService/GetPlayer")
}
```

### 3. Подключение в Env

Расширяем `tests/env.go` — добавляем gRPC:

```go
type TestEnv struct {
    // HTTP
    GameService game.Link `config:"gameService"`
    // Database
    PlayersRepo players.Link `db_config:"coreDatabase"`
    // Kafka
    Kafka kafka.Link `kafka_config:"kafka"`
    // Redis
    PlayerCache player.Link `redis_config:"redis.cache"`
    // gRPC - связываем конфиг "grpc.playerService" с пакетом "grpcPlayer"
    PlayerGRPC grpcPlayer.Link `grpc_config:"grpc.playerService"`
}
```

### 4. Тест

Добавляем **Шаг 3** — финальную верификацию через gRPC.

```go
func (s *PlayerSuite) TestCreatePlayerFullE2E(t provider.T) {
    // ... Шаг 1: HTTP (см. выше) ...
    // ... Шаг 2.1: Database (см. выше) ...
    // ... Шаг 2.2: Kafka (см. выше) ...
    // ... Шаг 2.3: Redis (см. выше) ...

    // ═══════════════════════════════════════════════════════════════
    // ШАГ 3: gRPC - Финальная верификация (ждёт завершения всех AsyncStep)
    // ═══════════════════════════════════════════════════════════════
    s.Step(t, "gRPC: Верификация игрока", func(sCtx provider.StepCtx) {
        grpcPlayer.GetPlayer(sCtx).
            RequestBody(pb.GetPlayerRequest{
                Id: httpResp.Body.ID,
            }).
            ExpectNoError().
            ExpectStatusCode(codes.OK).
            ExpectFieldValue("id", httpResp.Body.ID).
            ExpectFieldValue("username", username).
            ExpectFieldValue("status", "active").
            ExpectFieldValue("region", "EU").
            Send()
    })
}
```

---

### Справочник методов gRPC DSL

Объект `dsl.Call[TReq, TResp]` предоставляет методы для настройки gRPC запроса.

### 1. Настройка запроса

| Метод | Описание | Пример |
|:---|:---|:---|
| `.Method(fullMethod)` | Полный путь метода | `.Method("/player.PlayerService/GetPlayer")` |
| `.RequestBody(req)` | Тело запроса | `.RequestBody(pb.GetRequest{Id: "123"})` |
| `.Metadata(k, v)` | Добавить metadata | `.Metadata("authorization", "Bearer ...")` |

### 2. Ожидания (Expectations)

**Проверки статуса:**
*   `.ExpectNoError()` — Успешный вызов.
*   `.ExpectError()` — Ожидает ошибку.
*   `.ExpectStatusCode(codes.OK)` — Конкретный gRPC status code.

**Проверки полей (GJSON Path):**
*   `.ExpectFieldValue("path", value)` — Значение поля.
*   `.ExpectFieldNotEmpty("path")` — Непустое поле.
*   `.ExpectMetadata("key", "value")` — Metadata в ответе.

### 3. Выполнение

*   `.Send()` — Выполняет вызов, возвращает `*client.Response[TResp]`.

---

## Полный E2E тест

Соберём все шаги вместе. Это **полноценный E2E сценарий через 5 DSL**:

```go
import (
    "time"
    "my-project/internal/http_client/game"
    "my-project/internal/db/players"
    "my-project/internal/kafka"
    "my-project/internal/cache/player"
    grpcPlayer "my-project/internal/grpc_client/player"
    kafkaDSL "go-test-framework/pkg/kafka/dsl"
    httpClient "go-test-framework/pkg/http/client"
    pb "my-project/internal/models/grpc/player"
    "google.golang.org/grpc/codes"
)

func (s *PlayerSuite) TestCreatePlayerFullE2E(t provider.T) {
    t.Title("Full E2E: HTTP → [DB + Kafka + Redis] → gRPC")

    var (
        httpResp *httpClient.Response[models.CreatePlayerResp]
        username = "pro_gamer_2024"
    )

    // ═══════════════════════════════════════════════════════════════
    // ШАГ 1: HTTP - Создаём игрока через API
    // ═══════════════════════════════════════════════════════════════
    s.Step(t, "HTTP: Создание игрока", func(sCtx provider.StepCtx) {
        httpResp = game.CreatePlayer(sCtx).
            RequestBody(models.CreatePlayerReq{
                Username: username,
                Region:   "EU",
            }).
            ExpectResponseStatus(201).
            ExpectResponseBodyFieldNotEmpty("id").
            ExpectResponseBodyFieldValue("status", "active").
            Send()
    })

    // ═══════════════════════════════════════════════════════════════
    // ШАГ 2: Параллельные проверки (запускаются ОДНОВРЕМЕННО)
    // ═══════════════════════════════════════════════════════════════

    // 2.1: Database
    s.AsyncStep(t, "Database: Проверка записи", func(sCtx provider.StepCtx) {
        players.FindByID(sCtx, httpResp.Body.ID).
            ExpectFound().
            ExpectColumnEquals("username", username).
            ExpectColumnEquals("status", "active").
            ExpectColumnIsNotNull("created_at").
            Send()
    })

    // 2.2: Kafka
    s.AsyncStep(t, "Kafka: Ожидание события", func(sCtx provider.StepCtx) {
        kafkaDSL.Expect[kafka.PlayerEventsTopic](sCtx, kafka.Client()).
            With("playerId", httpResp.Body.ID).
            With("eventType", "PLAYER_CREATED").
            Unique().
            ExpectField("playerName", username).
            Send()
    })

    // 2.3: Redis
    s.AsyncStep(t, "Redis: Проверка кэша", func(sCtx provider.StepCtx) {
        playerCache.GetByID(sCtx, httpResp.Body.ID).
            ExpectExists().
            ExpectJSONField("username", username).
            ExpectTTL(4*time.Minute, 6*time.Minute).
            Send()
    })

    // ═══════════════════════════════════════════════════════════════
    // ШАГ 3: gRPC - Финальная верификация (ждёт завершения AsyncStep)
    // ═══════════════════════════════════════════════════════════════
    s.Step(t, "gRPC: Верификация игрока", func(sCtx provider.StepCtx) {
        grpcPlayer.GetPlayer(sCtx).
            RequestBody(pb.GetPlayerRequest{
                Id: httpResp.Body.ID,
            }).
            ExpectNoError().
            ExpectStatusCode(codes.OK).
            ExpectFieldValue("username", username).
            ExpectFieldValue("status", "active").
            Send()
    })
}
```

**Результат:** Один тест проверил **5 слоёв системы** — HTTP API, базу данных, Kafka, Redis и gRPC.

---

### Кодогенерация

Написание HTTP/gRPC клиентов, моделей запросов/ответов и DSL методов вручную — это рутина, которая **замедляет старт автоматизации**. Пока QA пишет boilerplate код, разработка уходит вперёд, и тесты всегда "догоняют".

**Решение:** Генераторы превращают спецификации (OpenAPI/Proto) в готовые клиенты **за минуты**:

| Генератор | Вход | Что генерирует |
|-----------|------|----------------|
| `openapi-gen` | OpenAPI 3.x (JSON/YAML) | HTTP клиенты + модели |
| `grpc-gen` | `.proto` файлы | gRPC DSL клиенты |

**Преимущества:**
- ⏱️ **Сокращение времени** с дней до часов
- ⬅️ **Shift-left testing** — для старта нужен только контракт, а не готовая реализация
- 🔄 **Синхронизация с API** — одна команда обновляет все клиенты
- ✅ **Чистый код** — без ручных ошибок и несоответствий

##### Маркировка сгенерированных файлов

Все сгенерированные файлы **автоматически** содержат маркер в первой строке:

```go
// Code generated by openapi-gen. DO NOT EDIT.

package auth
```

или для gRPC:

```go
// Code generated by grpc-gen. DO NOT EDIT.

package playerservice
```

**Зачем нужна маркировка:**
- 🚫 Предотвращает случайное редактирование сгенерированного кода
- 🔧 IDE и линтеры распознают файл как автогенерированный
- 📋 Стандартная практика Go (`go generate`)

> ⚠️ **ВАЖНО:** Никогда не редактируйте файлы с маркировкой `DO NOT EDIT` напрямую. Если нужны изменения — модифицируйте спецификацию и перегенерируйте код.

---

#### OpenAPI Generator (`openapi-gen`)

Автоматический генератор Go клиентов и моделей из OpenAPI 3.x спецификации.

##### Установка

```bash
go install github.com/gorelov-m-v/go-test-framework/cmd/openapi-gen@latest
```

##### Использование

```bash
# Базовое использование
openapi-gen openapi.json

# Только конкретный сервис
openapi-gen -service auth openapi.json

# Кастомные пути
openapi-gen -client pkg/http_client/auth openapi.json
```

##### Результат генерации

```
internal/
└── http_client/
    ├── auth/
    │   ├── client.go       # ✨ Link + DSL методы
    │   └── models.go       # ✨ Request/Response модели
    └── users/
        ├── client.go       # ✨ Link + DSL методы
        └── models.go       # ✨ Request/Response модели
```

##### Флаги командной строки

```bash
openapi-gen [options] <openapi-spec>

Флаги:
  -service string    Имя сервиса (default: генерирует все)
  -output string     Директория вывода (default: .)
  -client string     Путь для клиента (default: internal/http_client/{service})
```

##### Примеры возможностей

**1. Автоматическое разделение на сервисы по тегам:**

```yaml
# OpenAPI спецификация
paths:
  /auth/login:
    post:
      tags: ["Auth"]        # → internal/http_client/auth/
  /users/{id}:
    get:
      tags: ["Users"]       # → internal/http_client/users/
```

**2. Чистые имена методов:**

```
# Было в OpenAPI:
operationId: auth_jwt_login_auth_jwt_login_post

# Стало в Go:
auth.Login()   // POST /auth/login
```

**3. CRUD с path параметрами:**

```yaml
# OpenAPI:
GET    /users           # → Users()
GET    /users/{id}      # → GetUsers(id)
PUT    /users/{id}      # → UpdateUsers(id)
DELETE /users/{id}      # → DeleteUsers(id)
```

**4. Интеграция в тесты:**

```go
// tests/env.go
type TestEnv struct {
    Auth  auth.Link  `config:"authService"`   // Сгенерированный клиент
    Users users.Link `config:"usersService"`
}

// tests/auth_test.go
func (s *AuthSuite) TestLogin(t provider.T) {
    s.Step(t, "Login user", func(sCtx provider.StepCtx) {
        auth.Login(sCtx).
            RequestBody(auth.LoginRequest{
                Email:    "test@test.com",
                Password: "P@ssw0rd",
            }).
            ExpectResponseStatus(200).
            Send()
    })
}
```

---

#### gRPC Generator (`grpc-gen`)

Автоматический генератор DSL клиентов из `.proto` файлов.

##### Установка

```bash
go install github.com/gorelov-m-v/go-test-framework/cmd/grpc-gen@latest
```

##### Использование

```bash
# Базовое использование (pb-import обязателен!)
grpc-gen -pb-import "your-project/pkg/pb/player" player.proto

# Конкретный сервис
grpc-gen -service PlayerService -pb-import "your-project/pkg/pb" player.proto

# Кастомные пути
grpc-gen -pb-import "your-project/pb" -client internal/grpc_client/player player.proto
```

##### Результат генерации

```
internal/
└── grpc_client/
    └── playerservice/
        └── client.go   # ✨ Link + DSL методы
```

##### Флаги командной строки

```bash
grpc-gen [options] <proto-file>

Флаги:
  -service string    Имя сервиса (default: все сервисы в файле)
  -output string     Директория вывода (default: .)
  -client string     Путь для клиента (default: internal/grpc_client/{service})
  -pb-import string  Import path для protobuf типов (обязательный!)
  -module string     Go module name (default: auto-detect из go.mod)
```

##### Пример генерации

**Входной `.proto` файл:**

```protobuf
syntax = "proto3";

package player;
option go_package = "your-project/pkg/pb/player";

service PlayerService {
    rpc CreatePlayer(CreatePlayerRequest) returns (CreatePlayerResponse);
    rpc GetPlayer(GetPlayerRequest) returns (GetPlayerResponse);
    rpc UpdatePlayer(UpdatePlayerRequest) returns (UpdatePlayerResponse);
}

message CreatePlayerRequest {
    string username = 1;
    string region = 2;
}

message CreatePlayerResponse {
    string id = 1;
    string status = 2;
}
// ... остальные сообщения
```

**Сгенерированный код:**

```go
package playerservice

import (
    "github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
    "github.com/gorelov-m-v/go-test-framework/pkg/grpc/dsl"
    "github.com/ozontech/allure-go/pkg/framework/provider"

    pb "your-project/pkg/pb/player"
)

var grpcClient *client.Client

type Link struct{}

func (l *Link) SetGRPC(c *client.Client) {
    grpcClient = c
}

// CreatePlayer calls /player.PlayerService/CreatePlayer
func CreatePlayer(sCtx provider.StepCtx) *dsl.Call[pb.CreatePlayerRequest, pb.CreatePlayerResponse] {
    return dsl.NewCall[pb.CreatePlayerRequest, pb.CreatePlayerResponse](sCtx, grpcClient).
        Method("/player.PlayerService/CreatePlayer")
}

// GetPlayer calls /player.PlayerService/GetPlayer
func GetPlayer(sCtx provider.StepCtx) *dsl.Call[pb.GetPlayerRequest, pb.GetPlayerResponse] {
    return dsl.NewCall[pb.GetPlayerRequest, pb.GetPlayerResponse](sCtx, grpcClient).
        Method("/player.PlayerService/GetPlayer")
}

// UpdatePlayer calls /player.PlayerService/UpdatePlayer
func UpdatePlayer(sCtx provider.StepCtx) *dsl.Call[pb.UpdatePlayerRequest, pb.UpdatePlayerResponse] {
    return dsl.NewCall[pb.UpdatePlayerRequest, pb.UpdatePlayerResponse](sCtx, grpcClient).
        Method("/player.PlayerService/UpdatePlayer")
}
```

##### Интеграция в тесты

```go
// tests/env.go
type TestEnv struct {
    PlayerGRPC playerservice.Link `grpc_config:"playerService"`
}

// tests/player_grpc_test.go
func (s *PlayerSuite) TestCreatePlayerGRPC(t provider.T) {
    s.Step(t, "Create player via gRPC", func(sCtx provider.StepCtx) {
        playerservice.CreatePlayer(sCtx).
            RequestBody(pb.CreatePlayerRequest{
                Username: "test_user",
                Region:   "EU",
            }).
            ExpectNoError().
            ExpectStatusCode(codes.OK).
            ExpectFieldValue("status", "active").
            Send()
    })
}
```

---

#### Автоматизация генерации

##### Вариант 1: go:generate

```go
// internal/generate.go
package internal

//go:generate openapi-gen ../../openapi.json
//go:generate grpc-gen -pb-import "your-project/pkg/pb/player" ../../proto/player.proto
```

Запуск:
```bash
go generate ./...
```

##### Вариант 2: Makefile

```makefile
.PHONY: generate
generate:
	# Сначала protoc для генерации pb типов
	protoc --go_out=. --go-grpc_out=. proto/*.proto

	# Затем генераторы DSL клиентов
	openapi-gen openapi.json
	grpc-gen -pb-import "your-project/pkg/pb/player" proto/player.proto

	go fmt ./...
```

##### Рекомендации

- ✅ Коммитьте `openapi.json` и `.proto` файлы в репозиторий
- ✅ Регенерируйте клиенты при обновлении спецификаций
- ✅ Для gRPC: сначала `protoc`, затем `grpc-gen`
- ⚠️ Не редактируйте сгенерированные файлы вручную

---

## Расширенные возможности

### Асинхронные шаги (AsyncStep)

#### Зачем это нужно

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

### Параметризованные тесты (Table-Driven Tests)

Параметризованные тесты позволяют запустить один и тот же тест с разными наборами данных. Это особенно полезно для негативных тестов, когда нужно проверить множество граничных случаев и валидаций.

#### Зачем это нужно

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

## Полный пример параметризованного теста с RequestBodyMap

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

### Маскировка чувствительных данных

Тесты часто работают с конфиденциальной информацией: токенами авторизации, паролями, API ключами. Эти данные попадают в Allure отчёты, которые могут быть доступны широкому кругу лиц. Фреймворк предоставляет механизм **настраиваемой маскировки** чувствительных данных в HTTP запросах, SQL запросах и результатах из БД.

---

#### Зачем это нужно

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

### Контрактное тестирование (Contract Testing)

Модуль `pkg/contract` позволяет валидировать HTTP-ответы против OpenAPI спецификации. Это гарантирует, что API соответствует задокументированному контракту.

---

#### Зачем нужно контрактное тестирование

**Проблема:** API возвращает ответы, не соответствующие спецификации:
- Отсутствуют обязательные поля
- Неверные типы данных (string вместо integer)
- Лишние поля, не описанные в контракте

**Решение:** Автоматическая валидация каждого ответа против OpenAPI схемы.

---

#### Конфигурация контрактного тестирования

```yaml
http:
  gameService:
    baseURL: "https://api.example.com"
    timeout: 30s
    contractSpec: "openapi/openapi.json"      # Путь к OpenAPI спецификации
    contractBasePath: "/api"                   # Опционально: префикс пути в спецификации
```

| Параметр | Описание |
|----------|----------|
| `contractSpec` | Путь к файлу OpenAPI спецификации (JSON или YAML) |
| `contractBasePath` | Префикс пути, если DSL-пути не совпадают с путями в спецификации |

---

#### Использование контрактного тестирования

##### Валидация по операции (метод + путь)

```go
s.Step(t, "Create player with contract validation", func(sCtx provider.StepCtx) {
    resp := players.Create(sCtx).
        RequestBody(models.CreateRequest{Username: "test"}).
        ExpectResponseStatus(http.StatusCreated).
        ExpectMatchesContract().  // Валидация против OpenAPI
        Send()
})
```

Метод `ExpectMatchesContract()` автоматически:
1. Находит операцию в спецификации по HTTP методу и пути
2. Извлекает схему ответа для полученного статус-кода
3. Валидирует JSON-ответ против схемы

##### Валидация по имени схемы

```go
s.Step(t, "Get player with schema validation", func(sCtx provider.StepCtx) {
    resp := players.GetByID(sCtx, playerID).
        ExpectResponseStatus(http.StatusOK).
        ExpectMatchesSchema("PlayerResponse").  // Валидация против конкретной схемы
        Send()
})
```

Используйте `ExpectMatchesSchema(name)` когда:
- Нужно проверить против конкретной схемы из `components/schemas`
- Путь не описан в спецификации, но схема есть

##### Ошибки валидации

При несоответствии контракту тест падает с детальным сообщением:

```
Contract validation failed for POST /api/v1/players (status 201):
- /username: expected string, got integer
- /status: missing required field
- /extra_field: additional property not allowed
```

---

### Генерация тестовых данных

Тесты часто требуют уникальных данных для каждого запуска: email адреса, пароли, случайные строки. Модуль `pkg/datagen` предоставляет простой API для генерации тестовых данных с гарантией соответствия требованиям валидации.

---

#### Зачем это нужно

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

### BaseSuite и Cleanup

`BaseSuite` — базовая структура для тестовых suite, которая предоставляет удобные методы для организации тестов и автоматическое управление ресурсами.

#### Структура BaseSuite

```go
import "github.com/gorelov-m-v/go-test-framework/pkg/extension"

type MySuite struct {
    extension.BaseSuite
}

func TestMySuite(t *testing.T) {
    suite.RunSuite(t, new(MySuite))
}
```

`BaseSuite` предоставляет:
- `Step` — синхронный шаг теста
- `AsyncStep` — асинхронный шаг с автоматическими retry
- `Cleanup` — регистрация функции очистки

#### Методы Step и AsyncStep

```go
func (s *MySuite) TestExample(t provider.T) {
    // Синхронный шаг — выполняется последовательно
    s.Step(t, "Create resource", func(sCtx provider.StepCtx) {
        // ...
    })

    // Асинхронные шаги — выполняются параллельно друг с другом
    s.AsyncStep(t, "Verify in DB", func(sCtx provider.StepCtx) {
        // автоматические retry до успеха или таймаута
    })

    s.AsyncStep(t, "Verify in Kafka", func(sCtx provider.StepCtx) {
        // выполняется параллельно с предыдущим AsyncStep
    })

    // Синхронный шаг ЖДЁТ завершения всех предыдущих AsyncStep
    s.Step(t, "Final check", func(sCtx provider.StepCtx) {
        // все async шаги уже завершены
    })
}
```

#### Автоматический Cleanup

Метод `Cleanup` регистрирует функцию очистки, которая **гарантированно выполнится** в `AfterEach`, даже если тест упадёт.

```go
func (s *MySuite) TestCreateAndDelete(t provider.T) {
    var resourceID string

    s.Step(t, "Create resource", func(sCtx provider.StepCtx) {
        resp := api.CreateResource(sCtx).Send()
        resourceID = resp.Body.ID
    })

    // Регистрируем cleanup — выполнится в AfterEach
    s.Cleanup(func(t provider.T) {
        s.Step(t, "Delete resource", func(sCtx provider.StepCtx) {
            api.DeleteResource(sCtx, resourceID).Send()
        })
    })

    // Проверки — если упадут, cleanup всё равно выполнится
    s.Step(t, "Verify resource", func(sCtx provider.StepCtx) {
        // ...
    })
}
```

**Преимущества:**
- Не нужно переопределять `AfterEach`
- Не нужно вызывать родительский метод
- Cleanup выполняется автоматически после всех async шагов
- Ресурсы удаляются даже при падении теста

---

### Настройка Allure

По умолчанию allure-go создаёт папку `allure-results` в директории каждого тестового пакета. Чтобы собирать все результаты в одном месте, укажите путь в конфиге или переменной окружения.

#### Вариант 1: В конфиге (рекомендуется)

```yaml
# configs/config.local.yaml
allure:
  outputPath: "allure-results"
```

Фреймворк автоматически настроит `ALLURE_OUTPUT_PATH` относительно корня проекта. Результаты будут в `<project_root>/allure-results/`.

#### Вариант 2: При запуске тестов

```bash
ALLURE_OUTPUT_PATH=. go test ./...
```

> **Примечание:** allure-go автоматически создаёт папку `allure-results` внутри указанного пути. Поэтому `ALLURE_OUTPUT_PATH=.` создаст `./allure-results/`.

#### Генерация отчёта

```bash
# Запустить локальный сервер с отчётом
allure serve ./allure-results

# Или сгенерировать статический отчёт
allure generate ./allure-results -o ./allure-report
```

---

## Быстрый старт

### Вариант 1: Новый проект (рекомендуется)

Создайте тестовый проект одной командой:

```bash
# 1. Установите генератор проекта
go install github.com/gorelov-m-v/go-test-framework/cmd/test-init@latest

# 2. Создайте проект
test-init my-api-tests

# 3. Готово!
cd my-api-tests
go mod tidy
```

**Что получаете:**

```
my-api-tests/
├── configs/config.local.yaml     # Готовый конфиг
├── internal/
│   ├── http_client/example/      # Пример HTTP клиента
│   ├── grpc_client/              # Для gRPC клиентов
│   ├── db/                       # Для DB репозиториев
│   ├── redis/                    # Для Redis
│   └── kafka/                    # Для Kafka
├── tests/
│   ├── env.go                    # TestEnv с DI
│   └── example_test.go           # Рабочий пример теста
├── Makefile                      # Все нужные команды
└── go.mod
```

**Основные команды:**

```bash
make help                         # Показать все команды
make test                         # Запустить тесты
make test-run TEST=TestName       # Запустить конкретный тест
make allure                       # Открыть Allure отчёт
make install-tools                # Установить генераторы
make gen-http SPEC=openapi.json   # Сгенерировать HTTP клиенты
```

---

### Вариант 2: Демо-проект (для изучения)

Демо-проект показывает все возможности фреймворка на реальном примере — game-service с полным стеком: HTTP API, gRPC, PostgreSQL, Redis, Kafka.

#### Что входит в демо

| Репозиторий | Описание |
|-------------|----------|
| [game-service](https://github.com/gorelov-m-v/game-service) | Микросервис на Go: HTTP + gRPC + PostgreSQL + Redis + Kafka |
| [game-service-tests](https://github.com/gorelov-m-v/game-service-tests) | E2E тесты: 11 test suites, 50+ тестов |

#### Запуск демо

```bash
# 1. Скачайте сервис и тесты
git clone https://github.com/gorelov-m-v/game-service.git
git clone https://github.com/gorelov-m-v/game-service-tests.git

# 2. Запустите инфраструктуру
cd game-service
docker-compose up -d

# Поднимается:
# - PostgreSQL (5432)
# - Redis (6379)
# - Kafka + Zookeeper (9092)
# - game-service HTTP (8080) + gRPC (9090)

# 3. Подождите ~30 сек и запустите тесты
cd ../game-service-tests
go mod tidy
go test ./... -v

# 4. Посмотрите Allure отчёт
allure serve allure-results

# 5. После работы остановите инфраструктуру
cd ../game-service && docker-compose down
```

---

### Требования

| Инструмент | Проверка | Установка |
|------------|----------|-----------|
| **Go 1.21+** | `go version` | [golang.org/dl](https://golang.org/dl/) |
| **Allure** | `allure --version` | [docs.qameta.io](https://docs.qameta.io/allure/#_installing_a_commandline) |
| **Docker** (для демо) | `docker --version` | [docker.com](https://www.docker.com/products/docker-desktop) |

---

### Обзор демо-тестов

В демо-проекте **11 test suites**, покрывающих 100% функционала фреймворка:

#### HTTP DSL

| Test Suite | Что демонстрирует |
|------------|-------------------|
| **SimpleSuite** | Базовый E2E: HTTP → Kafka |
| **PlayerSuite** | Полный E2E: HTTP → DB + Kafka (параллельно) |
| **PlayerNegativeSuite** | Table-Driven тесты, `RequestBodyMap` |
| **PlayerGetSuite** | `GET` + `PathParam` + GJSON пути |
| **PlayerListSuite** | `QueryParam`, пагинация, массивы |
| **PlayerUpdateSuite** | `PATCH` + AsyncStep с HTTP retry |
| **PlayerDeleteSuite** | `DELETE` + `ExpectNotFound` |

#### Database DSL

| Test Suite | Что демонстрирует |
|------------|-------------------|
| **DatabaseFeaturesSuite** | `ExpectFound`, `ExpectNotFound`, `ExpectColumnEquals`, `ExpectColumnNotEquals`, `ExpectColumnTrue/False`, `ExpectColumnIsNull/IsNotNull`, `ExpectColumnEmpty`, `ExpectColumnJsonEquals` |

#### Kafka DSL

| Test Suite | Что демонстрирует |
|------------|-------------------|
| **KafkaFeaturesSuite** | `With` фильтры, `Unique`, `UniqueWithWindow`, `ExpectField`, `ExpectFieldTrue/False`, `ExpectFieldIsNull/IsNotNull` |

#### gRPC DSL

| Test Suite | Что демонстрирует |
|------------|-------------------|
| **GRPCFeaturesSuite** | CRUD через gRPC, `ExpectNoError`, `ExpectError`, `ExpectStatusCode`, `ExpectFieldValue`, `ExpectFieldNotEmpty`, AsyncStep retry |

#### Redis DSL

| Test Suite | Что демонстрирует |
|------------|-------------------|
| **RedisFeaturesSuite** | `ExpectExists`, `ExpectNotExists`, `ExpectJSONField`, `ExpectJSONFieldNotEmpty`, cache invalidation, AsyncStep retry |

---

#### Рекомендация по изучению

1. **Начните с `SimpleSuite`** — минимальный E2E: HTTP → Kafka
2. **Изучите `PlayerSuite`** — полный паттерн: HTTP → DB + Kafka (параллельно)
3. **Посмотрите `GRPCFeaturesSuite`** и **`RedisFeaturesSuite`** — gRPC/Redis DSL в действии
4. **Изучите `PlayerNegativeSuite`** — Table-Driven параметризованные тесты

---

## Рекомендуемая структура проекта

Фреймворк не навязывает жесткую структуру, но следование этой организации файлов упрощает поддержку и масштабирование тестов:

```
your-api-tests/
├── configs/
│   └── config.local.yaml         # Конфигурация (http, db, kafka, redis, grpc)
│
├── internal/
│   ├── http_client/              # HTTP клиенты + модели
│   │   └── [service_name]/
│   │       ├── client.go         # Link + DSL методы
│   │       └── models.go         # Request/Response
│   │
│   ├── grpc_client/              # gRPC клиенты + модели
│   │   └── [service_name]/
│   │       ├── client.go         # Link + DSL методы
│   │       └── *.pb.go           # Protobuf (сгенерировано)
│   │
│   ├── db/                       # Database репозитории
│   │   └── [table_name]/
│   │       ├── repo.go           # Link + DSL методы
│   │       └── models.go         # Модель таблицы
│   │
│   ├── redis/                    # Redis кэш
│   │   └── [entity_name]/
│   │       ├── cache.go          # Link + DSL методы
│   │       └── models.go         # Структура данных (если нужно)
│   │
│   └── kafka/                    # Kafka топики
│       └── [topic_name]/
│           ├── topic.go          # Определение топика + Link
│           └── models.go         # Модели сообщений
│
├── proto/                        # Исходные .proto файлы (опционально)
│   └── [service_name].proto
│
├── openapi.json                  # OpenAPI спецификация (или openapi/)
│
└── tests/
    ├── env.go                    # TestEnv с DI тегами
    ├── *_test.go                 # Тесты
    └── allure-results/           # Сгенерированные Allure отчёты
```

## Где создавать файлы

| Что | Где |
|-----|-----|
| HTTP клиент "auth" | `internal/http_client/auth/client.go` |
| HTTP модели "auth" | `internal/http_client/auth/models.go` |
| gRPC клиент "player" | `internal/grpc_client/player/client.go` |
| gRPC модели "player" | `internal/grpc_client/player/*.pb.go` |
| DB репозиторий "users" | `internal/db/users/repo.go` |
| DB модель "users" | `internal/db/users/models.go` |
| Redis "session" | `internal/redis/session/cache.go` |
| Redis модель (если есть) | `internal/redis/session/models.go` |
| Kafka топик "player_events" | `internal/kafka/player_events/topic.go` |
| Kafka модели | `internal/kafka/player_events/models.go` |
| Тесты | `tests/*_test.go` |
| TestEnv | `tests/env.go` |
| Конфигурация | `configs/config.{ENV}.yaml` |
| Proto файлы | `proto/*.proto` |

## Ключевые принципы организации

### 1. Модели рядом с клиентами
Всё связанное в одной папке — удалил сервис, удалил папку:
```
internal/http_client/auth/
├── client.go    # Link + Login(), Register(), Logout()
└── models.go    # LoginRequest, LoginResponse, ...
```

### 2. Группировка по доменам
Каждый сервис/таблица/топик имеет свою папку:
```
internal/
├── http_client/
│   ├── auth/         # Auth API
│   ├── game/         # Game API
│   └── payment/      # Payment API
├── grpc_client/
│   └── player/       # Player gRPC
├── db/
│   ├── users/        # users table
│   └── orders/       # orders table
├── redis/
│   └── session/      # session cache
└── kafka/
    └── player_events/  # player.events topic
```

### 3. Единая точка входа
`tests/env.go` содержит все зависимости:
```go
type TestEnv struct {
    // HTTP клиенты
    Auth    auth.Link    `config:"authService"`
    Game    game.Link    `config:"gameService"`

    // gRPC клиенты
    PlayerGRPC player.Link `grpc_config:"playerService"`

    // Database репозитории
    Users   users.Link   `db_config:"coreDatabase"`

    // Redis
    Session session.Link `redis_config:"redis"`

    // Kafka
    PlayerEvents player_events.Link `kafka_config:"kafka"`
}
```

## Преимущества такой структуры

- **Колокация:** Модели рядом с клиентами — всё связанное вместе
- **Предсказуемость:** Любой разработчик знает, где искать файлы
- **Масштабируемость:** Легко добавлять новые сервисы
- **Чистое удаление:** Удалил папку — удалил всё связанное с сервисом
- **Генерация:** `openapi-gen` и `grpc-gen` кладут файлы в правильные места

---

