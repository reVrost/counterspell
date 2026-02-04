package codex

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
)

func serializeConfigOverrides(configOverrides ConfigObject) ([]string, error) {
	overrides := []string{}
	if configOverrides == nil {
		return overrides, nil
	}
	if err := flattenConfigOverrides(configOverrides, "", &overrides); err != nil {
		return nil, err
	}
	return overrides, nil
}

func flattenConfigOverrides(value ConfigValue, prefix string, overrides *[]string) error {
	if !isPlainObject(value) {
		if prefix == "" {
			return fmt.Errorf("Codex config overrides must be a plain object")
		}
		*overrides = append(*overrides, fmt.Sprintf("%s=%s", prefix, toTomlValue(value, prefix)))
		return nil
	}

	valueMap, err := asStringMap(value)
	if err != nil {
		return err
	}
	if prefix == "" && len(valueMap) == 0 {
		return nil
	}
	if prefix != "" && len(valueMap) == 0 {
		*overrides = append(*overrides, fmt.Sprintf("%s={}", prefix))
		return nil
	}

	for key, child := range valueMap {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("Codex config override keys must be non-empty strings")
		}
		if child == nil {
			continue
		}
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		if isPlainObject(child) {
			if err := flattenConfigOverrides(child, path, overrides); err != nil {
				return err
			}
		} else {
			*overrides = append(*overrides, fmt.Sprintf("%s=%s", path, toTomlValue(child, path)))
		}
	}
	return nil
}

func toTomlValue(value ConfigValue, path string) string {
	switch v := value.(type) {
	case string:
		encoded, _ := json.Marshal(v)
		return string(encoded)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			panic(fmt.Sprintf("Codex config override at %s must be a finite number", path))
		}
		return fmt.Sprintf("%v", v)
	case float32:
		if math.IsInf(float64(v), 0) || math.IsNaN(float64(v)) {
			panic(fmt.Sprintf("Codex config override at %s must be a finite number", path))
		}
		return fmt.Sprintf("%v", v)
	case []any:
		parts := make([]string, 0, len(v))
		for idx, item := range v {
			parts = append(parts, toTomlValue(item, fmt.Sprintf("%s[%d]", path, idx)))
		}
		return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
	case []ConfigValue:
		parts := make([]string, 0, len(v))
		for idx, item := range v {
			parts = append(parts, toTomlValue(item, fmt.Sprintf("%s[%d]", path, idx)))
		}
		return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
	default:
		if isPlainObject(v) {
			obj, _ := asStringMap(v)
			parts := make([]string, 0, len(obj))
			for key, child := range obj {
				if strings.TrimSpace(key) == "" {
					panic("Codex config override keys must be non-empty strings")
				}
				if child == nil {
					continue
				}
				parts = append(parts, fmt.Sprintf("%s = %s", formatTomlKey(key), toTomlValue(child, path+"."+key)))
			}
			return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
		}
		if value == nil {
			panic(fmt.Sprintf("Codex config override at %s cannot be null", path))
		}
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Pointer && val.IsNil() {
			panic(fmt.Sprintf("Codex config override at %s cannot be null", path))
		}
		panic(fmt.Sprintf("Unsupported Codex config override value at %s: %T", path, value))
	}
}

var tomlBareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func formatTomlKey(key string) string {
	if tomlBareKey.MatchString(key) {
		return key
	}
	encoded, _ := json.Marshal(key)
	return string(encoded)
}

func isPlainObject(value any) bool {
	if value == nil {
		return false
	}
	if _, ok := value.(map[string]any); ok {
		return true
	}
	if _, ok := value.(ConfigObject); ok {
		return true
	}
	val := reflect.ValueOf(value)
	return val.Kind() == reflect.Map && val.Type().Key().Kind() == reflect.String
}

func asStringMap(value any) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("nil config override")
	}
	if m, ok := value.(map[string]any); ok {
		return m, nil
	}
	if m, ok := value.(ConfigObject); ok {
		result := make(map[string]any, len(m))
		for key, child := range m {
			result[key] = child
		}
		return result, nil
	}
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("Codex config overrides must be a plain object")
	}
	result := make(map[string]any, val.Len())
	for _, key := range val.MapKeys() {
		if key.Kind() != reflect.String {
			return nil, fmt.Errorf("Codex config override keys must be strings")
		}
		result[key.String()] = val.MapIndex(key).Interface()
	}
	return result, nil
}
