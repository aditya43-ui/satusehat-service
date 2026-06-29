package custom

import (
	"strings"
	"time"
)

// CustomTime membungkus time.Time untuk mendukung format custom dan RFC3339
type CustomTime struct {
	time.Time
}

func (c *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}

	// Coba parse dengan format "2006-01-02 15:04:05" dan set timezone
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		loc, _ := time.LoadLocation("Asia/Jakarta")
		c.Time = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
		return nil
	}

	// Fallback ke format RFC3339 bawaan
	t, err = time.Parse(time.RFC3339, s)
	c.Time = t
	return err
}
