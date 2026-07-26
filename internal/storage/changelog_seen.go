package storage

import "github.com/go-musicfox/go-musicfox/internal/types"

// ChangelogSeen tracks the last version for which the user has seen the
// changelog popup. Stored as a JSON-serialized KV model in BoltDB.
type ChangelogSeen struct {
	Version string `json:"version"`
}

func (c ChangelogSeen) GetDbName() string   { return types.AppDBName }
func (c ChangelogSeen) GetTableName() string { return "default_bucket" }
func (c ChangelogSeen) GetKey() string      { return "changelog_seen" }
