
# CLAUDE.md

This file provides guidance to Claude Code when working with the Go E2E Test Framework.

## Project Overview
A declarative, DSL-based E2E test framework for Go. It uses strict typing (Generics), Fluent Interface, and built-in async retries.

## Commands
```bash
# Run all tests
go test -v ./...

# Run single test
go test -v -run TestName ./tests/...

# Generate Allure report
allure serve allure-results

# Debug DI injection
GO_TEST_FRAMEWORK_DEBUG=1 go test -v ./...
```

## Recommended Project Structure

```
your-api-tests/
├── configs/
│   └── config.local.yaml         # Configuration (ENV=local by default)
│
├── internal/
│   ├── client/                   # HTTP clients (DSL methods)
│   │   └── [service_name]/       # One folder per microservice
│   │       └── client.go         # Link struct + DSL methods
│   │
│   ├── db/                       # Database repositories
│   │   └── [table_name]/         # One folder per table/domain
│   │       └── repo.go           # Link struct + DSL methods
│   │
│   ├── models/
│   │   ├── http/                 # HTTP models (by service)
│   │   │   └── [service_name]/
│   │   │       └── *.go          # Request/Response structs with `json` tags
│   │   └── db/                   # DB models (by database)
│   │       └── [database_name]/
│   │           └── *.go          # Table structs with `db` tags
│   │
│   └── kafka/
│       └── topics.go             # Topic types + message models
│
├── tests/
│   ├── env.go                    # TestEnv struct with DI tags
│   └── *_test.go                 # Test suites
│
├── allure-results/               # Auto-generated Allure results
├── go.mod
└── go.sum
```

### Where to Create Files
| What | Where |
|------|-------|
| New HTTP client for service "auth" | `internal/client/auth/client.go` |
| New DB repo for table "orders" | `internal/db/orders/repo.go` |
| HTTP models for "auth" service | `internal/models/http/auth/*.go` |
| DB models for "core" database | `internal/models/db/core/*.go` |
| Kafka topics and messages | `internal/kafka/topics.go` |
| New test suite | `tests/*_test.go` |
| Environment/DI setup | `tests/env.go` |
| Configuration | `configs/config.{ENV}.yaml` |

## Coding Rules
1.  **Strict Generics:** Always specify types: `dsl.NewCall[Req, Resp]` or `dsl.NewQuery[Model]`.
2.  **No time.Sleep:** Use `s.AsyncStep` for retries. Use `s.Step` for immediate checks.
3.  **Link Pattern:** All clients/repos must implement `Link` struct and be registered in `TestEnv`.

---

## DSL API Reference (Don't guess methods, use these)

### 1. HTTP DSL (`dsl.Call[Req, Resp]`)
**Setup:**
- `.GET("/path")` / `.POST("/path")` / `.PUT("/path")` / `.PATCH("/path")` / `.DELETE("/path")`
- `.RequestBody(reqModel)` - for typed requests
- `.RequestBodyMap(map[string]interface{}{...})` - for negative tests (missing fields, extra fields, wrong types)
- `.Header("Key", "Val")` / `.QueryParam("Key", "Val")` / `.PathParam("key", "val")`

**Expectations (Chain before .Send):**
- `.ExpectResponseStatus(200)`
- `.ExpectResponseBodyFieldValue("json.path", value)`
- `.ExpectResponseBodyFieldNotEmpty("json.path")`
- `.ExpectResponseBodyFieldIsNull("json.path")`
- `.ExpectResponseHeader("Key", "Val")`

**Execute:**
- `.Send()` -> Returns `*client.Response[Resp]`

### 2. Database DSL (`dsl.Query[Model]`)
**Setup:**
- `.SQL("SELECT * FROM table WHERE id = ?", id)` (Use `?` for MySQL, `$1` for Postgres)

**Expectations:**
- `.ExpectFound()` / `.ExpectNotFound()`
- `.ExpectColumnEquals("db_tag_name", value)`
- `.ExpectColumnTrue("is_active")` / `.ExpectColumnFalse("is_deleted")`
- `.ExpectColumnIsNotNull("created_at")` / `.ExpectColumnIsNull("updated_at")`

**Execute:**
- `.Send()` -> Returns `Model` struct

