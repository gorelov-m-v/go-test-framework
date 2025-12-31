# Go Test Framework

Мощный и типобезопасный фреймворк для написания автотестов на Go. Спроектирован для построения сложных интеграционных тестов с глубокой интеграцией в **Allure Report**.

## 🚀 Быстрый старт

### Установка

```bash
git clone <repository-url>
cd go-test-framework
go mod download
```

### Запуск тестов фреймворка

```bash
# Запустить все unit-тесты
go test ./pkg/... -v

# Только database DSL
go test ./pkg/database/dsl -v

# С покрытием
go test ./pkg/... -cover
```

## 🚀 Особенности

*   **Fluent DSL:**
    Декларативный стиль написания тестов через цепочки вызовов. Вы описываете *что* отправить и *чего* ожидать.
    ```go
    service.CreateUser(ctx).
        RequestBody(user).
        ExpectResponseStatus(http.StatusCreated).
        ExpectResponseBodyFieldNotEmpty("id").
        RequestSend()
    ```

*   **Type-Safe:**
    Использование Go Generics гарантирует, что вы отправляете в `RequestBody` правильную структуру и получаете типизированный ответ. Ошибки несоответствия типов ловятся на этапе компиляции.

*   **Allure Native:**
    Каждый запрос, проверка и действие автоматически оборачиваются в шаги (Steps) отчета.

*   **Auto-Logging:**
    Полные данные HTTP-запроса и ответа (JSON, заголовки, cURL) автоматически форматируются и прикрепляются к отчету.

## 📦 Установка

```bash
go get github.com/your-org/go-test-framework
```

## 🧩 Модули

Фреймворк состоит из независимых модулей для разных протоколов:

