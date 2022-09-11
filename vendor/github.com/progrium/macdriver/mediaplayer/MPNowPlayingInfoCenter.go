//go:build darwin
// +build darwin

package mediaplayer

import "github.com/progrium/macdriver/core"

type MPNowPlayingInfoCenter struct {
	gen_MPNowPlayingInfoCenter
}

const (
	MPNowPlayingPlaybackStateUnknown core.NSUInteger = iota
	MPNowPlayingPlaybackStatePlaying
	MPNowPlayingPlaybackStatePaused
	MPNowPlayingPlaybackStateStopped
	MPNowPlayingPlaybackStateInterrupted
)

const (
	MPRemoteCommandHandlerStatusSuccess                    core.NSInteger = 0
	MPRemoteCommandHandlerStatusNoSuchContent              core.NSInteger = 100
	MPRemoteCommandHandlerStatusNoActionableNowPlayingItem core.NSInteger = 110
	MPRemoteCommandHandlerStatusDeviceNotFound             core.NSInteger = 120
	MPRemoteCommandHandlerStatusCommandFailed              core.NSInteger = 200
)

const (
	MPNowPlayingInfoMediaTypeNone core.NSUInteger = iota
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

const (
	MPMediaTypeMusic        = 1 << 0
	MPMediaTypePodcast      = 1 << 1
	MPMediaTypeAudioBook    = 1 << 2
	MPMediaTypeAudioITunesU = 1 << 3
	MPMediaTypeAnyAudio     = 0x00ff

	MPMediaTypeMovie        = 1 << 8
	MPMediaTypeTVShow       = 1 << 9
	MPMediaTypeVideoPodcast = 1 << 10
	MPMediaTypeMusicVideo   = 1 << 11
	MPMediaTypeVideoITunesU = 1 << 12
	MPMediaTypeHomeVideo    = 1 << 13
	MPMediaTypeAnyVideo     = 0xff00

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
