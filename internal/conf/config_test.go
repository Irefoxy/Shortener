package conf

import (
	"os"
	"testing"
)

const (
	HostAddressName    = "HostAddress"
	TargetAddressName  = "TargetAddress"
	DatabaseStringName = "DatabaseString"
	FileLocationName   = "FileLocation"
	dbEnvName          = "DATABASE_DSN"
	fileEncName        = "FILE_STORAGE_PATH"
	baseEnvName        = "BASE_URL"
	severEnvName       = "SERVER_ADDRESS"
)

func TestConfig(t *testing.T) {
	var tests = []struct {
		name     string
		argv     []string
		envName  []string
		envVal   []string
		expected map[string]string
	}{
		{"Empty", []string{"config_test.go"}, []string{}, []string{},
			map[string]string{HostAddressName: "localhost:8888", TargetAddressName: "localhost:8888", DatabaseStringName: "", FileLocationName: ""}},
		{"EmptyWithDeclarations", []string{"config_test_go", "a"}, []string{baseEnvName}, []string{""},
			map[string]string{HostAddressName: "localhost:8888", TargetAddressName: "localhost:8888", DatabaseStringName: "", FileLocationName: ""}},
		{"OK", []string{"config_test_go", "-a", "localhost:1"}, []string{baseEnvName}, []string{"localhost:2"},
			map[string]string{HostAddressName: "localhost:1", TargetAddressName: "localhost:2", DatabaseStringName: "", FileLocationName: ""}},
		{"OK", []string{"config_test_go", "-a", "localhost:1"}, []string{severEnvName}, []string{"localhost:2"},
			map[string]string{HostAddressName: "localhost:2", TargetAddressName: "localhost:8888", DatabaseStringName: "", FileLocationName: ""}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := New()
			setEnvs(t, test.envName, test.envVal)
			cfg.Parse(test.argv[0], test.argv[1:])
			clearEnvs(t, test.envName)

			if *cfg.GetApiConf().HostAddress != test.expected[HostAddressName] {
				t.Errorf("Expected %s: %s, but got %s", HostAddressName, test.expected[HostAddressName], *cfg.GetApiConf().HostAddress)
			}
			if *cfg.GetApiConf().TargetAddress != test.expected[TargetAddressName] {
				t.Errorf("Expected %s: %s, but got %s", TargetAddressName, test.expected[TargetAddressName], *cfg.GetApiConf().TargetAddress)
			}
			if cfg.GetDatabaseString() != test.expected[DatabaseStringName] {
				t.Errorf("Expected %s: %s, but got %s", DatabaseStringName, test.expected[DatabaseStringName], cfg.GetDatabaseString())
			}
			if cfg.GetFileLocation() != test.expected[FileLocationName] {
				t.Errorf("Expected %s: %s, but got %s", FileLocationName, test.expected[FileLocationName], cfg.GetFileLocation())
			}
		})
	}
}

func setEnvs(t *testing.T, envName, envVal []string) {
	for i, name := range envName {
		err := os.Setenv(name, envVal[i])
		if err != nil {
			t.Fatal(err)
		}
	}
}

func clearEnvs(t *testing.T, envName []string) {
	for _, name := range envName {
		err := os.Unsetenv(name)
		if err != nil {
			t.Fatal(err)
		}
	}
}
