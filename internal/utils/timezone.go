package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// TimezoneOption описывает одну опцию часового пояса для select.
type TimezoneOption struct {
	Value string
	Label string
}

// TimezoneGroup описывает группу часовых поясов (например, Europe, Asia).
type TimezoneGroup struct {
	Name    string
	Options []TimezoneOption
}

var (
	timezoneGroupsOnce  sync.Once
	timezoneGroupsCache []TimezoneGroup
)

// GetTimezoneGroups возвращает сгруппированный список часовых поясов IANA с UTC смещением.
//
// Формат label: "Europe/Moscow (UTC+03:00)".
// Результат кэшируется и переиспользуется между запросами.
func GetTimezoneGroups() []TimezoneGroup {
	timezoneGroupsOnce.Do(func() {
		timezoneGroupsCache = buildTimezoneGroups(time.Now().UTC())
		if len(timezoneGroupsCache) == 0 {
			timezoneGroupsCache = []TimezoneGroup{
				{
					Name: "UTC",
					Options: []TimezoneOption{
						{Value: "UTC", Label: "UTC (UTC+00:00)"},
					},
				},
			}
		}
	})

	return timezoneGroupsCache
}

func buildTimezoneGroups(baseTime time.Time) []TimezoneGroup {
	zones := collectTimezones()
	if len(zones) == 0 {
		return nil
	}

	grouped := make(map[string][]TimezoneOption)

	for _, zone := range zones {
		loc, err := time.LoadLocation(zone)
		if err != nil {
			continue
		}

		_, offset := baseTime.In(loc).Zone()
		label := fmt.Sprintf("%s (UTC%s)", zone, formatUTCOffset(offset))
		groupName := timezoneGroupName(zone)

		grouped[groupName] = append(grouped[groupName], TimezoneOption{
			Value: zone,
			Label: label,
		})
	}

	if len(grouped) == 0 {
		return nil
	}

	for groupName := range grouped {
		sort.Slice(grouped[groupName], func(i, j int) bool {
			return grouped[groupName][i].Value < grouped[groupName][j].Value
		})
	}

	groupNames := make([]string, 0, len(grouped))
	for groupName := range grouped {
		groupNames = append(groupNames, groupName)
	}

	groupPriority := map[string]int{
		"UTC":        0,
		"Africa":     1,
		"America":    2,
		"Antarctica": 3,
		"Arctic":     4,
		"Asia":       5,
		"Atlantic":   6,
		"Australia":  7,
		"Europe":     8,
		"Indian":     9,
		"Pacific":    10,
		"Etc":        11,
		"Other":      99,
	}

	sort.Slice(groupNames, func(i, j int) bool {
		pi, okI := groupPriority[groupNames[i]]
		if !okI {
			pi = 50
		}
		pj, okJ := groupPriority[groupNames[j]]
		if !okJ {
			pj = 50
		}
		if pi != pj {
			return pi < pj
		}
		return groupNames[i] < groupNames[j]
	})

	result := make([]TimezoneGroup, 0, len(groupNames))
	for _, groupName := range groupNames {
		result = append(result, TimezoneGroup{
			Name:    groupName,
			Options: grouped[groupName],
		})
	}

	return result
}

func collectTimezones() []string {
	paths := []string{
		"/usr/share/zoneinfo",
	}

	seen := make(map[string]struct{})

	for _, root := range paths {
		if _, err := os.Stat(root); err != nil {
			continue
		}

		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			rel, err := filepath.Rel(root, path)
			if err != nil || rel == "." {
				return nil
			}

			rel = filepath.ToSlash(rel)

			if d.IsDir() {
				if shouldSkipTimezoneDir(rel) {
					return filepath.SkipDir
				}
				return nil
			}

			if shouldSkipTimezoneFile(rel) {
				return nil
			}

			seen[rel] = struct{}{}
			return nil
		})
	}

	// Гарантируем наличие UTC
	seen["UTC"] = struct{}{}

	zones := make([]string, 0, len(seen))
	for zone := range seen {
		zones = append(zones, zone)
	}

	sort.Strings(zones)
	return zones
}

func shouldSkipTimezoneDir(rel string) bool {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return false
	}

	skipPrefixes := []string{"posix", "right"}
	for _, prefix := range skipPrefixes {
		if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
			return true
		}
	}

	return false
}

func shouldSkipTimezoneFile(rel string) bool {
	base := filepath.Base(rel)

	if strings.HasPrefix(base, ".") {
		return true
	}

	skipNames := map[string]struct{}{
		"localtime":    {},
		"posixrules":   {},
		"leapseconds":  {},
		"tzdata.zi":    {},
		"zone.tab":     {},
		"zone1970.tab": {},
		"iso3166.tab":  {},
		"+VERSION":     {},
	}

	if _, exists := skipNames[base]; exists {
		return true
	}

	if strings.Contains(base, ".tab") {
		return true
	}

	return false
}

func timezoneGroupName(zone string) string {
	if zone == "UTC" {
		return "UTC"
	}

	parts := strings.Split(zone, "/")
	if len(parts) > 1 && strings.TrimSpace(parts[0]) != "" {
		return parts[0]
	}

	return "Other"
}

func formatUTCOffset(seconds int) string {
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}
