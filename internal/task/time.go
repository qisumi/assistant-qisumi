package task

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// FlexibleTime 是一个支持多种日期格式的自定义时间类型
// 用于处理 LLM 返回的不同日期格式
type FlexibleTime struct {
	Time time.Time
}

// 支持的日期格式列表
var timeFormats = []string{
	time.RFC3339,                // "2006-01-02T15:04:05Z07:00"
	"2006-01-02T15:04:05Z",     // "2006-01-02T15:04:05Z"
	"2006-01-02T15:04:05",      // "2006-01-02T15:04:05"
	"2006-01-02 15:04:05",      // "2006-01-02 15:04:05"
	"2006-01-02",                // "2006-01-02" (仅日期)
	"2006/01/02",                // "2006/01/02"
	"2006年01月02日",             // "2006年01月02日"
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	// 移除引号
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		ft.Time = time.Time{}
		return nil
	}

	// 尝试用各种格式解析
	for _, format := range timeFormats {
		t, err := time.Parse(format, s)
		if err == nil {
			ft.Time = t
			return nil
		}
	}

	return fmt.Errorf("无法解析时间: %q", s)
}

// MarshalJSON 实现 json.Marshaler 接口
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	if ft.IsZero() {
		return []byte("null"), nil
	}
	// 使用 RFC3339 格式输出
	return []byte(`"` + ft.Time.Format(time.RFC3339) + `"`), nil
}

// Value 实现 driver.Valuer 接口，用于数据库存储
func (ft *FlexibleTime) Value() (driver.Value, error) {
	if ft == nil || ft.IsZero() {
		return nil, nil
	}
	return ft.Time, nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (ft *FlexibleTime) Scan(value any) error {
	if value == nil {
		ft.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ft.Time = v
	case []byte:
		t, err := time.Parse(time.RFC3339, string(v))
		if err != nil {
			return err
		}
		ft.Time = t
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return err
		}
		ft.Time = t
	default:
		return fmt.Errorf("无法将 %T 转换为 FlexibleTime", value)
	}
	return nil
}

// ToTime 返回标准的 time.Time
func (ft FlexibleTime) ToTime() time.Time {
	return ft.Time
}

// IsZero 检查时间是否为零值
func (ft FlexibleTime) IsZero() bool {
	return ft.Time.IsZero()
}

// ParseFlexibleTime 从字符串解析时间，支持多种格式
func ParseFlexibleTime(s string) (*FlexibleTime, error) {
	if s == "" {
		return nil, nil
	}
	var ft FlexibleTime
	if err := ft.UnmarshalJSON([]byte(`"` + s + `"`)); err != nil {
		return nil, err
	}
	return &ft, nil
}