### 3. Kafka DSL (`dsl.Expect[Topic]`)
**Setup (Filters):**
- `.With("json.field", value)` (AND logic for multiple filters)
- `.Unique()` / `.UniqueWithWindow(duration)`

**Expectations:**
- `.ExpectField("json.path", value)`
- `.ExpectFieldNotEmpty("id")`
- `.ExpectFieldTrue("isActive")` / `.ExpectFieldFalse("isDeleted")`
- `.ExpectFieldIsNull("field")` / `.ExpectFieldIsNotNull("field")`

**Execute:**
- `.Send()` -> Returns nothing (fails test if not found)

### 4. gRPC DSL (`dsl.Call[Req, Resp]`)
**Setup:**
- `.Method("/package.Service/Method")` - Full gRPC method path
- `.Service("player.PlayerService")` + `.MethodName("CreatePlayer")` - Alternative syntax
- `.RequestBody(reqModel)` - Protobuf request message
- `.Metadata("key", "value")` / `.MetadataMap(map[string]string{})`

**Expectations:**
- `.ExpectNoError()` / `.ExpectError()`
- `.ExpectStatusCode(codes.OK)` - gRPC status code
- `.ExpectFieldValue("json.path", value)` - Uses GJSON paths
- `.ExpectFieldNotEmpty("path")` / `.ExpectFieldExists("path")`
- `.ExpectMetadata("key", "value")`

**Execute:**
- `.Send()` -> Returns `*client.Response[Resp]`

### 5. Redis DSL (`dsl.Query`)
**Setup:**
- `.Key("player:123")` - Redis key to query

**Expectations:**
- `.ExpectExists()` / `.ExpectNotExists()`
- `.ExpectValue("expected_string")` / `.ExpectValueNotEmpty()`
- `.ExpectJSONField("json.path", value)` / `.ExpectJSONFieldNotEmpty("path")`
- `.ExpectTTL(minDuration, maxDuration)` / `.ExpectNoTTL()`

**Execute:**
- `.Send()` -> Returns `*client.Result`

**Utilities (for setup/cleanup):**
- `client.Set(ctx, "key", "value", ttl)`
- `client.Del(ctx, "key1", "key2")`
- `client.RDB()` -> Returns underlying `*redis.Client`

---

## Step Types: Step vs AsyncStep

### `s.Step()` - Synchronous, Immediate Failure
- Uses `Require` assertions (stops test on first error)
- No retries
- Use for: API calls that must succeed immediately (create resource, get immediate response)

### `s.AsyncStep()` - Async with Retries + Parallel Execution
- Uses `Assert` assertions (accumulates errors, retries)
- Automatic retries with exponential backoff when expectations are present
- **Adjacent AsyncSteps run in PARALLEL goroutines**
- `Step()` automatically waits for all preceding `AsyncStep()` to complete
- Use for: Waiting for DB records, Kafka events, status changes

**Parallel Execution Example:**
```go
// These 3 run SIMULTANEOUSLY
s.AsyncStep(t, "Wait DB", func(sCtx provider.StepCtx) { ... })
s.AsyncStep(t, "Wait Kafka", func(sCtx provider.StepCtx) { ... })
s.AsyncStep(t, "Wait Status", func(sCtx provider.StepCtx) { ... })

// This Step waits for ALL 3 above to complete
s.Step(t, "Verify", func(sCtx provider.StepCtx) { ... })
```

---

## GJSON Path Syntax (for JSON field access)
- Simple field: `"name"`
- Nested field: `"user.email"`
- Array element: `"items.0"`, `"items.1"`
- Array count: `"items.#"`
- Nested in array: `"users.0.name"`

---

## Configuration Structure (`configs/config.local.yaml`)

```yaml
http:
  serviceName:
    baseURL: "https://api.example.com"
    timeout: 30s
    maskHeaders: "Authorization,Cookie"  # Masked in Allure reports

database:
  dbName:
    driver: "postgres"  # or "mysql"
    dsn: "postgres://user:pass@host:5432/db?sslmode=disable"
    maskColumns: "password,api_token"  # Masked in Allure reports

kafka:
  bootstrapServers: ["kafka:9092"]
  groupId: "qa-test-group"
  topics: ["events-topic"]
  bufferSize: 1000

grpc:
  serviceName:
    target: "localhost:9090"
    insecure: true  # No TLS for local development

redis:
  cacheName:
    addr: "localhost:6379"
    password: ""
    db: 0

# Async retry settings per DSL
http_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
    backoff: { enabled: true, factor: 1.5, max_interval: 1s }
    jitter: 0.2

db_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms

kafka_dsl:
  async:
    enabled: true
    timeout: 30s
    interval: 200ms

grpc_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms

redis_dsl:
  async:
    enabled: true
    timeout: 10s
    interval: 200ms
```

