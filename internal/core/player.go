package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/playermiddleware"
	"github.com/go-musicfox/go-musicfox/internal/playlist"
	control "github.com/go-musicfox/go-musicfox/internal/remote_control"
	"github.com/go-musicfox/go-musicfox/internal/reporter"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

// PlayerOptions carries the playback dependencies injected by the frontend
// shell. Every field is optional; a nil dependency degrades gracefully.
type PlayerOptions struct {
	LyricService  *lyric.Service
	TrackManager  *track.Manager
	DesktopLyrics desktop_lyrics.Controller
	User          **structs.User   // shared user slot owned by the shell (like n.user)
	LastfmTracker *lastfm.Tracker  // used to build the Last.fm reporter, if enabled
}

// Player 网易云音乐播放器
type Player struct {
	player.Player             // embedded engine (verbatim)
	playlistManager  playlist.PlaylistManager
	lastMode         types.Mode
	playingMenuKey   string
	lyricService     *lyric.Service
	trackManager     *track.Manager
	desktopLyrics    desktop_lyrics.Controller
	user             **structs.User
	observer         Observer          // set later via SetObserver
	loading          LoadingIndicator  // set later via SetLoading
	locator          SongLocator       // set later via SetLocator
	stateHandler     *control.RemoteControl
	ctrl             chan CtrlSignal
	reporter         reporter.Service
	middlewareChain  *playermiddleware.Chain
	playErrCount     int
	cancel           context.CancelFunc
	playlistUpdateAt time.Time
	gaplessMu        sync.Mutex
	gaplessPending   int64
	gaplessLoading   bool
	gaplessTriedFor  int64
	mprisPosThrottle mprisPositionThrottle
}

// NewPlayer builds the playback coordinator: engine, playlist manager,
// reporter, middleware chain, remote control and the three background loops
// (ctrl signals, state changes, time ticks). Observer/loading/locator are
// attached afterwards by the frontend via SetObserver/SetLoading/SetLocator.
func NewPlayer(opts PlayerOptions) *Player {
	reporterOptions := []reporter.Option{}
	if configs.AppConfig.Reporter.Lastfm.Enable {
		skipDjRadio := configs.AppConfig.Reporter.Lastfm.SkipDjRadio
		reporterOptions = append(reporterOptions, reporter.WithLastFM(opts.LastfmTracker, skipDjRadio))
	}
	if configs.AppConfig.Reporter.Netease.Enable {
		reporterOptions = append(reporterOptions, reporter.WithNetease())
	}

	p := &Player{
		lyricService:    opts.LyricService,
		trackManager:    opts.TrackManager,
		desktopLyrics:   opts.DesktopLyrics,
		user:            opts.User,
		playlistManager: playlist.NewPlaylistManager(),
		ctrl:            make(chan CtrlSignal, 10),
		reporter:        reporter.NewService(reporterOptions...),
		middlewareChain: playermiddleware.NewChain().Use(playermiddleware.NewUNMMiddleware()),
	}
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())

	p.Player = player.NewPlayerFromConfig()
	var gaplessTransitions <-chan player.GaplessTransition
	if gapless, ok := p.Player.(player.GaplessPlayer); ok {
		gaplessTransitions = gapless.GaplessTransitionChan()
	}
	p.stateHandler = control.NewRemoteControl(p, p.PlayingInfo())

	// remote control
	errorx.WaitGoStart(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case signal := <-p.ctrl:
				p.handleControlSignal(signal)
			}
		}
	})

	// 状态监听
	errorx.WaitGoStart(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-p.StateChan():
				p.stateHandler.SetPlayingInfo(p.PlayingInfo())
				p.updateDesktopLyrics()
				if s != types.Stopped {
					if p.observer != nil {
						p.observer.OnStateChanged(s)
					}
					break
				}
				p.NextSong(false)
			}
		}
	})

	// 时间监听
	errorx.WaitGoStart(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case transition := <-gaplessTransitions:
				p.commitGaplessTransition(transition)
			case duration := <-p.TimeChan():
				if pos := p.PassedTime(); p.mprisPosThrottle.shouldEmit(pos, time.Now()) {
					p.stateHandler.SetPosition(pos)
				}
				if duration.Seconds()-p.CurMusic().Duration.Seconds() > 10 {
					p.NextSong(false)
				}
				p.maybePreloadGapless(duration)

				p.lyricService.UpdatePosition(duration)

				// Update desktop lyrics
				p.updateDesktopLyrics()

				if p.observer != nil {
					p.observer.OnPosition(duration)
				}
			}
		}
	})

	return p
}

