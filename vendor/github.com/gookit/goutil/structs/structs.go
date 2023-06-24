// Package structs Provide some extends util functions for struct. eg: tag parse, struct init, value set
package structs

// IsExported field name on struct
func IsExported(name string) bool {
	return name[0] >= 'A' && name[0] <= 'Z'
}

// IsUnexported field name on struct
func IsUnexported(name string) bool {
	if name[0] == '_' {
		return true
	}
	return name[0] >= 'a' && name[0] <= 'z'
}
