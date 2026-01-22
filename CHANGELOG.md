# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.3] - 2026-01-22

### Added
- Unit tests for HTTP client (`buildBody`, `validateBuildInput`, `countTrue`)
- Unit tests for HTTP DSL Call builder methods and validation
- Mock `StepCtx` and `Asserts` for testing

### Changed
- HTTP client coverage: 38.4% → 64.2%
- HTTP DSL coverage: 38.3% → 57.5%

## [0.9.0] - 2026-01-21

### Added
- Struct-based assertion methods for HTTP, Kafka, and DB DSL
- `ExpectResponseBody()` and `ExpectResponseBodyPartial()` for HTTP
- `ExpectRow()` and `ExpectRowPartial()` for Database
- `ExpectMessage()` and `ExpectMessagePartial()` for Kafka
- `ExpectArrayContains()` and `ExpectArrayContainsExact()` for JSON array assertions
- Tests for `map[string]interface{}`, `interface{}`, and `*string` comparison

### Fixed
- Support pointer to struct in `CompareObjectExact()` and `CompareObjectPartial()`

## [0.7.0] - 2026-01-20

### Added
- Environment-based prefixes for Kafka topics
- Environment-based prefixes for DB schemas

## [0.6.0] - 2026-01-19

### Added
- Contract testing against OpenAPI specs
- `ExpectMatchesContract()` method for HTTP DSL
- `ExpectMatchesSchema()` method for explicit schema validation
- `ExpectCount()` method to Kafka DSL
- `BaseSuite.Cleanup()` for test cleanup

### Fixed
- Dynamic config path lookup
- Preserve API path prefixes in generated clients

## [0.5.0] - 2026-01-17

### Added
- `ExpectResponseBodyFieldTrue()` and `ExpectResponseBodyFieldFalse()` to HTTP DSL
- `ExpectResponseBodyFieldIsNull()` and `ExpectResponseBodyFieldIsNotNull()` to HTTP DSL
- Codegen marker comments in generated files

### Changed
- Major project structure refactoring
- Clear separation between `pkg/` (public API) and `internal/` (implementation)

### Fixed
- Handle duplicate names for same path with different HTTP methods in codegen
- OpenAPI templates improvements

## [0.3.0] - 2026-01-14

### Added
- gRPC DSL with `Call[Req, Resp]` generic struct
- Redis DSL with `Query` struct
- `grpc-gen` CLI tool for generating gRPC clients from .proto files
- `test-init` CLI tool for scaffolding new test projects

## [0.2.0] - 2026-01-13

### Added
- `openapi-gen` CLI tool for generating HTTP clients from OpenAPI specs
- HTTP client code generation with models

## [0.1.0] - 2026-01-12

### Added
- Initial release
- HTTP DSL with fluent interface (`Call[Req, Resp]`)
- Database DSL (`Query[T]`) for PostgreSQL and MySQL
- Kafka DSL for event validation
- DI container with struct tags (`config`, `db_config`, `kafka_config`)
- Allure reporting integration
- Automatic retry with exponential backoff
- GJSON-based JSON path access

[Unreleased]: https://github.com/gorelov-m-v/go-test-framework/compare/v0.9.3...HEAD
[0.9.3]: https://github.com/gorelov-m-v/go-test-framework/compare/v0.9.0...v0.9.3
[0.9.0]: https://github.com/gorelov-m-v/go-test-framework/compare/v0.7.0...v0.9.0
[0.7.0]: https://github.com/gorelov-m-v/go-test-framework/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/gorelov-m-v/go-test-framework/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/gorelov-m-v/go-test-framework/compare/v0.3.0...v0.5.0
[0.3.0]: https://github.com/gorelov-m-v/go-test-framework/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/gorelov-m-v/go-test-framework/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/gorelov-m-v/go-test-framework/releases/tag/v0.1.0
