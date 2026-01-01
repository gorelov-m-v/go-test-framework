package dsl

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-test-framework/pkg/database/client"
)

type TestQueryUser struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	IsActive int64  `db:"is_active"`
}

func TestNewQuery_CreatesQueryWithCorrectTypes(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbClient := &client.Client{DB: db}

	_ = NewQuery[TestQueryUser](nil, dbClient)
	_ = NewQuery[any](nil, dbClient)
	_ = NewQuery[struct {
		Count int `db:"count"`
	}](nil, dbClient)

	assert.True(t, true, "Query creation with generics works")
}

func TestQuery_SQL_SetsQueryAndArgs(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbClient := &client.Client{DB: db}
	query := NewQuery[TestQueryUser](nil, dbClient)

	query.SQL("SELECT * FROM users WHERE id = ?", 123)

	assert.Equal(t, "SELECT * FROM users WHERE id = ?", query.sql)
	assert.Equal(t, []any{123}, query.args)
}

func TestQuery_WithContext_SetsCustomContext(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbClient := &client.Client{DB: db}
	query := NewQuery[TestQueryUser](nil, dbClient)

	assert.NotNil(t, query.ctx)

	result := query.SQL("SELECT * FROM users", 1).WithContext(query.ctx)
	assert.NotNil(t, result)
}

func TestQuery_ExpectationsChaining(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbClient := &client.Client{DB: db}
	query := NewQuery[TestQueryUser](nil, dbClient)

	result := query.
		SQL("SELECT * FROM users WHERE id = ?", 1).
		ExpectFound().
		ExpectColumnEquals("username", "test").
		ExpectColumnTrue("is_active")

	assert.Len(t, result.expectations, 3)
}

func TestQuery_MultipleExpectationsAddedCorrectly(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbClient := &client.Client{DB: db}
	query := NewQuery[TestQueryUser](nil, dbClient)

	query.
		ExpectFound().
		ExpectColumnEquals("id", 1).
		ExpectColumnNotEmpty("username").
		ExpectColumnIsNull("deleted_at").
		ExpectColumnIsNotNull("created_at").
		ExpectColumnTrue("is_active").
		ExpectColumnFalse("is_deleted")

	assert.Len(t, query.expectations, 7)
}