---

## TestEnv Setup (DI via struct tags)

```go
type TestEnv struct {
    // HTTP clients - tag links config key to Link struct
    GameService game.Link `config:"serviceName"`

    // Database repos
    PlayersRepo players.Link `db_config:"dbName"`

    // Kafka
    Kafka kafka.Link `kafka_config:"kafka"`

    // gRPC clients
    PlayerGRPC playergrpc.Link `grpc_config:"grpc.serviceName"`

    // Redis
    RedisCache rediscache.Link `redis_config:"redis.cacheName"`
}

func init() {
    env = &TestEnv{}
    if err := builder.BuildEnv(env); err != nil {
        log.Fatalf("Failed to build env: %v", err)
    }
}
```

---

## Link Pattern (Client/Repo Implementation)

```go
package game

var httpClient *client.Client

type Link struct{}

func (l *Link) SetHTTP(c *client.Client) {
    httpClient = c
}

// DSL Method
func CreatePlayer(sCtx provider.StepCtx) *dsl.Call[models.CreateReq, models.CreateResp] {
    return dsl.NewCall[models.CreateReq, models.CreateResp](sCtx, httpClient).
        POST("/api/v1/players")
}
```

---

## Full E2E Test Example

```go
func (s *PlayerSuite) TestCreatePlayerE2E(t provider.T) {
    var playerID string
    username := "test_user"

    // Step 1: HTTP - Create (immediate, no retry)
    s.Step(t, "Create player", func(sCtx provider.StepCtx) {
        resp := game.CreatePlayer(sCtx).
            RequestBody(models.CreateReq{Username: username}).
            ExpectResponseStatus(201).
            ExpectResponseBodyFieldNotEmpty("id").
            Send()
        playerID = resp.Body.ID
    })

    // Step 2: DB - Wait for record (async with retry)
    s.AsyncStep(t, "Verify in DB", func(sCtx provider.StepCtx) {
        players.FindByID(sCtx, playerID).
            ExpectFound().
            ExpectColumnEquals("username", username).
            ExpectColumnEquals("status", "active").
            Send()
    })

    // Step 3: Kafka - Wait for event (async with retry)
    s.AsyncStep(t, "Verify Kafka event", func(sCtx provider.StepCtx) {
        kafkaDSL.Expect[kafka.PlayerEventsTopic](sCtx, kafka.Client()).
            With("playerId", playerID).
            With("eventType", "PLAYER_CREATED").
            ExpectField("playerName", username).
            Send()
    })
}
```

---
## Best Practices: "Full Context" Pattern**Recommendation:** Store full response/DB object structures instead of extracting individual fields.### ❌ Avoid (old approach):```gofunc (s *Suite) TestFlow(t provider.T) {    var userID string    var userURL string    var email string    s.Step(t, "Register", func(ctx provider.StepCtx) {        resp := auth.Register(ctx).            RequestBody(models.RegisterRequest{...}).            Send()        userID = resp.Body.ID       // Extracting individual fields        userURL = resp.Body.UserURL        email = resp.Body.Email    })    s.Step(t, "Verify", func(ctx provider.StepCtx) {        // Lost context: Where did userURL come from?        auth.Verify(ctx, userURL).Send()    })}```### ✅ Use (Full Context):```gofunc (s *Suite) TestFlow(t provider.T) {    // 1. Define: Declare full structures at the beginning    var (        regResp *client.Response[models.RegisterResp] // Store entire HTTP response        userDB  models.UserDB                          // Store entire DB row    )    // 2. Capture: Fill the variable    s.Step(t, "Register", func(ctx provider.StepCtx) {        regResp = auth.Register(ctx).            RequestBody(models.RegisterRequest{...}).            Send()    })    // 3. Use: Use data with clear context    s.Step(t, "Verify", func(ctx provider.StepCtx) {        // Context is obvious: searching by ID from registration response        auth.Verify(ctx, regResp.Body.UserURL).            RequestBody(models.VerifyRequest{                Email:    regResp.Body.Email,    // <- All fields accessible                Password: "test",            }).            Send()    })    s.AsyncStep(t, "Check DB", func(ctx provider.StepCtx) {        userDB = users.FindByID(ctx, regResp.Body.ID).            ExpectColumnEquals("email", regResp.Body.Email).            Send()    })}```**Benefits:**- **Self-documenting code:** `regResp.Body.ID` is clearer than anonymous `userID`- **Flexibility:** If you need an additional field later - it is already available- **Type safety:** Compiler knows all field types- **Data traceability:** Easy to see where each value comes from**When to simplify:**For simple single-step negative tests, you do not need to overcomplicate:```gofunc (s *Suite) TestEmailEmpty(t provider.T) {    // No need for var if only used here    s.Step(t, "Send", func(ctx provider.StepCtx) {        auth.Register(ctx).            RequestBody(models.RegisterRequest{Email: ""}).            ExpectResponseStatus(422).            Send()    })}```---

