package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

type Config struct {
	Driver          string             `mapstructure:"driver" yaml:"driver" json:"driver"`
	DSN             string             `mapstructure:"dsn" yaml:"dsn" json:"dsn"`
	MaxOpenConns    int                `mapstructure:"maxOpenConns" yaml:"maxOpenConns" json:"maxOpenConns"`
	MaxIdleConns    int                `mapstructure:"maxIdleConns" yaml:"maxIdleConns" json:"maxIdleConns"`
	ConnMaxLifetime time.Duration      `mapstructure:"connMaxLifetime" yaml:"connMaxLifetime" json:"connMaxLifetime"`
	MaskColumns     string             `mapstructure:"maskColumns" yaml:"maskColumns" json:"maskColumns"`
	Schemas         map[string]string  `mapstructure:"schemas" yaml:"schemas" json:"schemas"`
	AsyncConfig     config.AsyncConfig `mapstructure:"async" yaml:"async" json:"async"`
}

type Client struct {
	DB          *sqlx.DB
	AsyncConfig config.AsyncConfig
	maskColumns []string
	schemas     map[string]string
}

func New(cfg Config) (*Client, error) {
	if cfg.Driver == "" {
		return nil, fmt.Errorf("driver is required: must be 'mysql' or 'postgres'")
	}

	if cfg.Driver != "mysql" && cfg.Driver != "postgres" {
		return nil, fmt.Errorf("unsupported driver '%s': must be 'mysql' or 'postgres'", cfg.Driver)
	}

	db, err := sqlx.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	var maskColumns []string
	if cfg.MaskColumns != "" {
		parts := strings.Split(cfg.MaskColumns, ",")
		for _, part := range parts {
			trimmed := strings.ToLower(strings.TrimSpace(part))
			if trimmed != "" {
				maskColumns = append(maskColumns, trimmed)
			}
		}
	}

	asyncCfg := cfg.AsyncConfig.WithDefaults()

	return &Client{DB: db, AsyncConfig: asyncCfg, maskColumns: maskColumns, schemas: cfg.Schemas}, nil
}

func (c *Client) ShouldMaskColumn(name string) bool {
	if len(c.maskColumns) == 0 {
		return false
	}
	key := strings.ToLower(strings.TrimSpace(name))
	for _, col := range c.maskColumns {
		if strings.Contains(key, col) {
			return true
		}
	}
	return false
}

func (c *Client) Close() error {
	return c.DB.Close()
}

func (c *Client) Schema(alias string) string {
	if c.schemas == nil {
		return alias
	}
	if schema, ok := c.schemas[alias]; ok {
		return schema
	}
	return alias
}
