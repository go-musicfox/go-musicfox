package player

type Player struct {
    CurMusic Music
    Playlist []Music
    Lyric    map[float64]string
}