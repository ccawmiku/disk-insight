package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ccawmiku/disk-insight/internal/model"
)

type Config struct {
	Address      string
	DatabasePath string
	WebPath      string
	Roots        []model.RootConfig
}

func FromEnvironment() (Config, error) {
	result := Config{
		Address:      valueOrDefault("DISK_INSIGHT_ADDRESS", ":8080"),
		DatabasePath: valueOrDefault("DISK_INSIGHT_DATABASE", "data/disk-insight.db"),
		WebPath:      valueOrDefault("DISK_INSIGHT_WEB", "web/dist"),
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" && os.Getenv("DISK_INSIGHT_ADDRESS") == "" {
		if _, err := strconv.Atoi(port); err != nil {
			return result, fmt.Errorf("invalid PORT: %w", err)
		}
		result.Address = ":" + port
	}
	rawRoots := strings.TrimSpace(os.Getenv("DISK_INSIGHT_ROOTS"))
	if rawRoots == "" {
		if info, err := os.Stat("/data"); err == nil && info.IsDir() {
			rawRoots = "/data::Data"
		}
	}
	seen := make(map[string]struct{})
	for _, raw := range strings.Split(rawRoots, ";") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		parts := strings.SplitN(raw, "::", 2)
		rootPath, err := filepath.Abs(strings.TrimSpace(parts[0]))
		if err != nil {
			return result, fmt.Errorf("invalid scan root %q: %w", parts[0], err)
		}
		rootPath = filepath.Clean(rootPath)
		key := strings.ToLower(rootPath)
		if _, exists := seen[key]; exists {
			return result, fmt.Errorf("duplicate scan root %q", rootPath)
		}
		seen[key] = struct{}{}
		name := filepath.Base(rootPath)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			name = strings.TrimSpace(parts[1])
		}
		result.Roots = append(result.Roots, model.RootConfig{Name: name, Path: rootPath, Enabled: true})
	}
	return result, nil
}

func valueOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
