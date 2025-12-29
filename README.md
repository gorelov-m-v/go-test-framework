# Go Test Framework

Мощный и типобезопасный фреймворк для написания автотестов на Go. Спроектирован для построения сложных интеграционных тестов с глубокой интеграцией в **Allure Report**.

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
    *   [Переопределение через переменные окружения](#переопределение-через-переменные-окружения)
*   [**HTTP Клиент**](#-http-клиент)
    *   [1. Описание Моделей](#1-описание-моделей)
    *   [2. Создание и Конфигурация Клиента](#2-создание-и-конфигурация-клиента)
    *   [3. Написание Теста](#3-написание-теста)
    *   [3.1. Получение данных из ответа](#31-получение-данных-из-ответа)
    *   [4. Справочник HTTP DSL](#4-справочник-http-dsl)
    *   [5. Пример отчета Allure](#5-пример-отчета-allure)
*   NATS Клиент
*   Kafka Клиент
*   Redis Клиент

---

# ⚙️ Конфигурация

Модуль `pkg/config` предоставляет мощный механизм управления конфигурацией для ваших тестов. Он автоматически загружает настройки из YAML-файлов, поддерживает переопределение через переменные окружения и использует DI-подобную технику для автоматической инициализации HTTP-клиентов.

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

Ваш клиент **обязан** иметь публичное поле `HTTP *httpclient.Client`:

```go
package client

import (
	"go-test-framework/pkg/httpclient"
	"go-test-framework/pkg/httpdsl"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type CapClient struct {
	HTTP *httpclient.Client  // ⚠️ Обязательное поле для автоинъекции
}

func (c *CapClient) TokenCheck(sCtx provider.StepCtx) *httpdsl.Call[TokenCheckRequest, TokenCheckResponse] {
	return httpdsl.NewCall[TokenCheckRequest, TokenCheckResponse](sCtx, c.HTTP).
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

### Шаг 4: Используйте в тестах

```go
func (s *CAPTokenSuite) TestTokenCheck(t provider.T) {
	t.Title("CAP API: Token check")

	t.WithNewStep("Token check request", func(sCtx provider.StepCtx) {
		// env.CapClient уже полностью инициализирован!
		env.CapClient.TokenCheck(sCtx).
			RequestBody(models.TokenCheckRequest{
				Username: "admin",
				Password: "admin",
			}).
			ExpectResponseStatus(http.StatusOK).
			ExpectResponseBodyFieldNotEmpty("token").
			RequestSend()
	})
}
```

## Переопределение через переменные окружения

Вы можете переопределить любое значение из `config.yaml` через переменные окружения. Формат: `SECTION_KEY=value` (точки заменяются на подчеркивания).

**Примеры:**

```bash
# Переопределить baseURL для capService
export CAPSERVICE_BASEURL=https://prod.example.com

# Переопределить timeout
export CAPSERVICE_TIMEOUT=60s

# Запустить тесты
go test ./tests/...
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

1.  **Core (`pkg/httpclient`)** — Транспортный уровень. Отвечает за таймауты, хедеры по умолчанию и `net/http` обертку.
2.  **DSL (`pkg/httpdsl`)** — Fluent API уровень. Отвечает за построение сценариев, проверки (Expectations) и работу с Allure.

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
Именно здесь происходит **конфигурация** `httpclient`: установка базового URL, таймаутов и заголовков.

**Файл:** `internal/client/cap_client.go`

```go
package client

import (
	"time"
	"my-project/internal/models"
	
	"go-test-framework/pkg/httpclient"
	"go-test-framework/pkg/httpdsl"
	
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

// CapClient - обертка для конкретного сервиса (CAP)
type CapClient struct {
	http *httpclient.Client
}

// NewCapClient инициализирует httpclient с нужной конфигурацией
func NewCapClient(baseURL string) *CapClient {
	return &CapClient{
		// КОНФИГУРАЦИЯ ЗДЕСЬ
		http: httpclient.New(httpclient.Config{
			BaseURL: baseURL,
			Timeout: 30 * time.Second,
			DefaultHeaders: map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
		}),
	}
}

// TokenCheck - метод API, возвращающий DSL-объект Call
func (c *CapClient) TokenCheck(sCtx provider.StepCtx) *httpdsl.Call[models.TokenCheckRequest, models.TokenCheckResponse] {
	// Создаем Call, типизированный нашими моделями Request/Response
	return httpdsl.NewCall[models.TokenCheckRequest, models.TokenCheckResponse](sCtx, c.http).
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
Метод `.Response()` возвращает типизированный объект `*httpclient.Response[TResp]`, доступный после выполнения запроса.

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

        // Безопасно сохраняем данные (IDE подскажет поля структуры)
        // Если ExpectStatus упал, тест остановится раньше, и паники не будет.
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
*   `.ExpectResponseBodyFieldNotEmpty(jsonPath)` — Проверяет, что поле в JSON существует и не пустое (поддерживает вложенность: `"data.user.id"`).

### Выполнение и Результат
*   `.RequestSend()` — Финализирующий метод.
    1.  Создает шаг в Allure.
    2.  Прикрепляет данные запроса.
    3.  Отправляет запрос.
    4.  Прикрепляет данные ответа.
    5.  Запускает все проверки.
*   `.Response()` — Возвращает типизированную структуру `*httpclient.Response[TResp]`. Используется для извлечения данных из тела ответа для дальнейших шагов.

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