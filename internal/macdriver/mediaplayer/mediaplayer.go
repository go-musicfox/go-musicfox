//go:build darwin

package mediaplayer

import (
	"sync"

	"github.com/ebitengine/purego"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

var importOnce sync.Once

func importFramework() {
	importOnce.Do(func() {
		_, err := purego.Dlopen("/System/Library/Frameworks/MediaPlayer.framework/MediaPlayer", purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
	})
}

type MPNowPlayingPlaybackState core.NSUInteger

const (
	MPNowPlayingPlaybackStateUnknown MPNowPlayingPlaybackState = iota
	MPNowPlayingPlaybackStatePlaying
	MPNowPlayingPlaybackStatePaused
	MPNowPlayingPlaybackStateStopped
	MPNowPlayingPlaybackStateInterrupted
)

type MPRemoteCommandHandlerStatus core.NSInteger

const (
	MPRemoteCommandHandlerStatusSuccess                    MPRemoteCommandHandlerStatus = 0
	MPRemoteCommandHandlerStatusNoSuchContent              MPRemoteCommandHandlerStatus = 100
	MPRemoteCommandHandlerStatusNoActionableNowPlayingItem MPRemoteCommandHandlerStatus = 110
	MPRemoteCommandHandlerStatusDeviceNotFound             MPRemoteCommandHandlerStatus = 120
	MPRemoteCommandHandlerStatusCommandFailed              MPRemoteCommandHandlerStatus = 200
)

type MPNowPlayingInfoMediaType core.NSUInteger

const (
	MPNowPlayingInfoMediaTypeNone MPNowPlayingInfoMediaType = iota
	MPNowPlayingInfoMediaTypeAudio
	MPNowPlayingInfoMediaTypeVideo
)

const (
	MPNowPlayingInfoPropertyElapsedPlaybackTime           = "MPNowPlayingInfoPropertyElapsedPlaybackTime"
	MPNowPlayingInfoPropertyPlaybackRate                  = "MPNowPlayingInfoPropertyPlaybackRate"
	MPNowPlayingInfoPropertyDefaultPlaybackRate           = "MPNowPlayingInfoPropertyDefaultPlaybackRate"
	MPNowPlayingInfoPropertyPlaybackQueueIndex            = "MPNowPlayingInfoPropertyPlaybackQueueIndex"
	MPNowPlayingInfoPropertyPlaybackQueueCount            = "MPNowPlayingInfoPropertyPlaybackQueueCount"
	MPNowPlayingInfoPropertyChapterNumber                 = "MPNowPlayingInfoPropertyChapterNumber"
	MPNowPlayingInfoPropertyChapterCount                  = "MPNowPlayingInfoPropertyChapterCount"
	MPNowPlayingInfoPropertyIsLiveStream                  = "MPNowPlayingInfoPropertyIsLiveStream"
	MPNowPlayingInfoPropertyAvailableLanguageOptions      = "MPNowPlayingInfoPropertyAvailableLanguageOptions"
	MPNowPlayingInfoPropertyCurrentLanguageOptions        = "MPNowPlayingInfoPropertyCurrentLanguageOptions"
	MPNowPlayingInfoCollectionIdentifier                  = "MPNowPlayingInfoCollectionIdentifier"
	MPNowPlayingInfoPropertyExternalContentIdentifier     = "MPNowPlayingInfoPropertyExternalContentIdentifier"
	MPNowPlayingInfoPropertyExternalUserProfileIdentifier = "MPNowPlayingInfoPropertyExternalUserProfileIdentifier"
	MPNowPlayingInfoPropertyServiceIdentifier             = "MPNowPlayingInfoPropertyServiceIdentifier"
	MPNowPlayingInfoPropertyPlaybackProgress              = "MPNowPlayingInfoPropertyPlaybackProgress"
	MPNowPlayingInfoPropertyMediaType                     = "MPNowPlayingInfoPropertyMediaType"
	MPNowPlayingInfoPropertyAssetURL                      = "MPNowPlayingInfoPropertyAssetURL"
	MPNowPlayingInfoPropertyCurrentPlaybackDate           = "MPNowPlayingInfoPropertyCurrentPlaybackDate"
)

type MPMediaType core.NSUInteger

const (
	MPMediaTypeMusic        MPMediaType = 1 << 0
	MPMediaTypePodcast      MPMediaType = 1 << 1
	MPMediaTypeAudioBook    MPMediaType = 1 << 2
	MPMediaTypeAudioITunesU MPMediaType = 1 << 3
	MPMediaTypeAnyAudio     MPMediaType = 0x00ff

	MPMediaTypeMovie        MPMediaType = 1 << 8
	MPMediaTypeTVShow       MPMediaType = 1 << 9
	MPMediaTypeVideoPodcast MPMediaType = 1 << 10
	MPMediaTypeMusicVideo   MPMediaType = 1 << 11
	MPMediaTypeVideoITunesU MPMediaType = 1 << 12
	MPMediaTypeHomeVideo    MPMediaType = 1 << 13
	MPMediaTypeAnyVideo     MPMediaType = 0xff00

	MPMediaTypeAny = MPMediaTypeAnyAudio | MPMediaTypeAnyVideo
)

const (
	MPMediaItemPropertyPersistentID            = "persistentID"
	MPMediaItemPropertyMediaType               = "mediaType"
	MPMediaItemPropertyTitle                   = "title"
	MPMediaItemPropertyAlbumTitle              = "albumTitle"
	MPMediaItemPropertyAlbumPersistentID       = "albumPersistentID"
	MPMediaItemPropertyArtist                  = "artist"
	MPMediaItemPropertyArtistPersistentID      = "artistPersistentID"
	MPMediaItemPropertyAlbumArtist             = "albumArtist"
	MPMediaItemPropertyAlbumArtistPersistentID = "albumArtistPersistentID"
	MPMediaItemPropertyGenre                   = "genre"
	MPMediaItemPropertyGenrePersistentID       = "genrePersistentID"
	MPMediaItemPropertyComposer                = "composer"
	MPMediaItemPropertyComposerPersistentID    = "composerPersistentID"
	MPMediaItemPropertyPlaybackDuration        = "playbackDuration"
	MPMediaItemPropertyAlbumTrackNumber        = "albumTrackNumber"
	MPMediaItemPropertyAlbumTrackCount         = "albumTrackCount"
	MPMediaItemPropertyDiscNumber              = "discNumber"
	MPMediaItemPropertyDiscCount               = "discCount"
	MPMediaItemPropertyArtwork                 = "artwork"
	MPMediaItemPropertyIsExplicit              = "explicitItem"
	MPMediaItemPropertyLyrics                  = "lyrics"
	MPMediaItemPropertyIsCompilation           = "compilation"
	MPMediaItemPropertyReleaseDate             = "releaseDate"
	MPMediaItemPropertyBeatsPerMinute          = "beatsPerMinute"
	MPMediaItemPropertyComments                = "comments"
	MPMediaItemPropertyAssetURL                = "assetURL"
	MPMediaItemPropertyIsCloudItem             = "cloudItem"
	MPMediaItemPropertyHasProtectedAsset       = "protectedAsset"
	MPMediaItemPropertyPodcastTitle            = "podcastTitle"
	MPMediaItemPropertyPodcastPersistentID     = "podcastPersistentID"
	MPMediaItemPropertyPlayCount               = "playCount"
	MPMediaItemPropertySkipCount               = "skipCount"
	MPMediaItemPropertyRating                  = "rating"
	MPMediaItemPropertyLastPlayedDate          = "lastPlayedDate"
	MPMediaItemPropertyUserGrouping            = "userGrouping"
	MPMediaItemPropertyBookmarkTime            = "bookmarkTime"
	MPMediaItemPropertyDateAdded               = "dateAdded"
	MPMediaItemPropertyPlaybackStoreID         = "playbackStoreID"
	MPMediaItemPropertyIsPreorder              = "preorder"
)
