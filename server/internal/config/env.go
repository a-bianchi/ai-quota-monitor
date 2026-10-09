package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// LoadEnv loads a local dotenv file without executing shell commands. Existing
// environment variables (including empty values) take precedence.
func LoadEnv(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	valid := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	pending := map[string]string{}
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if !ok || !valid.MatchString(key) {
			return fmt.Errorf("invalid dotenv assignment at line %d", lineNo)
		}
		if strings.HasPrefix(value, "\"") {
			decoded, e := strconv.Unquote(value)
			if e != nil {
				return fmt.Errorf("invalid dotenv quoting at line %d", lineNo)
			}
			value = decoded
		} else if strings.HasPrefix(value, "'") {
			if len(value) < 2 || !strings.HasSuffix(value, "'") {
				return fmt.Errorf("invalid dotenv quoting at line %d", lineNo)
			}
			value = value[1 : len(value)-1]
		}
		pending[key] = value
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	for key, value := range pending {
		if _, exists := os.LookupEnv(key); !exists {
			if err = os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}
