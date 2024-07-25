package optional

import (
	"database/sql/driver"
)

// Scan assigns a value from a database driver.
// This method is required from database/sql.Scanner interface.
func (o *Option[T]) Scan(src any) error {
	if src == nil {
		*o = None[T]()
		return nil
	}

	var v T
	err := sqlConvertAssign(&v, src)
	if err != nil {
		return err
	}

	*o = Some[T](v)
	return nil
}

// Value returns a driver Value.
// This method is required from database/sql/driver.Valuer interface.
func (o Option[T]) Value() (driver.Value, error) {
	if o.IsNone() {
		return nil, nil
	}
	return driver.DefaultParameterConverter.ConvertValue(o.Unwrap())
}
