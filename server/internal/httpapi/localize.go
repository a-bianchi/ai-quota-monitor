package httpapi

import (
	"ai-quota-monitor/server/internal/usage"
	"regexp"
	"strings"
	"time"
)

// defaultZone applies when the request has no X-Timezone header.
const defaultZone = "America/Argentina/Buenos_Aires"

// Providers produce quota names in Spanish; these are their English forms.
var english = map[string]string{
	"Semanal":          "Weekly",
	"Mensual":          "Monthly",
	"Saldo prepago":    "Prepaid balance",
	"Créditos":         "Credits",
	"Uso incluido":     "Included usage",
	"Modelos API":      "API models",
	"On-demand equipo": "Team on-demand",
}
var englishUnits = map[string]string{"días": "days", "horas": "hours", "minutos": "minutes"}
var window = regexp.MustCompile(`^(\d+) (días|horas|minutos)( .+)?$`)

// language picks the first supported language of Accept-Language; Spanish is the default.
func language(header string) string {
	for _, tag := range strings.Split(header, ",") {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "en" || strings.HasPrefix(tag, "en-") || strings.HasPrefix(tag, "en;") {
			return "en"
		}
		if tag == "es" || strings.HasPrefix(tag, "es-") || strings.HasPrefix(tag, "es;") {
			return "es"
		}
	}
	return "es"
}

// translate leaves names it does not know untouched, such as those sent by collectors.
func translate(name string) string {
	parts := strings.Split(name, " · ")
	for i, p := range parts {
		if en, ok := english[p]; ok {
			parts[i] = en
		} else if m := window.FindStringSubmatch(p); m != nil {
			unit := englishUnits[m[2]]
			if m[1] == "1" {
				unit = strings.TrimSuffix(unit, "s")
			}
			parts[i] = m[1] + " " + unit + m[3]
		}
	}
	return strings.Join(parts, " · ")
}

// localize returns a copy of the quotas with names in lang, reset times in loc and the usage level.
func localize(quotas []usage.Quota, lang string, loc *time.Location) []usage.Quota {
	out := make([]usage.Quota, len(quotas))
	for i, q := range quotas {
		if lang == "en" {
			q.Name = translate(q.Name)
		}
		if q.ResetsAt != nil {
			t := q.ResetsAt.In(loc)
			q.ResetsAt = &t
		}
		q.UsageLevel = q.Level()
		out[i] = q
	}
	return out
}
