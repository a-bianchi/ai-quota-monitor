package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	t.Setenv("QUOTA_TEST_EXISTING", "environment")
	for _, key := range []string{"QUOTA_TEST_VALUE", "QUOTA_TEST_LITERAL"} {
		old, ok := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if ok {
				os.Setenv(key, old)
			} else {
				os.Unsetenv(key)
			}
		})
	}
	contents := `# local settings
QUOTA_TEST_EXISTING=file
QUOTA_TEST_VALUE="hello world"
export QUOTA_TEST_LITERAL='$(echo must-not-execute)'
`
	if e := os.WriteFile(path, []byte(contents), 0600); e != nil {
		t.Fatal(e)
	}
	if e := LoadEnv(path); e != nil {
		t.Fatal(e)
	}
	if os.Getenv("QUOTA_TEST_EXISTING") != "environment" {
		t.Fatal("overrode environment")
	}
	if os.Getenv("QUOTA_TEST_VALUE") != "hello world" {
		t.Fatal("quoted value lost")
	}
	if os.Getenv("QUOTA_TEST_LITERAL") != "$(echo must-not-execute)" {
		t.Fatal("literal was changed")
	}
	if e := LoadEnv(filepath.Join(dir, "missing")); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(path, []byte("QUOTA_TEST_VALUE=changed\ninvalid line"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := LoadEnv(path); e == nil {
		t.Fatal("invalid dotenv accepted")
	}
	if os.Getenv("QUOTA_TEST_VALUE") != "hello world" {
		t.Fatal("partial file applied")
	}
}
