package service

import (
	"encoding/json"
	"fmt"
)

// MOEX ISS returns: {"securities": {"columns": ["SECID", ...], "data": [["RU000...", ...], ...]}}
// extractRows turns that columnar layout into a slice of column-keyed maps.
func extractRows(data map[string]any, blockName string) []map[string]any {
	block, ok := data[blockName]
	if !ok {
		return nil
	}

	blockMap, ok := block.(map[string]any)
	if !ok {
		return nil
	}

	colsRaw, ok := blockMap["columns"].([]any)
	if !ok {
		return nil
	}
	dataRaw, ok := blockMap["data"].([]any)
	if !ok {
		return nil
	}

	columns := make([]string, len(colsRaw))
	for i, c := range colsRaw {
		columns[i], _ = c.(string)
	}

	rows := make([]map[string]any, 0, len(dataRaw))
	for _, rowRaw := range dataRaw {
		rowArr, ok := rowRaw.([]any)
		if !ok || len(rowArr) != len(columns) {
			continue
		}
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			row[col] = rowArr[i]
		}
		rows = append(rows, row)
	}

	return rows
}

func safeFirst(rows []map[string]any) map[string]any {
	if len(rows) == 0 {
		return nil
	}
	return rows[0]
}

func getString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getFloat(m map[string]any, key string) float64 {
	if m == nil {
		return 0
	}
	if v, ok := m[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return n
		case json.Number:
			f, _ := n.Float64()
			return f
		}
	}
	return 0
}

func getFloatPtr(m map[string]any, key string) *float64 {
	if m == nil {
		return nil
	}
	if v, ok := m[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return &n
		case json.Number:
			f, _ := n.Float64()
			return &f
		}
	}
	return nil
}

func getInt(m map[string]any, key string) int {
	return int(getFloat(m, key))
}

func getIntPtr(m map[string]any, key string) *int {
	f := getFloatPtr(m, key)
	if f == nil {
		return nil
	}
	i := int(*f)
	return &i
}

func getInt64(m map[string]any, key string) int64 {
	return int64(getFloat(m, key))
}

func getInt64Ptr(m map[string]any, key string) *int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	f, ok := v.(float64)
	if !ok {
		return nil
	}
	i := int64(f)
	return &i
}

func safeFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func safeInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
