package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StringSlice []string

// Scan implements the sql.Scanner interface
func (ss *StringSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSON value: %v", value)
	}
	
	// Try to unmarshal as number first
	var num int
	if err := json.Unmarshal(bytes, &num); err == nil {
		// If it's a number, create an empty slice
		*ss = make([]string, 0)
		return nil
	}

	// If not a number, try to unmarshal as array
	var links []string
	if err := json.Unmarshal(bytes, &links); err != nil {
		return err
	}
	*ss = links
	return nil
}

// Value implements the driver.Valuer interface
func (ss StringSlice) Value() (driver.Value, error) {
	return json.Marshal(ss)
}
