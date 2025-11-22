package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// ConvertParams converts variadic parameters to string array for D1 API
// Supports basic types (int, float, bool, string), time.Time, and JSON serialization
func ConvertParams(args ...interface{}) ([]string, error) {
	if len(args) == 0 {
		return []string{}, nil
	}

	result := make([]string, len(args))

	for i, arg := range args {
		if arg == nil {
			result[i] = ""
			continue
		}

		switch v := arg.(type) {
		case string:
			result[i] = v
		case int:
			result[i] = fmt.Sprintf("%d", v)
		case int8:
			result[i] = fmt.Sprintf("%d", v)
		case int16:
			result[i] = fmt.Sprintf("%d", v)
		case int32:
			result[i] = fmt.Sprintf("%d", v)
		case int64:
			result[i] = fmt.Sprintf("%d", v)
		case uint:
			result[i] = fmt.Sprintf("%d", v)
		case uint8:
			result[i] = fmt.Sprintf("%d", v)
		case uint16:
			result[i] = fmt.Sprintf("%d", v)
		case uint32:
			result[i] = fmt.Sprintf("%d", v)
		case uint64:
			result[i] = fmt.Sprintf("%d", v)
		case float32:
			result[i] = fmt.Sprintf("%v", v)
		case float64:
			result[i] = fmt.Sprintf("%v", v)
		case bool:
			if v {
				result[i] = "1"
			} else {
				result[i] = "0"
			}
		case time.Time:
			result[i] = v.Format("2006-01-02 15:04:05")
		case []byte:
			result[i] = string(v)
		default:
			// Complex types use JSON serialization
			b, err := json.Marshal(arg)
			if err != nil {
				return nil, fmt.Errorf("无法转换参数 #%d (类型:%T): %v", i, arg, err)
			}
			result[i] = string(b)
		}
	}

	return result, nil
}
