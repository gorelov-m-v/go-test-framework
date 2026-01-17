package builder

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"

	dbclient "github.com/gorelov-m-v/go-test-framework/pkg/database/client"
)

func injectDBClient(v *viper.Viper, fieldValue reflect.Value, field reflect.StructField, dbConfigKey, structName string) error {
	debugLog("found tag 'db_config:%s' on field '%s' (type=%s)", dbConfigKey, field.Name, field.Type)

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag db_config:\"%s\" but is not exported", structName, field.Name, dbConfigKey)
	}

	if !v.IsSet(dbConfigKey) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag db_config:\"%s\": config key '%s' not found", structName, field.Name, dbConfigKey, dbConfigKey)
	}

	var dbCfg dbclient.Config
	if err := v.UnmarshalKey(dbConfigKey, &dbCfg); err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag db_config:\"%s\": failed to unmarshal config: %w", structName, field.Name, dbConfigKey, err)
	}

	debugLog("injecting config '%s' into field '%s'", dbConfigKey, field.Name)

	dbClient, err := dbclient.New(dbCfg)
	if err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag db_config:\"%s\": failed to create db client: %w", structName, field.Name, dbConfigKey, err)
	}

	target := fieldValue.Addr().Interface()
	setter, ok := target.(dbclient.DBSetter)
	if !ok {
		return fmt.Errorf("BuildEnv Error: Field '%s' has tag 'db_config' but does not implement 'dbclient.DBSetter'. Please use a Link struct", field.Name)
	}

	setter.SetDB(dbClient)
	debugLog("injected DB client into '%s' via SetDB", field.Name)
	return nil
}
