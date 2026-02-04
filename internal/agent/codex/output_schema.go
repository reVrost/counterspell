package codex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

type OutputSchemaFile struct {
	SchemaPath string
	Cleanup    func() error
}

func CreateOutputSchemaFile(schema any) (OutputSchemaFile, error) {
	if schema == nil {
		return OutputSchemaFile{Cleanup: func() error { return nil }}, nil
	}
	if !isJSONObject(schema) {
		return OutputSchemaFile{}, fmt.Errorf("outputSchema must be a plain JSON object")
	}

	schemaDir, err := os.MkdirTemp("", "codex-output-schema-")
	if err != nil {
		return OutputSchemaFile{}, err
	}
	schemaPath := filepath.Join(schemaDir, "schema.json")
	cleanup := func() error {
		return os.RemoveAll(schemaDir)
	}

	data, err := json.Marshal(schema)
	if err != nil {
		_ = cleanup()
		return OutputSchemaFile{}, err
	}
	if err := os.WriteFile(schemaPath, data, 0o600); err != nil {
		_ = cleanup()
		return OutputSchemaFile{}, err
	}
	return OutputSchemaFile{SchemaPath: schemaPath, Cleanup: cleanup}, nil
}

func isJSONObject(value any) bool {
	if value == nil {
		return false
	}
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Map {
		return false
	}
	return val.Type().Key().Kind() == reflect.String
}
