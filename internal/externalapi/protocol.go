package externalapi

// PlayerState mirrors SPlayer-Next's externally visible playback states.
type PlayerState string

const (
	StateIdle    PlayerState = "idle"
	StateLoading PlayerState = "loading"
	StatePlaying PlayerState = "playing"
	StatePaused  PlayerState = "paused"
	StateStopped PlayerState = "stopped"
)

type Artist struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

type Album struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Cover string `json:"cover,omitempty"`
}

type Track struct {
	ID            string   `json:"id"`
	Source        string   `json:"source"`
	Title         string   `json:"title"`
	Artists       []Artist `json:"artists"`
	Album         *Album   `json:"album,omitempty"`
	Duration      int64    `json:"duration"`
	Cover         string   `json:"cover,omitempty"`
	CoverOriginal string   `json:"coverOriginal,omitempty"`
}

type LyricWord struct {
	Word      string `json:"word"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

type LyricLine struct {
	StartTime       int64       `json:"startTime"`
	EndTime         int64       `json:"endTime"`
	Words           []LyricWord `json:"words,omitempty"`
	Text            string      `json:"text,omitempty"`
	TranslatedLyric string      `json:"translatedLyric,omitempty"`
	RomanLyric      string      `json:"romanLyric,omitempty"`
}

type LyricSource struct {
	Source string `json:"source"`
	Format string `json:"format"`
}

type LyricSnapshot struct {
	TrackID       string       `json:"trackId"`
	Lyric         []LyricLine  `json:"lyric"`
	Source        *LyricSource `json:"source"`
	LyricOffsetMs int64        `json:"lyricOffsetMs"`
}

type Status struct {
	State      PlayerState `json:"state"`
	Position   int64       `json:"position"`
	Duration   int64       `json:"duration"`
	Volume     float64     `json:"volume"`
	Speed      float64     `json:"speed"`
	IsFinished bool        `json:"isFinished"`
}

type NowPlaying struct {
	Track          *Track      `json:"track"`
	Position       int64       `json:"position"`
	Playing        bool        `json:"playing"`
	State          PlayerState `json:"state"`
	Speed          float64     `json:"speed"`
	LyricOffsetMs  int64       `json:"lyricOffsetMs"`
	LyricAvailable bool        `json:"lyricAvailable"`
	LyricLineCount int         `json:"lyricLineCount"`
	SendTimestamp  int64       `json:"sendTimestamp"`
}

type Snapshot struct {
	NowPlaying NowPlaying
	Status     Status
	Lyrics     LyricSnapshot
}

// Provider is implemented by the UI player adapter. Method names are prefixed
// to avoid colliding with the embedded player.Player methods.
type Provider interface {
	ExternalAPISnapshot() Snapshot
	ExternalAPIPlay()
	ExternalAPIPause()
	ExternalAPIStop()
	ExternalAPINext()
	ExternalAPIPrevious()
	ExternalAPISeek(positionMs int64)
	ExternalAPISetVolume(volume float64)
}

type Options struct {
	Enable   bool
	EnableWS bool
	AllowLAN bool
	Port     int
}