// NewEmptyPlayer builds a bare Player with a fresh playlist manager but no
// engine, remote control, reporter or goroutines. It exists so frontends (and
// their unit tests) can exercise the playlist/state API without starting
// playback.
func NewEmptyPlayer() *Player {
	return &Player{playlistManager: playlist.NewPlaylistManager()}
}

// SetObserver attaches the frontend observer (nil allowed).
func (p *Player) SetObserver(o Observer) {
	p.observer = o
}

// SetLoading attaches the frontend loading indicator (nil allowed).
func (p *Player) SetLoading(l LoadingIndicator) {
	p.loading = l
}

// SetLocator attaches the frontend playing-song locator (nil allowed).
func (p *Player) SetLocator(l SongLocator) {
	p.locator = l
}

// Engine returns the underlying playback engine. Frontends use it to reach
// engine-only capabilities (SpectrumProvider, VolumeStorable, ...).
func (p *Player) Engine() player.Player {
	return p.Player
}

// SetPlayingMenu records the key of the menu that owns the currently playing
// playlist. The TUI wrapper pairs this with its own menu reference.
func (p *Player) SetPlayingMenu(key string) {
	p.playingMenuKey = key
}

// PlayingMenuKey returns the menu key of the currently playing menu.
func (p *Player) PlayingMenuKey() string {
	return p.playingMenuKey
}

// MarkPlaylistModified invalidates the playing-menu association after the
// playlist is mutated externally, avoiding stale menu data. The TUI wrapper
// additionally clears its menu reference.
func (p *Player) MarkPlaylistModified() {
	p.playingMenuKey += "modified"
}

// MarkPlaylistUpdated stamps the playlist update time.
func (p *Player) MarkPlaylistUpdated() {
	p.playlistUpdateAt = time.Now()
}

// PlaylistUpdateAt returns the playlist update time.
func (p *Player) PlaylistUpdateAt() time.Time {
	return p.playlistUpdateAt
}

// CompareWithCurPlaylist 与当前播放列表对比，是否一致
func (p *Player) CompareWithCurPlaylist(playlist []structs.Song) bool {
	if len(playlist) != len(p.Playlist()) {
		return false
	}

	// 如果前20个一致，则认为相同
	for i := 0; i < 20 && i < len(playlist); i++ {
		if playlist[i].Id != p.Playlist()[i].Id {
			return false
		}
	}

	return true
}