## Parametrized Tests (Table-Driven Tests)

For testing multiple scenarios with different data (especially negative tests):

### Structure:
```go
// 1. Define test case struct
type EmailTestCase struct {
    Name           string
    Email          string
    ExpectedStatus int
    ExpectedCode   string
}

// 2. Add parameter field to Suite (MUST be Param + <method suffix>)
type RegisterNegativeSuite struct {
    extension.BaseSuite
    ParamEmailValidation []EmailTestCase  // For TableTestEmailValidation
}

// 3. Initialize in BeforeAll
func (s *RegisterNegativeSuite) BeforeAll(t provider.T) {
    s.ParamEmailValidation = []EmailTestCase{
        {Name: "Empty email", Email: "", ExpectedStatus: 422, ExpectedCode: "EMAIL_IS_EMPTY"},
        {Name: "Invalid email", Email: "test", ExpectedStatus: 422, ExpectedCode: "INVALID_EMAIL"},
    }
}

// 4. Create TableTest method (MUST start with TableTest)
func (s *RegisterNegativeSuite) TableTestEmailValidation(t provider.T, tc EmailTestCase) {
    t.Title(tc.Name)
    s.Step(t, "Test email validation", func(sCtx provider.StepCtx) {
        auth.Register(sCtx).
            RequestBody(models.RegisterRequest{Email: tc.Email}).
            ExpectResponseStatus(tc.ExpectedStatus).
            Send()
    })
}

// 5. Run with RunSuite (NOT RunNamedSuite)
func TestRegisterNegativeSuite(t *testing.T) {
    suite.RunSuite(t, new(RegisterNegativeSuite))
}
```

### Using RequestBodyMap for negative tests:
```go
// Test missing fields (field completely absent from JSON)
s.Step(t, "Register without email", func(sCtx provider.StepCtx) {
    auth.Register(sCtx).
        RequestBodyMap(map[string]interface{}{
            "password": "P@ssw0rd",
            // email field is completely missing
        }).
        ExpectResponseStatus(400).
        ExpectResponseBodyFieldValue("detail.code", "EMAIL_REQUIRED").
        Send()
})

// Test with extra fields
s.Step(t, "Register with extra field", func(sCtx provider.StepCtx) {
    auth.Register(sCtx).
        RequestBodyMap(map[string]interface{}{
            "email":    "test@test.com",
            "password": "P@ssw0rd",
            "extra":    "unexpected",
        }).
        ExpectResponseStatus(400).
        Send()
})
```

**IMPORTANT:**
- Parameter field name MUST match pattern: `Param` + `<TableTest method suffix>`
- Method name MUST start with `TableTest`
- Use `suite.RunSuite()`, NOT `suite.RunNamedSuite()`
- Each test case will appear as separate test in Allure

### Data Reuse Patterns in Parametrized Tests

**Each test case gets its own isolated copy of data** - this is important for parallel execution.

#### ✅ Pattern 1: Local Variables (Recommended)
Use **local variables** in BeforeAll for data that needs to be shared across multiple test cases:

