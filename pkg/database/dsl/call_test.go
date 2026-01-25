package dsl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTableName_SimpleSelect(t *testing.T) {
	result := extractTableName("SELECT * FROM users")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_SelectWithColumns(t *testing.T) {
	result := extractTableName("SELECT id, name, email FROM users")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_SelectWithWhere(t *testing.T) {
	result := extractTableName("SELECT * FROM orders WHERE id = 1")
	assert.Equal(t, "orders", result)
}

func TestExtractTableName_LowercaseFrom(t *testing.T) {
	result := extractTableName("select * from products")
	assert.Equal(t, "products", result)
}

func TestExtractTableName_MixedCase(t *testing.T) {
	result := extractTableName("SELECT * From Users WHERE id = 1")
	assert.Equal(t, "Users", result)
}

func TestExtractTableName_WithSchema(t *testing.T) {
	result := extractTableName("SELECT * FROM public.users")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithSchemaAndQuotes(t *testing.T) {
	result := extractTableName("SELECT * FROM `public`.`users`")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithDoubleQuotes(t *testing.T) {
	result := extractTableName(`SELECT * FROM "users"`)
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithBackticks(t *testing.T) {
	result := extractTableName("SELECT * FROM `users`")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithAlias(t *testing.T) {
	result := extractTableName("SELECT u.id FROM users u")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithAsAlias(t *testing.T) {
	result := extractTableName("SELECT u.id FROM users AS u")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithCTE(t *testing.T) {
	query := `WITH cte AS (SELECT * FROM other) SELECT * FROM users`
	result := extractTableName(query)
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithCTEComplex(t *testing.T) {
	query := `WITH
		cte1 AS (SELECT * FROM table1),
		cte2 AS (SELECT * FROM table2)
	SELECT * FROM main_table`
	result := extractTableName(query)
	assert.Equal(t, "main_table", result)
}

func TestExtractTableName_WithNestedCTE(t *testing.T) {
	query := `WITH cte AS (SELECT * FROM (SELECT * FROM nested) n) SELECT * FROM target`
	result := extractTableName(query)
	assert.Equal(t, "target", result)
}

func TestExtractTableName_Join(t *testing.T) {
	query := "SELECT * FROM users JOIN orders ON users.id = orders.user_id"
	result := extractTableName(query)
	assert.Equal(t, "users", result)
}

func TestExtractTableName_LeftJoin(t *testing.T) {
	query := "SELECT * FROM products LEFT JOIN categories ON products.cat_id = categories.id"
	result := extractTableName(query)
	assert.Equal(t, "products", result)
}

func TestExtractTableName_Subquery(t *testing.T) {
	query := "SELECT * FROM (SELECT * FROM users) AS subq"
	result := extractTableName(query)
	assert.Contains(t, []string{"(SELECT", "users"}, result)
}

func TestExtractTableName_MultipleSpaces(t *testing.T) {
	result := extractTableName("SELECT   *   FROM    users")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithNewlines(t *testing.T) {
	query := `SELECT *
FROM users
WHERE id = 1`
	result := extractTableName(query)
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithTabs(t *testing.T) {
	result := extractTableName("SELECT\t*\tFROM\tusers")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_NoFrom(t *testing.T) {
	result := extractTableName("SELECT 1")
	assert.Equal(t, "query", result)
}

func TestExtractTableName_EmptyString(t *testing.T) {
	result := extractTableName("")
	assert.Equal(t, "query", result)
}

func TestExtractTableName_OnlyWhitespace(t *testing.T) {
	result := extractTableName("   ")
	assert.Equal(t, "query", result)
}

func TestExtractTableName_WithLimit(t *testing.T) {
	result := extractTableName("SELECT * FROM users LIMIT 10")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithOrderBy(t *testing.T) {
	result := extractTableName("SELECT * FROM users ORDER BY created_at DESC")
	assert.Equal(t, "users", result)
}

func TestExtractTableName_WithGroupBy(t *testing.T) {
	result := extractTableName("SELECT status, COUNT(*) FROM orders GROUP BY status")
	assert.Equal(t, "orders", result)
}

func TestExtractTableName_ComplexQuery(t *testing.T) {
	query := `SELECT u.id, u.name, COUNT(o.id) as order_count
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.status = 'active'
		GROUP BY u.id, u.name
		ORDER BY order_count DESC
		LIMIT 10`
	result := extractTableName(query)
	assert.Equal(t, "users", result)
}

func TestCleanTableName_Simple(t *testing.T) {
	result := cleanTableName("users")
	assert.Equal(t, "users", result)
}

func TestCleanTableName_WithBackticks(t *testing.T) {
	result := cleanTableName("`users`")
	assert.Equal(t, "users", result)
}

func TestCleanTableName_WithDoubleQuotes(t *testing.T) {
	result := cleanTableName(`"users"`)
	assert.Equal(t, "users", result)
}

func TestCleanTableName_WithSingleQuotes(t *testing.T) {
	result := cleanTableName("'users'")
	assert.Equal(t, "users", result)
}

func TestCleanTableName_WithSchema(t *testing.T) {
	result := cleanTableName("public.users")
	assert.Equal(t, "users", result)
}

func TestCleanTableName_WithSchemaAndQuotes(t *testing.T) {
	result := cleanTableName("`public`.`users`")
	assert.Equal(t, "users", result)
}

func TestCleanTableName_WithTrailingComma(t *testing.T) {
	result := cleanTableName("users,")
	assert.Equal(t, "users", result)
}

func TestCleanTableName_WithTrailingSemicolon(t *testing.T) {
	result := cleanTableName("users;")
	assert.Equal(t, "users", result)
}

func TestCleanTableName_WithTrailingParen(t *testing.T) {
	result := cleanTableName("users)")
	assert.Equal(t, "users", result)
}

func TestCleanTableName_WithOpenParen(t *testing.T) {
	result := cleanTableName("users(")
	assert.Equal(t, "users", result)
}

func TestCleanTableName_ComplexSchema(t *testing.T) {
	result := cleanTableName("catalog.schema.table")
	assert.Equal(t, "table", result)
}

func TestCleanTableName_Empty(t *testing.T) {
	result := cleanTableName("")
	assert.Equal(t, "", result)
}

func TestExtractTableFromKeyword_Simple(t *testing.T) {
	result := extractTableFromKeyword("SELECT * FROM users", "SELECT * FROM USERS", "FROM")
	assert.Equal(t, "users", result)
}

func TestExtractTableFromKeyword_NotFound(t *testing.T) {
	result := extractTableFromKeyword("SELECT 1", "SELECT 1", "FROM")
	assert.Equal(t, "", result)
}

func TestExtractTableFromKeyword_MultipleFrom(t *testing.T) {
	query := "SELECT * FROM users WHERE name = 'FROM'"
	result := extractTableFromKeyword(query, "SELECT * FROM USERS WHERE NAME = 'FROM'", "FROM")
	assert.Equal(t, "users", result)
}
