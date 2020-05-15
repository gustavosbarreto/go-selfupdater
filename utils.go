package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"
)

func GetCurrentContainerID() (string, error) {
	const idLength = 64

	f, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return "", scanner.Err()
	}

	line := scanner.Text()
	return line[len(line)-64:], nil
}

func ReplaceOrAppendEnvValues(defaults, overrides []string) []string {
	cache := make(map[string]int, len(defaults))
	for i, e := range defaults {
		index := strings.Index(e, "=")
		cache[e[:index]] = i
	}

	for _, value := range overrides {
		// Values w/o = means they want this env to be removed/unset.
		index := strings.Index(value, "=")
		if index < 0 {
			// no "=" in value
			if i, exists := cache[value]; exists {
				defaults[i] = "" // Used to indicate it should be removed
			}
			continue
		}

		if i, exists := cache[value[:index]]; exists {
			defaults[i] = value
		} else {
			defaults = append(defaults, value)
		}
	}

	// Now remove all entries that we want to "unset"
	for i := 0; i < len(defaults); i++ {
		if defaults[i] == "" {
			defaults = append(defaults[:i], defaults[i+1:]...)
			i--
		}
	}

	return defaults
}

func IsUpdateAvailable(v1, v2 string) (bool, error) {
	c, err := semver.NewConstraint(fmt.Sprintf("> %s", v1))
	if err != nil {
		return false, err
	}

	v, err := semver.NewVersion(v2)
	if err != nil {
		return false, err
	}

	return c.Check(v), nil
}
