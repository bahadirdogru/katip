package tarz

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Rule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	Description string `json:"description"`
}

type Profile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Rules       []Rule `json:"rules"`
}

func ProfilesDir() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "Katip", "profiles")
}

func ListProfiles() ([]string, error) {
	dir := ProfilesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tarz") {
			names = append(names, strings.TrimSuffix(e.Name(), ".tarz"))
		}
	}
	return names, nil
}

func LoadProfile(name string) (*Profile, error) {
	path := filepath.Join(ProfilesDir(), name+".tarz")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profil okunamadı: %w", err)
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("profil ayrıştırılamadı: %w", err)
	}
	return &p, nil
}

func SaveProfile(name string, profile Profile) error {
	dir := ProfilesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".tarz"), data, 0644)
}

type StyleViolation struct {
	RuleID      string `json:"ruleId"`
	RuleName    string `json:"ruleName"`
	Original    string `json:"original"`
	Suggestion  string `json:"suggestion"`
	Description string `json:"description"`
}

func CheckText(profile *Profile, text string) []StyleViolation {
	if profile == nil {
		return nil
	}
	var violations []StyleViolation
	lower := strings.ToLower(text)
	for _, rule := range profile.Rules {
		pattern := strings.ToLower(rule.Pattern)
		if pattern != "" && strings.Contains(lower, pattern) {
			violations = append(violations, StyleViolation{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Original:    rule.Pattern,
				Suggestion:  rule.Replacement,
				Description: rule.Description,
			})
		}
	}
	return violations
}

func DefaultProfile() Profile {
	return Profile{
		Name:        "varsayilan",
		Description: "Temel Türkçe yayınevi kuralları",
		Rules: []Rule{
			{ID: "r1", Name: "vb. kısaltması", Pattern: "ve benzeri", Replacement: "vb.", Description: "Akademik metinlerde 've benzeri' yerine 'vb.' kullanılmalı"},
			{ID: "r2", Name: "yani kısaltması", Pattern: "yani", Replacement: "d.h.", Description: "Resmi metinlerde 'yani' yerine 'd.h.' tercih edilebilir"},
		},
	}
}

func EnsureDefaultProfile() error {
	dir := ProfilesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "varsayilan.tarz")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return SaveProfile("varsayilan", DefaultProfile())
}
