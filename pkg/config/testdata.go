package config

const testdataRootKey = "testdata"

func testdataFullPath(path string) string {
	if path == "" {
		return testdataRootKey
	}
	return testdataRootKey + "." + path
}

func TestDataGet(path string) interface{} {
	cfg, err := Viper()
	if err != nil {
		return nil
	}
	return cfg.Get(testdataFullPath(path))
}

func TestDataGetString(path string) string {
	cfg, err := Viper()
	if err != nil {
		return ""
	}
	return cfg.GetString(testdataFullPath(path))
}

func TestDataGetInt(path string) int {
	cfg, err := Viper()
	if err != nil {
		return 0
	}
	return cfg.GetInt(testdataFullPath(path))
}

func TestDataGetBool(path string) bool {
	cfg, err := Viper()
	if err != nil {
		return false
	}
	return cfg.GetBool(testdataFullPath(path))
}

func TestDataGetFloat64(path string) float64 {
	cfg, err := Viper()
	if err != nil {
		return 0
	}
	return cfg.GetFloat64(testdataFullPath(path))
}

func TestDataGetStringSlice(path string) []string {
	cfg, err := Viper()
	if err != nil {
		return nil
	}
	return cfg.GetStringSlice(testdataFullPath(path))
}

func TestDataGetStringMap(path string) map[string]interface{} {
	cfg, err := Viper()
	if err != nil {
		return nil
	}
	return cfg.GetStringMap(testdataFullPath(path))
}

func TestDataGetStringMapString(path string) map[string]string {
	cfg, err := Viper()
	if err != nil {
		return nil
	}
	return cfg.GetStringMapString(testdataFullPath(path))
}

func TestDataUnmarshal(path string, out interface{}) error {
	cfg, err := Viper()
	if err != nil {
		return err
	}
	return cfg.UnmarshalKey(testdataFullPath(path), out)
}

func TestDataIsSet(path string) bool {
	cfg, err := Viper()
	if err != nil {
		return false
	}
	return cfg.IsSet(testdataFullPath(path))
}