// PlaySong 播放歌曲
func (p *Player) PlaySong(song structs.Song, direction PlayDirection) {
	p.cancelGaplessPreload()
	p.reporter.ReportEnd(p.PlayedTime())

	if p.loading != nil {
		p.loading.Start()
		defer p.loading.Complete()
	}

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayerSnapshot{}, storage.PlayerSnapshot{
		CurSongIndex:     p.CurSongIndex(),
		Playlist:         p.Playlist(),
		PlaylistUpdateAt: p.playlistUpdateAt,
	})

	if p.locator != nil {
		p.locator.LocatePlayingSong()
	}
	p.Pause()
	url, musicType, err := p.getPlayInfo(song)

	// 构造 URLMusic 并经过中间件链（如 UNM banned-path 拦截），
	// 中间件可改写 URL 或以 ErrBlockedTrack 拦截播放。
	urlMusic := &player.URLMusic{URL: url, Song: song, Type: player.SongTypeMapping[musicType]}
	var skip bool
	if err == nil {
		if mwErr := p.middlewareChain.Execute(context.Background(), urlMusic); mwErr != nil {
			if errors.Is(mwErr, playermiddleware.ErrBlockedTrack) {
				skip = true
			} else {
				err = mwErr
			}
		}
	}
	url = urlMusic.URL
	// logger 在中间件链执行之后构造，记录的是实际用于播放的 URL
	// （可能已被改写型中间件重写）。
	logger := slog.With(slog.String("url", url), slog.String("type", musicType), slog.Any("song", song))

	if url == "" || err != nil || skip {
		p.playErrCount++
		if skip {
			logger.Info("已拦截无效播放")
		} else {
			logger.Error("Play song error", slog.Any("err", err))
		}
		if p.playErrCount >= configs.AppConfig.Player.MaxPlayErrCount {
			return
		}
		switch direction {
		case DurationPrev:
			p.PreviousSong(false)
		case DurationNext:
			p.NextSong(false)
		}
		return
	}

	errorx.Go(func() {
		_ = p.lyricService.SetSong(context.Background(), song)
	}, true)

	p.Play(*urlMusic)
	slog.Info("Start play song", slog.String("url", url), slog.String("type", musicType), slog.Any("song", song))

	// 上报开始播放
	p.reporter.ReportStart(song)

	if p.observer != nil {
		p.observer.OnSongChanged(song)
	}

	go notify.Notify(notify.NotifyContent{
		Title:   "正在播放: " + song.Name,
		Text:    fmt.Sprintf("%s - %s", song.ArtistName(), song.Album.Name),
		Icon:    app.AddResizeParamForPicUrl(song.PicUrl, 60),
		Url:     netease.WebUrlOfSong(song.Id),
		GroupId: types.GroupID,
	})

	p.playErrCount = 0
}

func (p *Player) StartPlay() {
	if len(p.Playlist()) <= p.CurSongIndex() {
		return
	}
	p.PlaySong(p.CurSong(), DurationNext)
}

func (p *Player) Mode() types.Mode {
	return p.playlistManager.GetPlayMode()
}

func (p *Player) Playlist() []structs.Song {
	return p.playlistManager.GetPlaylist()
}

func (p *Player) InitSongManager(index int, playlist []structs.Song) {
	p.cancelGaplessPreload()
	p.ReinitializePlaylist(index, playlist)
}

// ReinitializePlaylist re-initializes the playlist with the given index and songs.
func (p *Player) ReinitializePlaylist(index int, playlist []structs.Song) {
	_ = p.playlistManager.Initialize(index, playlist)
}

// NextPlaylistSong advances the playlist to the next song without starting
// playback and returns it (mirrors playlistManager.NextSong). It is used by
// the TUI wrapper's Intelligence append flow to pick the song to play next.
func (p *Player) NextPlaylistSong(manual bool) (structs.Song, error) {
	return p.playlistManager.NextSong(manual)
}

// RemoveSong removes the song at the given index from the playlist.
func (p *Player) RemoveSong(index int) (structs.Song, error) {
	return p.playlistManager.RemoveSong(index)
}

// LoadPlaylistState restores the persisted playlist state from storage.
func (p *Player) LoadPlaylistState() error {
	return p.playlistManager.LoadState()
}

func (p *Player) CurSongIndex() int {
	return p.playlistManager.GetCurrentIndex()
}

func (p *Player) CurSong() structs.Song {
	index := p.CurSongIndex()
	if index < 0 || len(p.Playlist()) <= index {
		return structs.Song{}
	}
	return p.Playlist()[index]
}

// NextSong 下一曲
func (p *Player) NextSong(manual bool) {
	index := p.CurSongIndex()
	playlistLen := len(p.Playlist())

	// 到达底部，则触发翻页或加载更多
	if playlistLen == 0 || index >= playlistLen-1 {
		if o, ok := p.observer.(PlaylistExhaustedObserver); ok {
			o.OnPlaylistExhausted(DurationNext)
		}
	}

	// 尝试获取下一首歌曲
	song, err := p.playlistManager.NextSong(manual)
	if err != nil {
		slog.Error("Get next song error", slog.Any("err", err), slog.String("play_mode", p.playlistManager.GetPlayMode().Name()))
		return
	}

	p.PlaySong(song, DurationNext)
}

