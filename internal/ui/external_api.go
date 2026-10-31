package ui

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/externalapi"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

var _ externalapi.Provider = (*Player)(nil)

func (p *Player) ExternalAPISnapshot() externalapi.Snapshot {
	now := time.Now().UnixMilli()
	track := p.externalAPITrack(p.CurSong())
	state := externalAPIState(p.State())
	duration := int64(0)
	if track != nil {
		duration = track.Duration
	}
	lyrics := p.externalAPILyrics(track)
	position := p.PassedTime().Milliseconds()
	volume := float64(p.Volume()) / 100
	status := externalapi.Status{
		State:      state,
		Position:   position,
		Duration:   duration,
		Volume:     volume,
		Speed:      1,
		IsFinished: false,
	}
	nowPlaying := externalapi.NowPlaying{
		Track:          track,
		Position:       position,
		Playing:        state == externalapi.StatePlaying,
		State:          state,
		Speed:          1,
		LyricOffsetMs:  lyrics.LyricOffsetMs,
		LyricAvailable: len(lyrics.Lyric) > 0,
		LyricLineCount: len(lyrics.Lyric),
		SendTimestamp:  now,
	}
	return externalapi.Snapshot{NowPlaying: nowPlaying, Status: status, Lyrics: lyrics}
}

func (p *Player) externalAPITrack(song structs.Song) *externalapi.Track {
	if song.Id == 0 {
		return nil
	}
	artists := make([]externalapi.Artist, 0, len(song.Artists))
	for _, artist := range song.Artists {
		artists = append(artists, externalapi.Artist{ID: strconv.FormatInt(artist.Id, 10), Name: artist.Name})
	}
	cover := song.PicUrl
	if cover == "" {
		cover = song.Album.PicUrl
	}
	album := &externalapi.Album{ID: strconv.FormatInt(song.Album.Id, 10), Name: song.Album.Name, Cover: cover}
	return &externalapi.Track{
		ID:            strconv.FormatInt(song.Id, 10),
		Source:        "netease",
		Title:         song.Name,
		Artists:       artists,
		Album:         album,
		Duration:      song.Duration.Milliseconds(),
		Cover:         cover,
		CoverOriginal: cover,
	}
}

func (p *Player) externalAPILyrics(track *externalapi.Track) externalapi.LyricSnapshot {
	state := p.lyricService.State()
	snapshot := externalapi.LyricSnapshot{LyricOffsetMs: state.OffsetMs}
	if track != nil {
		snapshot.TrackID = track.ID
	}
	if !state.IsRunning {
		return snapshot
	}
	if len(state.YRCLines) > 0 {
		snapshot.Source = &externalapi.LyricSource{Source: "online", Format: "yrc"}
		for index, line := range state.YRCLines {
			endTime := line.EndTime
			if endTime <= line.StartTime {
				endTime = nextYRCEnd(state.YRCLines, index)
			}
			words := make([]externalapi.LyricWord, 0, len(line.Words))
			var text strings.Builder
			for _, word := range line.Words {
				words = append(words, externalapi.LyricWord{Word: word.Word, StartTime: word.StartTime, EndTime: word.EndTime})
				text.WriteString(word.Word)
			}
			snapshot.Lyric = append(snapshot.Lyric, externalapi.LyricLine{
				StartTime:       line.StartTime,
				EndTime:         endTime,
				Words:           words,
				Text:            text.String(),
				TranslatedLyric: line.TranslatedLyric,
				RomanLyric:      line.RomanLyric,
			})
		}
		return snapshot
	}
	snapshot.Source = &externalapi.LyricSource{Source: "online", Format: "lrc"}
	for index, fragment := range state.Fragments {
		endTime := nextFragmentEnd(state.Fragments, index)
		snapshot.Lyric = append(snapshot.Lyric, externalapi.LyricLine{
			StartTime:       fragment.StartTimeMs,
			EndTime:         endTime,
			Text:            fragment.Content,
			TranslatedLyric: state.TranslatedFragments[fragment.StartTimeMs],
		})
	}
	return snapshot
}

func nextYRCEnd(lines []lyric.YRCLine, index int) int64 {
	if index+1 < len(lines) && lines[index+1].StartTime > lines[index].StartTime {
		return lines[index+1].StartTime
	}
	return lines[index].StartTime + 4000
}

func nextFragmentEnd(fragments []lyric.LRCFragment, index int) int64 {
	if index+1 < len(fragments) && fragments[index+1].StartTimeMs > fragments[index].StartTimeMs {
		return fragments[index+1].StartTimeMs
	}
	return fragments[index].StartTimeMs + 4000
}

func externalAPIState(state types.State) externalapi.PlayerState {
	switch state {
	case types.Playing:
		return externalapi.StatePlaying
	case types.Paused:
		return externalapi.StatePaused
	case types.Stopped, types.Interrupted:
		return externalapi.StateStopped
	case types.Unknown:
		return externalapi.StateIdle
	default:
		return externalapi.StateIdle
	}
}

func (p *Player) ExternalAPIPlay()     { p.CtrlResume() }
func (p *Player) ExternalAPIPause()    { p.CtrlPause() }
func (p *Player) ExternalAPIStop()     { p.CtrlStop() }
func (p *Player) ExternalAPINext()     { p.CtrlNext() }
func (p *Player) ExternalAPIPrevious() { p.CtrlPrevious() }

func (p *Player) ExternalAPISeek(positionMs int64) {
	if positionMs < 0 {
		return
	}
	p.CtrlSeek(time.Duration(positionMs) * time.Millisecond)
}

func (p *Player) ExternalAPISetVolume(volume float64) {
	if volume < 0 || volume > 1 {
		return
	}
	p.CtrlSetVolume(int(volume*100 + 0.5))
}
