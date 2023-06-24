package structs

import "fmt"

// Aliases implemented a simple string alias map.
type Aliases struct {
	mapping map[string]string
	// Checker custom add alias name checker func
	Checker func(alias string) // should return bool OR error ??
}

// NewAliases create
func NewAliases(checker func(alias string)) *Aliases {
	return &Aliases{Checker: checker}
}

// AddAlias to the Aliases
func (as *Aliases) AddAlias(real, alias string) {
	if as.mapping == nil {
		as.mapping = make(map[string]string)
	}

	if as.Checker != nil {
		as.Checker(alias)
	}

	if rn, ok := as.mapping[alias]; ok {
		panic(fmt.Sprintf("The alias '%s' is already used by '%s'", alias, rn))
	}
	as.mapping[alias] = real
}

// AddAliases to the Aliases
func (as *Aliases) AddAliases(real string, aliases []string) {
	for _, a := range aliases {
		as.AddAlias(real, a)
	}
}

// AddAliasMap to the Aliases
func (as *Aliases) AddAliasMap(alias2real map[string]string) {
	for a, r := range alias2real {
		as.AddAlias(r, a)
	}
}

// HasAlias in the Aliases
func (as *Aliases) HasAlias(alias string) bool {
	if _, ok := as.mapping[alias]; ok {
		return true
	}
	return false
}

// ResolveAlias by given name.
func (as *Aliases) ResolveAlias(alias string) string {
	if name, ok := as.mapping[alias]; ok {
		return name
	}
	return alias
}

// Mapping get all aliases mapping
func (as *Aliases) Mapping() map[string]string {
	return as.mapping
}