// PreviousSong 上一曲
func (p *Player) PreviousSong(manual bool) {
	index := p.CurSongIndex()
	playlistLen := len(p.Playlist())
	if playlistLen == 0 || index >= playlistLen-1 {
		if o, ok := p.observer.(PlaylistExhaustedObserver); ok {
			o.OnPlaylistExhausted(DurationPrev)
		}
	}

	if song, err := p.playlistManager.PreviousSong(manual); err == nil {
		p.PlaySong(song, DurationNext)
	}
}

func (p *Player) Seek(duration time.Duration) {
	p.Player.Seek(duration)
	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
	p.stateHandler.EmitSeeked(duration)
}

// SetMode 设置播放模式
func (p *Player) SetMode(playMode types.Mode) {
	p.cancelGaplessPreload()
	if p.lastMode != p.Mode() {
		p.lastMode = p.Mode()
	}

	// 直接使用PlaylistManager设置播放模式
	_ = p.playlistManager.SetPlayMode(playMode)

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayMode{}, playMode)
}

// SwitchMode 顺序切换播放模式
func (p *Player) SwitchMode() {
	mode := p.Mode()
	supportedModes := p.playlistManager.SupportedPlayModes()
	index := 0
	for i, m := range supportedModes {
		if mode != m {
			continue
		}
		index = i + 1
		break
	}

	for {
		if index >= len(supportedModes) {
			index = 0
		}
		if supportedModes[index] == types.PmIntelligent {
			index++
			continue
		}
		break
	}

	p.SetMode(supportedModes[index])
}

// toggleShuffle toggles between shuffle on (PmListRandom) and shuffle off (PmListLoop)
func (p *Player) toggleShuffle() {
	mode := p.Mode()
	if mode == types.PmListRandom {
		p.SetMode(types.PmListLoop)
	} else {
		p.SetMode(types.PmListRandom)
	}
}

// cycleRepeat cycles through repeat modes: PmOrdered -> PmListLoop -> PmSingleLoop -> PmOrdered
func (p *Player) cycleRepeat() {
	mode := p.Mode()
	switch mode {
	case types.PmOrdered:
		p.SetMode(types.PmListLoop)
	case types.PmListLoop:
		p.SetMode(types.PmSingleLoop)
	case types.PmSingleLoop:
		p.SetMode(types.PmOrdered)
	default:
		// For other modes (like PmListRandom), set to PmListLoop
		p.SetMode(types.PmListLoop)
	}
}

// setRepeat sets the repeat mode directly based on MPRepeatType
// repeatType should be int (0=off, 1=one, 2=all)
func (p *Player) setRepeat(repeatType any) {
	if repeatType == nil {
		return
	}
	mode, ok := repeatType.(int)
	if !ok {
		return
	}
	switch mode {
	case 0: // MPRepeatTypeOff
		p.SetMode(types.PmOrdered)
	case 1: // MPRepeatTypeOne
		p.SetMode(types.PmSingleLoop)
	case 2: // MPRepeatTypeAll
		p.SetMode(types.PmListLoop)
	}
}

// setShuffle sets the shuffle mode directly based on MPShuffleType
// shuffleType should be int (0=off, 1=items, 2=collections)
func (p *Player) setShuffle(shuffleType any) {
	if shuffleType == nil {
		return
	}
	mode, ok := shuffleType.(int)
	if !ok {
		return
	}
	switch mode {
	case 0: // MPShuffleTypeOff
		// Keep current repeat mode but disable shuffle
		currentMode := p.Mode()
		if currentMode == types.PmListRandom {
			p.SetMode(types.PmListLoop)
		}
	case 1, 2: // MPShuffleTypeItems, MPShuffleTypeCollections
		p.SetMode(types.PmListRandom)
	}
}

// Close 关闭
func (p *Player) Close() error {
	// 退出前上报
	p.reporter.ReportEnd(p.PlayedTime())

	p.cancel()
	if p.stateHandler != nil {
		p.stateHandler.Release()
	}
	p.Player.Close()
	return nil
}

func (p *Player) getPlayInfo(song structs.Song) (string, string, error) {
	source, err := p.trackManager.ResolvePlayableSource(context.Background(), song)
	if err != nil || source.Info == nil {
		return "", "", err
	}
	url := source.Info.URL
	musicType := source.Info.MusicType
	return url, musicType, err
}

