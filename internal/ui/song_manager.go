package ui

import (
	"math/rand"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type optionalSong []structs.Song

func some(song structs.Song) optionalSong {
	return optionalSong([]structs.Song{song})
}

func (o optionalSong) unwrap() structs.Song {
	if o != nil {
		return ([]structs.Song)(o)[0]
	}
	panic("unwrap() called on an empty value")
}

func (o optionalSong) ifSome(f func(structs.Song)) {
	if o != nil {
		f(o.unwrap())
	}
}

type songManager interface {
	getPlaylist() []structs.Song
	init(index int, playlist []structs.Song)
	getCurSongIndex() int
	nextSong(manual bool) optionalSong
	prevSong(manual bool) optionalSong
	delSong(index int) optionalSong

	ordered() orderedSongManager
	infRandom() infRandomSongManager
	listRandom() listRandomSongManager
	singleLoop() singleLoopSongManager
	listLoop() listLoopSongManager

	modeName() string
	mode() types.Mode
}

type orderedSongManager struct {
	index    int
	playlist []structs.Song
}

func (m *orderedSongManager) getPlaylist() []structs.Song {
	return m.playlist
}

func (m *orderedSongManager) getCurSongIndex() int {
	return m.index
}

func (m *orderedSongManager) init(index int, playlist []structs.Song) {
	m.index = index
	m.playlist = playlist
}

func (m *orderedSongManager) nextSong(_ bool) optionalSong {
	if len(m.playlist) == 0 || m.index == len(m.playlist)-1 {
		return nil
	}
	m.index++
	return some(m.playlist[m.index])
}

func (m *orderedSongManager) prevSong(_ bool) optionalSong {
	if len(m.playlist) == 0 || m.index == 0 {
		return nil
	}
	m.index--
	return some(m.playlist[m.index])
}

func (m *orderedSongManager) delSong(index int) (song optionalSong) {
	if index == m.index {
		if index == len(m.playlist)-1 {
			if len(m.playlist) > 1 {
				song = some(m.playlist[index-1])
				m.index--
			}
		} else {
			song = some(m.playlist[index+1])
			m.index++
		}
	} else if index < m.index {
		m.index--
	}
	m.playlist = append(m.playlist[:index], m.playlist[index+1:]...)
	return
}

func (m orderedSongManager) ordered() orderedSongManager {
	return m
}

func (m orderedSongManager) infRandom() infRandomSongManager {
	return infRandomSongManager{
		playlist: m.playlist,
		curSong: &randomSong{
			index: m.index,
			next:  nil,
			prev:  nil,
		},
	}
}

func (m orderedSongManager) listRandom() listRandomSongManager {
	manager := listRandomSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
	manager.shuffle()
	return manager
}

func (m orderedSongManager) singleLoop() singleLoopSongManager {
	return singleLoopSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
}

func (m orderedSongManager) listLoop() listLoopSongManager {
	return listLoopSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
}

func (m orderedSongManager) mode() types.Mode {
	return types.PmOrdered
}

func (m orderedSongManager) modeName() string {
	return "顺序"
}

type randomSong struct {
	index int
	next  *randomSong
	prev  *randomSong
}

type infRandomSongManager struct {
	playlist []structs.Song
	curSong  *randomSong
}

func (m *infRandomSongManager) getPlaylist() []structs.Song {
	return m.playlist
}

func (m *infRandomSongManager) getCurSongIndex() int {
	return m.curSong.index
}

func (m *infRandomSongManager) init(index int, playlist []structs.Song) {
	m.playlist = playlist
	m.curSong = &randomSong{
		index: index,
		next:  nil,
		prev:  nil,
	}
}

func (m *infRandomSongManager) nextSong(_ bool) optionalSong {
	if len(m.playlist) == 0 {
		return nil
	} else if len(m.playlist) == 1 {
		return some(m.playlist[0])
	}
	var index int
	if m.curSong != nil && m.curSong.next != nil {
		index = m.curSong.next.index
		m.curSong = m.curSong.next
		return some(m.playlist[index])
	}
	index = rand.Intn(len(m.playlist) - 1)
	for index == m.curSong.index {
		index = rand.Intn(len(m.playlist) - 1)
	}
	randSong := randomSong{
		index: index,
		next:  nil,
		prev:  m.curSong,
	}
	if m.curSong != nil {
		m.curSong.next = &randSong
	}
	m.curSong = &randSong
	return some(m.playlist[index])
}

func (m *infRandomSongManager) prevSong(_ bool) optionalSong {
	if len(m.playlist) == 0 {
		return nil
	} else if len(m.playlist) == 1 {
		return some(m.playlist[0])
	}
	var index int
	if m.curSong != nil && m.curSong.prev != nil {
		index = m.curSong.prev.index
		m.curSong = m.curSong.prev
		return some(m.playlist[index])
	}
	index = rand.Intn(len(m.playlist) - 1)
	for index == m.curSong.index {
		index = rand.Intn(len(m.playlist) - 1)
	}
	randSong := randomSong{
		index: index,
		next:  m.curSong,
		prev:  nil,
	}
	if m.curSong != nil {
		m.curSong.prev = &randSong
	}
	m.curSong = &randSong
	return some(m.playlist[index])
}

func (m *infRandomSongManager) delSong(index int) (song optionalSong) {
	if index == m.curSong.index {
		if index == len(m.playlist)-1 {
			if len(m.playlist) > 1 {
				song = some(m.playlist[index-1])
				m.curSong.index--
			}
		} else {
			song = some(m.playlist[index+1])
			m.curSong.index++
		}
	} else {
		if m.curSong.prev != nil && m.curSong.prev.index == index {
			m.curSong.prev = nil
		} else if m.curSong.next != nil && m.curSong.next.index == index {
			m.curSong.next = nil
		}
		if index < m.curSong.index {
			m.curSong.index--
		}
	}
	m.playlist = append(m.playlist[:index], m.playlist[index+1:]...)
	cur := m.curSong.prev
	for cur != nil {
		if index < cur.index {
			cur.index--
		}
		cur = cur.prev
	}
	cur = m.curSong.next
	for cur != nil {
		if index < cur.index {
			cur.index--
		}
	}
	return
}

func (m infRandomSongManager) ordered() orderedSongManager {
	return orderedSongManager{
		index:    m.curSong.index,
		playlist: m.playlist,
	}
}

func (m infRandomSongManager) infRandom() infRandomSongManager {
	return m
}

func (m infRandomSongManager) listRandom() listRandomSongManager {
	manager := listRandomSongManager{
		index:    m.curSong.index,
		playlist: m.playlist,
	}
	manager.shuffle()
	return manager
}

func (m infRandomSongManager) singleLoop() singleLoopSongManager {
	return singleLoopSongManager{
		index:    m.curSong.index,
		playlist: m.playlist,
	}
}

func (m infRandomSongManager) listLoop() listLoopSongManager {
	return listLoopSongManager{
		index:    m.curSong.index,
		playlist: m.playlist,
	}
}

func (m infRandomSongManager) mode() types.Mode {
	return types.PmInfRandom
}

func (m infRandomSongManager) modeName() string {
	return "无限随机"
}

type listRandomSongManager struct {
	index    int
	playlist []structs.Song
	_index   int
	order    []int
}

func (m *listRandomSongManager) shuffle() {
	if len(m.playlist) == 0 {
		m.order = make([]int, 0)
		m._index = 0
		return
	}
	m.order = make([]int, len(m.playlist))
	for idx := range m.playlist {
		m.order[idx] = idx
	}
	m._index = 0
	m.order[0], m.order[m.index] = m.order[m.index], m.order[0]
	rand.Shuffle(len(m.order)-1, func(i, j int) {
		m.order[j+1], m.order[i+1] = m.order[i+1], m.order[j+1]
	})
}

func (m *listRandomSongManager) getPlaylist() []structs.Song {
	return m.playlist
}

func (m *listRandomSongManager) getCurSongIndex() int {
	return m.index
}

func (m *listRandomSongManager) init(index int, playlist []structs.Song) {
	m.index = index
	m.playlist = playlist
	m.shuffle()
}

func (m *listRandomSongManager) nextSong(_ bool) optionalSong {
	if len(m.playlist) == 0 || m._index == len(m.playlist)-1 {
		return nil
	}
	m._index++
	m.index = m.order[m._index]
	return some(m.playlist[m.index])
}

func (m *listRandomSongManager) prevSong(_ bool) optionalSong {
	if len(m.playlist) == 0 || m._index == 0 {
		return nil
	}
	m._index--
	m.index = m.order[m._index]
	return some(m.playlist[m.index])
}

func (m *listRandomSongManager) delSong(index int) (song optionalSong) {
	if index == m.index {
		if index == len(m.playlist)-1 {
			if len(m.playlist) > 1 {
				song = some(m.playlist[index-1])
				m.index--
			}
		} else {
			song = some(m.playlist[index+1])
			m.index++
		}
		m.shuffle()
	} else if index < m.index {
		m.index--
	}
	m.playlist = append(m.playlist[:index], m.playlist[index+1:]...)
	return
}

func (m listRandomSongManager) ordered() orderedSongManager {
	return orderedSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
}

func (m listRandomSongManager) infRandom() infRandomSongManager {
	return infRandomSongManager{
		playlist: m.playlist,
		curSong: &randomSong{
			index: m.index,
			next:  nil,
			prev:  nil,
		},
	}
}

func (m listRandomSongManager) listRandom() listRandomSongManager {
	return m
}

func (m listRandomSongManager) singleLoop() singleLoopSongManager {
	return singleLoopSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
}

func (m listRandomSongManager) listLoop() listLoopSongManager {
	return listLoopSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
}

func (m listRandomSongManager) mode() types.Mode {
	return types.PmListRandom
}

func (m listRandomSongManager) modeName() string {
	return "列表随机"
}

type singleLoopSongManager struct {
	index    int
	playlist []structs.Song
}

func (m *singleLoopSongManager) getPlaylist() []structs.Song {
	return m.playlist
}

func (m *singleLoopSongManager) getCurSongIndex() int {
	return m.index
}

func (m *singleLoopSongManager) init(index int, playlist []structs.Song) {
	m.index = index
	m.playlist = playlist
}

func (m *singleLoopSongManager) nextSong(manual bool) optionalSong {
	if len(m.playlist) == 0 {
		return nil
	}
	if !manual {
		return some(m.playlist[m.index])
	}
	if m.index == len(m.playlist)-1 {
		return nil
	}
	m.index++
	return some(m.playlist[m.index])
}

func (m *singleLoopSongManager) prevSong(manual bool) optionalSong {
	if len(m.playlist) == 0 {
		return nil
	}
	if !manual {
		return some(m.playlist[m.index])
	}
	if m.index == 0 {
		return nil
	}
	m.index--
	return some(m.playlist[m.index])
}

func (m *singleLoopSongManager) delSong(index int) (song optionalSong) {
	if index == m.index {
		if index == len(m.playlist)-1 {
			if len(m.playlist) > 1 {
				song = some(m.playlist[index-1])
				m.index--
			}
		} else {
			song = some(m.playlist[index+1])
			m.index++
		}
	} else if index < m.index {
		m.index--
	}
	m.playlist = append(m.playlist[:index], m.playlist[index+1:]...)
	return
}

func (m singleLoopSongManager) ordered() orderedSongManager {
	return orderedSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
}

func (m singleLoopSongManager) infRandom() infRandomSongManager {
	return infRandomSongManager{
		playlist: m.playlist,
		curSong: &randomSong{
			index: m.index,
			next:  nil,
			prev:  nil,
		},
	}
}

func (m singleLoopSongManager) listRandom() listRandomSongManager {
	manager := listRandomSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
	manager.shuffle()
	return manager
}

func (m singleLoopSongManager) singleLoop() singleLoopSongManager {
	return m
}

func (m singleLoopSongManager) listLoop() listLoopSongManager {
	return listLoopSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
}

func (m singleLoopSongManager) mode() types.Mode {
	return types.PmSingleLoop
}

func (m singleLoopSongManager) modeName() string {
	return "单曲"
}

type listLoopSongManager struct {
	index    int
	playlist []structs.Song
}

func (m *listLoopSongManager) getPlaylist() []structs.Song {
	return m.playlist
}

func (m *listLoopSongManager) getCurSongIndex() int {
	return m.index
}

func (m *listLoopSongManager) init(index int, playlist []structs.Song) {
	m.index = index
	m.playlist = playlist
}

func (m *listLoopSongManager) nextSong(_ bool) optionalSong {
	if len(m.playlist) == 0 {
		return nil
	}
	m.index = (m.index + 1) % len(m.playlist)
	return some(m.playlist[m.index])
}

func (m *listLoopSongManager) prevSong(_ bool) optionalSong {
	if len(m.playlist) == 0 {
		return nil
	}
	m.index = (m.index - 1 + len(m.playlist)) % len(m.playlist)
	return some(m.playlist[m.index])
}

func (m *listLoopSongManager) delSong(index int) (song optionalSong) {
	if index == m.index {
		if index == len(m.playlist)-1 {
			if len(m.playlist) > 1 {
				song = some(m.playlist[index-1])
				m.index--
			}
		} else {
			song = some(m.playlist[index+1])
			m.index++
		}
	} else if index < m.index {
		m.index--
	}
	m.playlist = append(m.playlist[:index], m.playlist[index+1:]...)
	return
}

func (m listLoopSongManager) ordered() orderedSongManager {
	return orderedSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
}

func (m listLoopSongManager) infRandom() infRandomSongManager {
	return infRandomSongManager{
		playlist: m.playlist,
		curSong: &randomSong{
			index: m.index,
			next:  nil,
			prev:  nil,
		},
	}
}

func (m listLoopSongManager) listRandom() listRandomSongManager {
	manager := listRandomSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
	manager.shuffle()
	return manager
}

func (m listLoopSongManager) singleLoop() singleLoopSongManager {
	return singleLoopSongManager{
		index:    m.index,
		playlist: m.playlist,
	}
}

func (m listLoopSongManager) listLoop() listLoopSongManager {
	return m
}

func (m listLoopSongManager) mode() types.Mode {
	return types.PmListLoop
}

func (m listLoopSongManager) modeName() string {
	return "列表"
}