*   [**Конфигурация**](#-конфигурация)
    *   [Структура Config.yaml](#структура-configyaml)
    *   [Автоматическая инициализация окружения](#автоматическая-инициализация-окружения)
    *   [Загрузка кастомных данных](#загрузка-кастомных-данных)
*   [**HTTP Клиент**](#-http-клиент)
    *   [1. Описание Моделей](#1-описание-моделей)
    *   [2. Создание и Конфигурация Клиента](#2-создание-и-конфигурация-клиента)
    *   [3. Написание Теста](#3-написание-теста)
    *   [3.1. Получение данных из ответа](#31-получение-данных-из-ответа)
    *   [4. Справочник HTTP DSL](#4-справочник-http-dsl)
    *   [5. Пример отчета Allure](#5-пример-отчета-allure)
*   [**Database DSL (MySQL / PostgreSQL)**](#️-mysql-database-dsl)
    *   [1. Конфигурация БД](#1-конфигурация-бд)
    *   [2. Описание моделей с db тегами](#2-описание-моделей-с-db-тегами)
    *   [3. Создание Repository Pattern](#3-создание-repository-pattern)
    *   [4. Написание E2E тестов](#4-написание-e2e-тестов)
    *   [5. Справочник DB DSL](#5-справочник-db-dsl)
*   NATS Клиент
*   Kafka Клиент
*   Redis Клиент

---

# ⚙️ Конфигурация

Модуль `pkg/config` предоставляет мощный механизм управления конфигурацией для ваших тестов. Он автоматически загружает настройки из YAML-файлов и использует DI-подобную технику для автоматической инициализации HTTP-клиентов.

## Структура Config.yaml

Создайте файл `config.yaml` в корне проекта или в директории `./configs/`:

```yaml
# Конфигурация сервисов
capService:
  baseURL: https://cap.beta-09.b2bdev.pro
  timeout: 30s
  defaultHeaders:
    Accept: application/json
    Content-Type: application/json

# Тестовые данные
testData:
  defaultUsername: admin
  defaultPassword: admin
  validEmails:
    - user1@example.com
    - user2@example.com
```

## Автоматическая инициализация окружения

### Шаг 1: Подготовьте клиент-обертку

Ваш клиент **обязан** иметь публичное поле `HTTP *client.Client`:

```go
package client

import (
	"go-test-framework/pkg/http/client"
	"go-test-framework/pkg/http/dsl"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type CapClient struct {
	HTTP *client.Client  // ⚠️ Обязательное поле для автоинъекции
}

func (c *CapClient) TokenCheck(sCtx provider.StepCtx) *dsl.Call[TokenCheckRequest, TokenCheckResponse] {
	return dsl.NewCall[TokenCheckRequest, TokenCheckResponse](sCtx, c.HTTP).
		POST("/_cap/api/token/check")
}
```

### Шаг 2: Создайте структуру TestEnv

Определите структуру окружения с тегами `config:"ключ_в_yaml"`:

```go
package tests

import (
	"your-project/internal/client"
)

type TestEnv struct {
	CapClient *client.CapClient `config:"capService"`
	// Добавьте другие клиенты по необходимости
}

var env *TestEnv
```

### Шаг 3: Инициализируйте окружение в TestMain

```go
package tests

import (
	"log"
	"os"
	"testing"

	"go-test-framework/pkg/config"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

func TestMain(m *testing.M) {
	// Автоматическая инициализация всех клиентов из config.yaml
	env = &TestEnv{}
	if err := config.BuildEnv(env); err != nil {
		log.Fatalf("Failed to build test environment: %v", err)
	}

	// Запуск тестов
	suite.RunTests(m)
	os.Exit(0)
}
```

### Шаг 4: Загрузите тестовые данные из конфига

```go
type TestData struct {
	DefaultUsername string `mapstructure:"defaultUsername"`
	DefaultPassword string `mapstructure:"defaultPassword"`
}

var testData TestData

func init() {
	// ... инициализация env ...

	// Загрузка тестовых данных
	if err := config.UnmarshalByKey("testData", &testData); err != nil {
		log.Fatalf("Failed to load test data: %v", err)
	}
}
```

### Шаг 5: Используйте в тестах

```go
func (s *CAPTokenSuite) TestTokenCheck(t provider.T) {
	t.Title("CAP API: Token check")

	t.WithNewStep("Token check request", func(sCtx provider.StepCtx) {
		// Используем клиент и данные из конфига
		env.CapClient.TokenCheck(sCtx).
			RequestBody(models.TokenCheckRequest{
				Username: testData.DefaultUsername,
				Password: testData.DefaultPassword,
			}).
			ExpectResponseStatus(http.StatusOK).
			ExpectResponseBodyFieldNotEmpty("token").
			ExpectResponseBodyFieldValue("success", true).
			RequestSend()
	})
}
```

## Загрузка кастомных данных

Для загрузки произвольных секций конфига используйте `config.UnmarshalByKey`:

```go
type TestData struct {
	DefaultUsername string   `mapstructure:"defaultUsername"`
	DefaultPassword string   `mapstructure:"defaultPassword"`
	ValidEmails     []string `mapstructure:"validEmails"`
}

var testData TestData
if err := config.UnmarshalByKey("testData", &testData); err != nil {
	log.Fatal(err)
}

fmt.Println(testData.DefaultUsername) // "admin"
```

---

# 🌐 HTTP Клиент

Модуль предназначен для функционального и интеграционного тестирования REST API.
Архитектурно он разделен на два пакета:

1.  **Core (`pkg/http/client`)** — Транспортный уровень. Отвечает за таймауты, хедеры по умолчанию и `net/http` обертку.
2.  **DSL (`pkg/http/dsl`)** — Fluent API уровень. Отвечает за построение сценариев, проверки (Expectations) и работу с Allure.

## Руководство по интеграции

Ниже описан полный цикл создания клиента и написания теста.

### 1. Описание Моделей

Определите структуры данных для запросов и ответов API.

**Файл:** `internal/models/auth.go`

```go
package models

type TokenCheckRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenCheckResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}
```

### 2. Создание и Конфигурация Клиента

Мы рекомендуем создавать типизированную обертку (Client Wrapper) для каждого тестируемого сервиса.
Именно здесь происходит **конфигурация** HTTP-клиента: установка базового URL, таймаутов и заголовков.

**Файл:** `internal/client/cap_client.go`

```go
package client

import (
	"time"
	"my-project/internal/models"

	"go-test-framework/pkg/http/client"
	"go-test-framework/pkg/http/dsl"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

// CapClient - обертка для конкретного сервиса (CAP)
type CapClient struct {
	http *client.Client
}

// TokenCheck - метод API, возвращающий DSL-объект Call
func (c *CapClient) TokenCheck(sCtx provider.StepCtx) *dsl.Call[models.TokenCheckRequest, models.TokenCheckResponse] {
	// Создаем Call, типизированный нашими моделями Request/Response
	return dsl.NewCall[models.TokenCheckRequest, models.TokenCheckResponse](sCtx, c.http).
		POST("/_cap/api/token/check")
}
```

### 3. Написание Теста

В тесте вы используете методы клиента. DSL позволяет сфокусироваться на данных и проверках.

**Файл:** `tests/auth_test.go`

```go
package tests

import (
	"net/http"
	"testing"
	
	"my-project/internal/client"
	"my-project/internal/models"
	
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type CAPTokenSuite struct {
	suite.Suite
}

func (s *CAPTokenSuite) TestTokenCheck(t provider.T) {
	t.Title("CAP API: Token check")
	
	// Инициализация клиента (обычно выносится в SetupSuite)
	capService := client.NewCapClient("https://api.test.env")

	t.WithNewStep("Token check request", func(sCtx provider.StepCtx) {
		capService.TokenCheck(sCtx).
			// 1. Настройка запроса
			RequestBody(models.TokenCheckRequest{
				Username: "admin",
				Password: "password",
			}).
			// 2. Декларация проверок
			ExpectResponseStatus(http.StatusOK).
			ExpectResponseBodyNotEmpty().
			ExpectResponseBodyFieldNotEmpty("token").
			ExpectResponseBodyFieldValue("success", true).
			// 3. Выполнение
			RequestSend()
	})
}

func TestCAPTokenSuite(t *testing.T) {
	suite.RunSuite(t, new(CAPTokenSuite))
}
```

### 3.1. Получение данных из ответа

Часто необходимо извлечь данные из ответа (например, токен авторизации) для использования в следующих шагах.
Метод `.Response()` возвращает типизированный объект `*client.Response[TResp]`, доступный после выполнения запроса.

```go
func (s *CAPTokenSuite) TestLoginAndUsage(t provider.T) {
    // Переменная для сохранения токена
    var authToken string

    t.WithNewStep("Login", func(sCtx provider.StepCtx) {
        // Выполняем запрос и сразу получаем объект Response
        resp := capService.TokenCheck(sCtx).
            RequestBody(models.TokenCheckRequest{
                Username: "admin", 
                Password: "admin",
            }).
            ExpectResponseStatus(http.StatusOK).
            RequestSend(). // Выполнение запроса и проверок
            Response()     // <-- Получение типизированного ответа
			
        authToken = resp.Body.Token 
    })

    t.WithNewStep("Use Token", func(sCtx provider.StepCtx) {
        // Использование полученного токена
        capService.GetProfile(sCtx, authToken).
            RequestSend()
    })
}
```

### 4. Справочник HTTP DSL

Объект `Call[TReq, TResp]` предоставляет следующие методы для построения теста:

### Настройка запроса
*   `.GET(path)`, `.POST(path)`, `.PUT(path)`, `.DELETE(path)` — Установка метода и пути (если не заданы внутри клиента).
*   `.Header(key, val)` — Добавить заголовок к запросу.
*   `.QueryParam(key, val)` — Добавить Query-параметр (`?key=val`).
*   `.PathParam(key, val)` — Подставить переменную в путь (`/users/{id}`).
*   `.RequestBody(payload)` — Установить тело запроса (структура `TReq`).

### Ожидания (Expectations)
Проверки описываются **до** отправки запроса, а выполняются **после** получения ответа.
Используется механизм "Silent Pre-checks": технические проверки скрыты, в отчете видны только бизнес-проверки.

*   `.ExpectResponseStatus(code)` — Проверяет HTTP статус код.
*   `.ExpectResponseBodyNotEmpty()` — Проверяет, что тело ответа не пустое.
*   `.ExpectResponseBodyFieldNotEmpty(path)` — Проверяет, что поле в JSON существует и не пустое.
    *   Поддерживает [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md)
    *   Примеры:
        *   `"status"` — Простое поле
        *   `"user.name"` — Вложенное поле
        *   `"items.0.id"` — Элемент массива по индексу
        *   `"items.#"` — Длина массива
        *   `"meta.pagination"` — Вложенный объект
*   `.ExpectResponseBodyFieldValue(path, expected)` — Проверяет, что значение поля JSON точно совпадает с ожидаемым.
    *   Поддерживает те же пути GJSON, что и `ExpectResponseBodyFieldNotEmpty`
    *   Поддерживаемые типы: `string`, `bool`, `int/int8/16/32/64`, `uint/uint8/16/32/64`, `float32/64`, `nil`
    *   `nil` означает, что поле существует и равно `null`
    *   Примеры:
        *   `.ExpectResponseBodyFieldValue("status", "active")` — Проверка строки
        *   `.ExpectResponseBodyFieldValue("enabled", true)` — Проверка boolean
        *   `.ExpectResponseBodyFieldValue("user.id", 123)` — Проверка числа
        *   `.ExpectResponseBodyFieldValue("price", 10.5)` — Проверка float
        *   `.ExpectResponseBodyFieldValue("users.0.username", "admin")` — Элемент массива
        *   `.ExpectResponseBodyFieldValue("items.#", 5)` — Длина массива
        *   `.ExpectResponseBodyFieldValue("deletedAt", nil)` — Проверка null

### Выполнение и Результат
*   `.RequestSend()` — Финализирующий метод.
    1.  Создает шаг в Allure.
    2.  Прикрепляет данные запроса.
    3.  Отправляет запрос.
    4.  Прикрепляет данные ответа.
    5.  Запускает все проверки.
*   `.Response()` — Возвращает типизированную структуру `*client.Response[TResp]`. Используется для извлечения данных из тела ответа для дальнейших шагов.

---

### 5. Пример отчета Allure

Благодаря группировке, отчет выглядит чисто и понятно:

*   **Token check request**
    *   **POST /_cap/api/token/check**
        *   📎 *HTTP Request* (вложение: curl, headers, body)
        *   📎 *HTTP Response* (вложение: headers, body)
        *   ✅ **Expect response status 200 OK**
        *   ✅ **Expect response body not empty**
        *   ✅ **Expect JSON field not empty: token**

---

# 🗄️ Database DSL (MySQL / PostgreSQL)

Модуль `pkg/database/dsl` предоставляет Fluent DSL для интеграционного тестирования с MySQL и PostgreSQL базами данных. Он позволяет:
- Писать типобезопасные запросы с использованием Go Generics
- Использовать декларативный стиль проверок (Expectations)
- Работать с реальными именами колонок БД через теги `db`
- Получать детальные Allure-отчеты о каждом SQL-запросе

## 1. Конфигурация БД

### MySQL

```yaml
# config.yaml
mainDatabase:
  driver: mysql  # обязательно: mysql или postgres
  dsn: "user:password@tcp(localhost:3306)/dbname?parseTime=true"
  maxOpenConns: 10
  maxIdleConns: 5
  connMaxLifetime: 5m
```

### PostgreSQL

```yaml
# config.yaml
mainDatabase:
  driver: postgres  # обязательно: mysql или postgres
  dsn: "postgres://user:password@localhost:5432/dbname?sslmode=disable"
  maxOpenConns: 10
  maxIdleConns: 5
  connMaxLifetime: 5m
```

**Параметры конфигурации:**
- `driver` — **(обязательно)** Драйвер БД: `mysql` или `postgres`
- `dsn` — Data Source Name (строка подключения)
- `maxOpenConns` — Максимальное количество открытых соединений
- `maxIdleConns` — Максимальное количество простаивающих соединений
- `connMaxLifetime` — Максимальное время жизни соединения

### Инициализируйте через DI:

```go
type TestEnv struct {
	DB *client.Client `db_config:"mainDatabase"`
}

var env TestEnv

func TestMain(m *testing.M) {
	config.BuildEnv(&env)
	code := m.Run()
	env.DB.Close()
	os.Exit(code)
}
```

## 2. Описание моделей с db тегами

Модели данных используют теги `db` для маппинга полей на колонки БД:

**Файл:** `internal/models/user.go`

```go
package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int            `db:"id"`
	Username  string         `db:"username"`
	Email     sql.NullString `db:"email"`       // Для NULL значений
	IsActive  bool           `db:"is_active"`   // MySQL TINYINT(1)
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt sql.NullTime   `db:"updated_at"`  // Nullable timestamp
}
```

**Типы для NULL значений:**
- `sql.NullString` — для nullable VARCHAR/TEXT
- `sql.NullInt64` — для nullable INT
- `sql.NullBool` — для nullable BOOLEAN
- `sql.NullTime` — для nullable TIMESTAMP/DATETIME
- `*T` (pointer) — для любых nullable типов

## 3. Создание Repository Pattern

Рекомендуется создавать репозитории для изоляции SQL-запросов:

**Файл:** `internal/db/user_repo.go`

```go
package db

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/internal/models"
	"go-test-framework/pkg/database/client"
	"go-test-framework/pkg/database/dsl"
)

type UserRepo struct {
	client *client.Client
}

func NewUserRepo(c *client.Client) *UserRepo {
	return &UserRepo{client: c}
}

func (r *UserRepo) FindByID(sCtx provider.StepCtx, id int) *dsl.Query[models.User] {
	return dsl.NewQuery[models.User](sCtx, r.client).
		SQL("SELECT id, username, email, is_active, created_at FROM users WHERE id = ?", id)
}

func (r *UserRepo) CreateUser(sCtx provider.StepCtx, username string, email string) *dsl.Query[any] {
	return dsl.NewQuery[any](sCtx, r.client).
		SQL("INSERT INTO users (username, email, is_active, created_at) VALUES (?, ?, 1, NOW())",
			username, email)
}

func (r *UserRepo) DeleteUser(sCtx provider.StepCtx, id int) *dsl.Query[any] {
	return dsl.NewQuery[any](sCtx, r.client).
		SQL("DELETE FROM users WHERE id = ?", id)
}
```

## 4. Использование в тестах

**Важно:** В тестах используйте только методы репозитория, не пишите сырые SQL-запросы.

**Пример E2E теста:**

```go
func (s *UserTestSuite) TestCreateAndVerifyUser(t provider.T) {
	t.Title("User API: Create user and verify in DB")

	var userID int64
	userRepo := db.NewUserRepo(env.DB)

	// Шаг 1: Создание пользователя через API
	t.WithNewStep("Create user via API", func(sCtx provider.StepCtx) {
		resp := env.APIClient.CreateUser(sCtx).
			RequestBody(models.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
			}).
			ExpectResponseStatus(http.StatusCreated).
			ExpectResponseBodyFieldNotEmpty("id").
			RequestSend().
			Response()

		userID = resp.Body.ID
	})

	// Шаг 2: Проверка данных в БД
	t.WithNewStep("Verify user in database", func(sCtx provider.StepCtx) {
		user := userRepo.FindByID(sCtx, int(userID)).
			ExpectFound().
			ExpectColumnEquals("username", "testuser").
			ExpectColumnEquals("email", "test@example.com").
			ExpectColumnTrue("is_active").
			ExpectColumnIsNotNull("created_at").
			MustFetch()

		// Дополнительные проверки
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email.String)
	})

	// Шаг 3: Cleanup
	t.WithNewStep("Delete test user", func(sCtx provider.StepCtx) {
		userRepo.DeleteUser(sCtx, int(userID)).MustExec()

		// Проверка удаления
		userRepo.FindByID(sCtx, int(userID)).
			ExpectNotFound().
			MustFetch()
	})
}
```

**Пример теста проверки состояния в БД:**

```go
func (s *UserTestSuite) TestUpdateUserStatus(t provider.T) {
	t.Title("User API: Update user status")

	userRepo := db.NewUserRepo(env.DB)
	userID := 123 // Тестовый пользователь

	// Обновление через API
	t.WithNewStep("Deactivate user via API", func(sCtx provider.StepCtx) {
		env.APIClient.UpdateUserStatus(sCtx, userID).
			RequestBody(models.UpdateStatusRequest{IsActive: false}).
			ExpectResponseStatus(http.StatusOK).
			RequestSend()
	})

	// Проверка в БД
	t.WithNewStep("Verify user is deactivated in DB", func(sCtx provider.StepCtx) {
		userRepo.FindByID(sCtx, userID).
			ExpectFound().
			ExpectColumnFalse("is_active").
			ExpectColumnIsNotNull("updated_at").
			MustFetch()
	})
}
```

## 5. Справочник DB DSL

### Архитектура использования

**В репозитории** (internal/db/user_repo.go) создаются методы с SQL-запросами:
```go
func (r *UserRepo) FindByID(sCtx provider.StepCtx, id int) *dsl.Query[models.User] {
	return dsl.NewQuery[models.User](sCtx, r.client).
		SQL("SELECT * FROM users WHERE id = ?", id)
}
```

**В тестах** используются только методы репозитория:
```go
user := userRepo.FindByID(sCtx, 123).
	ExpectFound().
	ExpectColumnTrue("is_active").
	MustFetch()
```

### Методы Query объекта

#### Настройка запроса (используется в репозитории)

- `.SQL(query, args...)` — Устанавливает SQL-запрос и его аргументы
- `.WithContext(ctx)` — Устанавливает кастомный context.Context (опционально)

#### Expectations (проверки) — используются в тестах

Проверки выполняются **после** получения результата от БД:

**Проверки наличия данных:**
- `.ExpectFound()` — Проверяет, что запрос вернул хотя бы одну строку
- `.ExpectNotFound()` — Проверяет, что запрос не вернул строк (`sql.ErrNoRows`)

**⚠️ ВАЖНО: Поведение MustFetch() по умолчанию**

`MustFetch()` **требует наличие результата** по умолчанию. Если запрос возвращает `sql.ErrNoRows` без явного ожидания `ExpectNotFound()`, тест **упадет** с ошибкой:
> "Expected row to exist, but got sql.ErrNoRows. Use ExpectNotFound() if 'not found' is an expected scenario"

Это предотвращает **ложноположительные прогоны** тестов, когда отсутствие данных проходит незамеченным.

**Правильное использование:**
```go
// ✅ Ожидаем данные (по умолчанию)
user := userRepo.FindByID(sCtx, 1).
    ExpectFound().  // Опционально, для явности
    MustFetch()

// ✅ Явно разрешаем "не найдено"
userRepo.FindByID(sCtx, 999).
    ExpectNotFound().  // Обязательно!
    MustFetch()

// ❌ Опасно: тест пройдет даже если данных нет (старое поведение)
// Теперь это вызовет ошибку
userRepo.FindByID(sCtx, 1).
    MustFetch()  // Упадет, если данных нет
```

**Проверки значений колонок:**
- `.ExpectColumnEquals(columnName, expectedValue)` — Проверяет равенство значения колонки
- `.ExpectColumnNotEmpty(columnName)` — Проверяет, что значение не пустое
- `.ExpectColumnTrue(columnName)` — Проверяет, что значение `true` (для TINYINT(1) или bool)
- `.ExpectColumnFalse(columnName)` — Проверяет, что значение `false`

**Проверки NULL:**
- `.ExpectColumnIsNull(columnName)` — Проверяет, что значение NULL
- `.ExpectColumnIsNotNull(columnName)` — Проверяет, что значение NOT NULL

**Примечание:** Имена колонок (`columnName`) соответствуют тегам `db` в Go-структурах.

#### Выполнение запроса

**SELECT запросы** — используйте `.MustFetch()`:

```go
// В репозитории определен метод FindByID
user := userRepo.FindByID(sCtx, 1).
	ExpectFound().
	ExpectColumnTrue("is_active").
	MustFetch()
```

`.MustFetch()` — выполняет SELECT запрос, сканирует одну строку в `T` и возвращает заполненный объект.

**INSERT/UPDATE/DELETE запросы** — используйте `.MustExec()`:

```go
// В репозитории определен метод CreateUser
result := userRepo.CreateUser(sCtx, "john", "john@example.com").
	MustExec()

lastID, _ := result.LastInsertId()
rowsAffected, _ := result.RowsAffected()
```

`.MustExec()` — выполняет не-SELECT запрос и возвращает `sql.Result`.

### Особенности работы с типами MySQL

#### TINYINT(1) как Boolean

MySQL использует `TINYINT(1)` для boolean значений. Фреймворк автоматически:
- В `ExpectColumnEquals` конвертирует `int64` ↔ `bool` при сравнении
- В `ExpectColumnTrue/False` поддерживает как `bool`, так и `int64` (1/0)

В Go-моделях используйте:
```go
IsActive bool `db:"is_active"`  // MySQL TINYINT(1)
```

#### NULL значения

Для nullable колонок используйте типы из `database/sql`:

```go
Email sql.NullString `db:"email"`
Age   sql.NullInt64  `db:"age"`
```

Проверки:
```go
.ExpectColumnIsNull("email")     // Email IS NULL
.ExpectColumnIsNotNull("email")  // Email IS NOT NULL
```

### Allure отчеты

Каждый запрос автоматически создает шаг в Allure с вложениями:

**DB Fetch: SELECT * FROM users...**
- 📎 **SQL Query** (текст запроса + аргументы)
- 📎 **SQL Result** (JSON с данными)
- ✅ **Expect: Found**
- ✅ **Expect: Column 'is_active' is true**

**DB Exec: INSERT INTO users...**
- 📎 **SQL Query** (текст запроса + аргументы)
- 📎 **SQL Exec Result** (`{"rowsAffected": 1, "lastInsertId": 42}`)

---