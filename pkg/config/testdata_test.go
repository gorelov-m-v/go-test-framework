package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestdataFullPath_EmptyPath(t *testing.T) {
	result := testdataFullPath("")

	assert.Equal(t, "testdata", result)
}

func TestTestdataFullPath_SimplePath(t *testing.T) {
	result := testdataFullPath("users")

	assert.Equal(t, "testdata.users", result)
}

func TestTestdataFullPath_NestedPath(t *testing.T) {
	result := testdataFullPath("users.admin.email")

	assert.Equal(t, "testdata.users.admin.email", result)
}

func TestTestdataFullPath_SingleField(t *testing.T) {
	result := testdataFullPath("timeout")

	assert.Equal(t, "testdata.timeout", result)
}
