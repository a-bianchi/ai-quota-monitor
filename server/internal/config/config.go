package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
)

type Account struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Provider      string `json:"provider"`
	Scope         string `json:"scope"`
	Source        string `json:"source"`
	CredentialEnv string `json:"credentialEnv,omitempty"`
}
type Config struct {
	Accounts []Account `json:"accounts"`
}

func Load(path string) (Config, error) {
	var c Config
	f, e := os.Open(path)
	if e != nil {
		return c, e
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 1<<20))
	d.DisallowUnknownFields()
	if e = d.Decode(&c); e != nil {
		return c, e
	}
	if e = d.Decode(new(any)); e != io.EOF {
		return c, fmt.Errorf("configuration must contain one JSON object")
	}
	if len(c.Accounts) == 0 || len(c.Accounts) > 64 {
		return c, fmt.Errorf("configure 1 to 64 accounts")
	}
	seen := map[string]bool{}
	validID := regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,79}$`)
	for _, a := range c.Accounts {
		if !validID.MatchString(a.ID) || seen[a.ID] || a.Name == "" || len(a.Name) > 120 {
			return c, fmt.Errorf("invalid or duplicate account ID/name")
		}
		seen[a.ID] = true
		if a.Scope != "personal" && a.Scope != "work" {
			return c, fmt.Errorf("%s: scope must be personal or work", a.ID)
		}
		switch a.Provider {
		case "openai", "cursor", "kiro", "openrouter", "anthropic":
		default:
			return c, fmt.Errorf("%s: unknown provider", a.ID)
		}
		switch a.Source {
		case "mock", "collector":
		case "direct":
			switch a.Provider {
			case "openai", "anthropic":
			case "openrouter", "cursor":
				if a.CredentialEnv == "" {
					return c, fmt.Errorf("%s: direct %s requires credentialEnv", a.ID, a.Provider)
				}
			default:
				return c, fmt.Errorf("%s: direct supports openrouter, cursor, openai or anthropic", a.ID)
			}
		default:
			return c, fmt.Errorf("%s: source must be mock, direct or collector", a.ID)
		}
	}
	return c, nil
}
