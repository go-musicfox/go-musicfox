package player

import "time"

type Music struct {
    ID       int
    Name     string
    Artist   string
    Album    string
    Singer   string
    Duration time.Duration
}