func (p *Player) UpVolume() {
	p.Player.UpVolume()

	if v, ok := p.Player.(storage.VolumeStorable); ok {
		table := storage.NewTable()
		_ = table.SetByKVModel(storage.Volume{}, v.Volume())
	}

	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

func (p *Player) DownVolume() {
	p.Player.DownVolume()

	if v, ok := p.Player.(storage.VolumeStorable); ok {
		table := storage.NewTable()
		_ = table.SetByKVModel(storage.Volume{}, v.Volume())
	}

	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

func (p *Player) SetVolume(volume int) {
	p.Player.SetVolume(volume)

	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

func (p *Player) handleControlSignal(signal CtrlSignal) {
	switch signal.Type {
	case CtrlPaused:
		p.Pause()
	case CtrlResume:
		p.Resume()
	case CtrlStop:
		p.Stop()
	case CtrlToggle:
		p.Toggle()
	case CtrlPrevious:
		p.PreviousSong(true)
	case CtrlNext:
		p.NextSong(true)
	case CtrlSeek:
		p.Seek(signal.Duration)
	case CtrlRerender:
		if o, ok := p.observer.(RerenderObserver); ok {
			o.OnRerender()
		}
	case CtrlShuffle:
		if signal.ShuffleType != 0 {
			p.setShuffle(signal.ShuffleType)
		} else {
			p.toggleShuffle()
		}
	case CtrlRepeat:
		if signal.RepeatType != 0 {
			p.setRepeat(signal.RepeatType)
		} else {
			p.cycleRepeat()
		}
	}
}

func (p *Player) PlayingInfo() control.PlayingInfo {
	song := p.CurSong()
	loopStatus, shuffle := modeToLoopStatusAndShuffle(p.Mode())
	return control.PlayingInfo{
		TotalDuration:  song.Duration,
		PassedDuration: p.PassedTime(),
		State:          p.State(),
		Volume:         p.Volume(),
		TrackID:        song.Id,
		PicUrl:         song.PicUrl,
		Name:           song.Name,
		Album:          song.Album.Name,
		Artist:         song.ArtistName(),
		AlbumArtist:    song.Album.ArtistName(),
		LRCText:        p.lyricService.State().FormatAsLRC(),
		LoopStatus:     loopStatus,
		Shuffle:        shuffle,
	}
}

// modeToLoopStatusAndShuffle converts types.Mode to MPRIS LoopStatus and Shuffle values.
func modeToLoopStatusAndShuffle(mode types.Mode) (loopStatus string, shuffle bool) {
	switch mode {
	case types.PmOrdered:
		return "None", false
	case types.PmListLoop, types.PmInfRandom, types.PmIntelligent:
		return "Playlist", false
	case types.PmSingleLoop:
		return "Track", false
	case types.PmListRandom:
		return "Playlist", true
	default:
		return "None", false
	}
}

func (p *Player) updateDesktopLyrics() {
	dl := p.desktopLyrics
	if dl == nil {
		return
	}

	state := p.State()
	curLine, nextLine, currentIndex := p.getDesktopLyricsLines()
	hasContent := curLine.Text != "" || len(curLine.Words) > 0
	currentTimeMs := p.PassedTime().Milliseconds()
	if configs.AppConfig.Main.Lyric.DesktopLyrics.HideOnPause && state == types.Paused {
		dl.Update(curLine, nextLine, currentIndex, currentTimeMs, false)
		dl.Hide()
		return
	}

	// Show/hide based on playback state
	switch state {
	case types.Playing:
		if hasContent {
			if !dl.IsVisible() {
				dl.Show()
			}
			dl.Update(curLine, nextLine, currentIndex, currentTimeMs, true)
		}
	case types.Paused:
		if hasContent {
			dl.Update(curLine, nextLine, currentIndex, currentTimeMs, false)
		}
	case types.Stopped:
		dl.Hide()
	}

	// Notify desktop lyrics whether the current player supports spectrum.
	// When false, the desktop lyrics window does not reserve background space
	// for spectrum bars even if SpectrumEnabled is true in config.
	_, hasSpectrum := p.Player.(player.SpectrumProvider)
	dl.SetSpectrumAvailable(hasSpectrum)

	// Pass spectrum data to desktop lyrics for GPU visualization
	if state == types.Playing || state == types.Paused {
		if provider, ok := p.Player.(player.SpectrumProvider); ok {
			dl.UpdateSpectrum(provider.Spectrum())
			dl.UpdateRawSamples(provider.RawSamples())
		}
	}
}

// getDesktopLyricsLines returns the current lyrics lines for desktop display
// (current line, next line, and current index). The logic mirrors the former
// shell method but depends only on the lyric service.
func (p *Player) getDesktopLyricsLines() (curLine, nextLine desktop_lyrics.LyricLine, currentIndex int) {
	if p.desktopLyrics == nil || p.lyricService == nil {
		return desktop_lyrics.LyricLine{}, desktop_lyrics.LyricLine{}, -1
	}

	state := p.lyricService.State()
	if !state.IsRunning {
		return desktop_lyrics.LyricLine{}, desktop_lyrics.LyricLine{}, -1
	}

	// Helper to build display text including translation
	buildText := func(content, translation string) string {
		if translation != "" {
			return content + "\n" + translation
		}
		return content
	}

	if state.YRCEnabled && len(state.YRCLines) > 0 {
		idx := state.YRCLineIndex
		if idx < 0 {
			idx = 0
		}
		if idx >= len(state.YRCLines) {
			idx = len(state.YRCLines) - 1
		}

		// Current line — with word data
		if idx < len(state.YRCLines) {
			line := state.YRCLines[idx]
			var sb strings.Builder
			words := make([]desktop_lyrics.LyricWord, len(line.Words))
			for i, w := range line.Words {
				sb.WriteString(w.Word)
				words[i] = desktop_lyrics.LyricWord{
					Word:      w.Word,
					StartTime: w.StartTime,
					EndTime:   w.EndTime,
				}
			}
			curLine = desktop_lyrics.LyricLine{Text: buildText(sb.String(), line.TranslatedLyric), Words: words}
		}

		// Next line
		nextIdx := idx + 1
		if nextIdx < len(state.YRCLines) {
			line := state.YRCLines[nextIdx]
			var sb strings.Builder
			nextWords := make([]desktop_lyrics.LyricWord, len(line.Words))
			for i, w := range line.Words {
				sb.WriteString(w.Word)
				nextWords[i] = desktop_lyrics.LyricWord{
					Word:      w.Word,
					StartTime: w.StartTime,
					EndTime:   w.EndTime,
				}
			}
			nextLine = desktop_lyrics.LyricLine{Text: buildText(sb.String(), line.TranslatedLyric), Words: nextWords}
		}

		return curLine, nextLine, idx

	} else if len(state.Fragments) > 0 {
		idx := state.CurrentIndex
		if idx < 0 {
			idx = 0
		}
		if idx >= len(state.Fragments) {
			idx = len(state.Fragments) - 1
		}

		// Current line — plain text (no word data for LRC)
		if idx < len(state.Fragments) {
			f := state.Fragments[idx]
			trans := ""
			if state.ShowTranslation {
				trans = state.TranslatedFragments[f.StartTimeMs]
			}
			curLine = desktop_lyrics.LyricLine{Text: buildText(f.Content, trans)}
		}

		// Next line
		if idx+1 < len(state.Fragments) {
			f := state.Fragments[idx+1]
			trans := ""
			if state.ShowTranslation {
				trans = state.TranslatedFragments[f.StartTimeMs]
			}
			nextLine = desktop_lyrics.LyricLine{Text: buildText(f.Content, trans)}
		}

		return curLine, nextLine, idx
	}

	return desktop_lyrics.LyricLine{}, desktop_lyrics.LyricLine{}, -1
}

// User returns the shared user slot dereferenced once; nil when the slot or
// its value is missing.
func (p *Player) User() *structs.User {
	if p.user == nil || *p.user == nil {
		return nil
	}
	return *p.user
}