```go
type RegisterNegativeSuite struct {
    extension.BaseSuite
    ParamEmailValidation []EmailTestCase
}

func (s *RegisterNegativeSuite) BeforeAll(t provider.T) {
    // Local variables - safe for parallel execution
    validPassword := datagen.Password(8)
    validEmail := datagen.Email(10)

    s.ParamEmailValidation = []EmailTestCase{
        {
            Name:     "Empty email",
            Email:    "",
            Password: validPassword,  // Copied into test case
            ExpectedCode: "EMAIL_IS_EMPTY",
        },
        {
            Name:     "Invalid email format",
            Email:    "invalid-email",
            Password: validPassword,  // Same password, copied into test case
            ExpectedCode: "INVALID_EMAIL_FORMAT",
        },
    }
}
```

**Why local variables?**
- ✅ Thread-safe: Each test case gets a copy of the value
- ✅ No shared state between test cases
- ✅ Memory efficient: Variable is discarded after BeforeAll
- ✅ Safe for parallel test execution

#### ✅ Pattern 2: Generate Per Test Case
For data that should be unique per test case, generate directly in the test case definition:

```go
func (s *RegisterNegativeSuite) BeforeAll(t provider.T) {
    validEmail := datagen.Email(10)

    s.ParamPasswordValidation = []PasswordTestCase{
        {
            Name:     "Password too short",
            Email:    validEmail,
            Password: datagen.Password(4),  // Generated once, unique for this case
            ExpectedCode: "INVALID_PASSWORD",
        },
        {
            Name:     "Password without uppercase",
            Email:    validEmail,
            Password: datagen.Password(8, datagen.LatinLower, datagen.Digits),  // Unique
            ExpectedCode: "INVALID_PASSWORD",
        },
    }
}
```

#### ✅ Pattern 3: Setup with Real Data
When test cases need data from actual API/DB operations:

```go
type VerifyNegativeSuite struct {
    extension.BaseSuite
    ParamBusinessLogic []VerifyTestCase
}

func (s *VerifyNegativeSuite) BeforeAll(t provider.T) {
    var (
        regResp  *client.Response[models.RegisterResponse]
        userDB   models.User
        email    = datagen.Email(10)
        password = datagen.Password(8)
    )

    // Setup: Create real test data
    t.WithNewStep("Setup: Register test user", func(sCtx provider.StepCtx) {
        regResp = auth.Register(sCtx).
            RequestBody(models.RegisterRequest{Email: email, Password: password}).
            ExpectResponseStatus(201).
            Send()
    })

    t.WithNewStep("Setup: Get user from DB", func(sCtx provider.StepCtx) {
        userDB = users.FindByID(sCtx, regResp.Body.ID).
            ExpectFound().
            Send()
    })

    // Use real data in test cases
    s.ParamBusinessLogic = []VerifyTestCase{
        {
            Name:          "User URL does not exist",
            UserURL:       "nonexistent-url",
            Code:          userDB.VerificationCode.String,  // Real code
            Email:         email,
            Password:      password,
            ExpectedError: "User not found",
        },
        {
            Name:          "Verification code is incorrect",
            UserURL:       regResp.Body.UserURL,  // Real URL
            Code:          "999999",              // Wrong code
            Email:         email,
            Password:      password,
            ExpectedError: "Invalid code",
        },
    }
}
```

#### ❌ Avoid: Suite Fields for Shared Data
**Do NOT store shared data in suite fields** - it creates global state:

```go
// ❌ BAD: Suite field creates shared mutable state
type RegisterNegativeSuite struct {
    extension.BaseSuite
    validPassword string  // Avoid: global state
    ParamEmailValidation []EmailTestCase
}

func (s *RegisterNegativeSuite) BeforeAll(t provider.T) {
    s.validPassword = datagen.Password(8)  // Avoid: writes to suite field

    s.ParamEmailValidation = []EmailTestCase{
        {Password: s.validPassword},  // Avoid: reading from suite field
    }
}
```

**Why avoid suite fields?**
- ❌ Potential race conditions if test framework runs tests in parallel
- ❌ Harder to trace where data comes from
- ❌ Creates coupling between BeforeAll and test execution
- ❌ Fields persist in memory for entire suite lifetime

**Key Principle:**
> Test cases should be **self-contained with immutable data**. Data is captured at BeforeAll time and each test case operates on its own copy.
