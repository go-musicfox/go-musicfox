//go:build darwin
// +build darwin

package avcore

import (
	core "github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"unsafe"
)

/*
#cgo CFLAGS: -x objective-c -Wno-everything
#cgo LDFLAGS: -lobjc -framework AVFoundation
#define __OBJC2__ 1
#include <objc/message.h>
#include <stdlib.h>

#include <AVFoundation/AVFoundation.h>

bool avcore_convertObjCBool(BOOL b) {
	if (b) { return true; }
	return false;
}


void* AVAsset_type_alloc() {
	return [AVAsset
		alloc];
}
void* AVAsset_type_assetWithURL_(void* URL) {
	return [AVAsset
		assetWithURL: URL];
}
void* AVPlayerItem_type_alloc() {
	return [AVPlayerItem
		alloc];
}
void* AVPlayerItem_type_playerItemWithURL_(void* URL) {
	return [AVPlayerItem
		playerItemWithURL: URL];
}
void* AVPlayerItem_type_playerItemWithAsset_(void* asset) {
	return [AVPlayerItem
		playerItemWithAsset: asset];
}
void* AVPlayerItem_type_playerItemWithAsset_automaticallyLoadedAssetKeys_(void* asset, void* automaticallyLoadedAssetKeys) {
	return [AVPlayerItem
		playerItemWithAsset: asset
		automaticallyLoadedAssetKeys: automaticallyLoadedAssetKeys];
}
void* AVPlayer_type_alloc() {
	return [AVPlayer
		alloc];
}
void* AVPlayer_type_playerWithURL_(void* URL) {
	return [AVPlayer
		playerWithURL: URL];
}
void* AVPlayer_type_playerWithPlayerItem_(void* item) {
	return [AVPlayer
		playerWithPlayerItem: item];
}
BOOL AVPlayer_type_eligibleForHDRPlayback() {
	return [AVPlayer
		eligibleForHDRPlayback];
}
void* AVQueuePlayer_type_alloc() {
	return [AVQueuePlayer
		alloc];
}
void* AVQueuePlayer_type_queuePlayerWithItems_(void* items) {
	return [AVQueuePlayer
		queuePlayerWithItems: items];
}


void AVAsset_inst_cancelLoading(void *id) {
	[(AVAsset*)id
		cancelLoading];
}

void* AVAsset_inst_init(void *id) {
	return [(AVAsset*)id
		init];
}

CMTime AVAsset_inst_duration(void *id) {
	return [(AVAsset*)id
		duration];
}

BOOL AVAsset_inst_providesPreciseDurationAndTiming(void *id) {
	return [(AVAsset*)id
		providesPreciseDurationAndTiming];
}

CMTime AVAsset_inst_minimumTimeOffsetFromLive(void *id) {
	return [(AVAsset*)id
		minimumTimeOffsetFromLive];
}

void* AVAsset_inst_tracks(void *id) {
	return [(AVAsset*)id
		tracks];
}

void* AVAsset_inst_trackGroups(void *id) {
	return [(AVAsset*)id
		trackGroups];
}

void* AVAsset_inst_metadata(void *id) {
	return [(AVAsset*)id
		metadata];
}

void* AVAsset_inst_commonMetadata(void *id) {
	return [(AVAsset*)id
		commonMetadata];
}

void* AVAsset_inst_availableMetadataFormats(void *id) {
	return [(AVAsset*)id
		availableMetadataFormats];
}

void* AVAsset_inst_lyrics(void *id) {
	return [(AVAsset*)id
		lyrics];
}

BOOL AVAsset_inst_isPlayable(void *id) {
	return [(AVAsset*)id
		isPlayable];
}

BOOL AVAsset_inst_isExportable(void *id) {
	return [(AVAsset*)id
		isExportable];
}

BOOL AVAsset_inst_isReadable(void *id) {
	return [(AVAsset*)id
		isReadable];
}

BOOL AVAsset_inst_isComposable(void *id) {
	return [(AVAsset*)id
		isComposable];
}

BOOL AVAsset_inst_isCompatibleWithAirPlayVideo(void *id) {
	return [(AVAsset*)id
		isCompatibleWithAirPlayVideo];
}

float AVAsset_inst_preferredRate(void *id) {
	return [(AVAsset*)id
		preferredRate];
}

float AVAsset_inst_preferredVolume(void *id) {
	return [(AVAsset*)id
		preferredVolume];
}

void* AVAsset_inst_allMediaSelections(void *id) {
	return [(AVAsset*)id
		allMediaSelections];
}

void* AVAsset_inst_availableMediaCharacteristicsWithMediaSelectionOptions(void *id) {
	return [(AVAsset*)id
		availableMediaCharacteristicsWithMediaSelectionOptions];
}

void* AVAsset_inst_availableChapterLocales(void *id) {
	return [(AVAsset*)id
		availableChapterLocales];
}

BOOL AVAsset_inst_hasProtectedContent(void *id) {
	return [(AVAsset*)id
		hasProtectedContent];
}

BOOL AVAsset_inst_canContainFragments(void *id) {
	return [(AVAsset*)id
		canContainFragments];
}

BOOL AVAsset_inst_containsFragments(void *id) {
	return [(AVAsset*)id
		containsFragments];
}

CMTime AVAsset_inst_overallDurationHint(void *id) {
	return [(AVAsset*)id
		overallDurationHint];
}

void* AVPlayerItem_inst_initWithURL_(void *id, void* URL) {
	return [(AVPlayerItem*)id
		initWithURL: URL];
}

void* AVPlayerItem_inst_initWithAsset_(void *id, void* asset) {
	return [(AVPlayerItem*)id
		initWithAsset: asset];
}

void* AVPlayerItem_inst_initWithAsset_automaticallyLoadedAssetKeys_(void *id, void* asset, void* automaticallyLoadedAssetKeys) {
	return [(AVPlayerItem*)id
		initWithAsset: asset
		automaticallyLoadedAssetKeys: automaticallyLoadedAssetKeys];
}

void AVPlayerItem_inst_stepByCount_(void *id, long stepCount) {
	[(AVPlayerItem*)id
		stepByCount: stepCount];
}

void AVPlayerItem_inst_cancelPendingSeeks(void *id) {
	[(AVPlayerItem*)id
		cancelPendingSeeks];
}

CMTime AVPlayerItem_inst_currentTime(void *id) {
	return [(AVPlayerItem*)id
		currentTime];
}

void* AVPlayerItem_inst_currentDate(void *id) {
	return [(AVPlayerItem*)id
		currentDate];
}

void AVPlayerItem_inst_cancelContentAuthorizationRequest(void *id) {
	[(AVPlayerItem*)id
		cancelContentAuthorizationRequest];
}

//void* AVPlayerItem_inst_init(void *id) {
//	return [(AVPlayerItem*)id
//		init];
//}

void* AVPlayerItem_inst_tracks(void *id) {
	return [(AVPlayerItem*)id
		tracks];
}

void* AVPlayerItem_inst_error(void *id) {
	return [(AVPlayerItem*)id
		error];
}

BOOL AVPlayerItem_inst_canPlayReverse(void *id) {
	return [(AVPlayerItem*)id
		canPlayReverse];
}

BOOL AVPlayerItem_inst_canPlayFastForward(void *id) {
	return [(AVPlayerItem*)id
		canPlayFastForward];
}

BOOL AVPlayerItem_inst_canPlayFastReverse(void *id) {
	return [(AVPlayerItem*)id
		canPlayFastReverse];
}

BOOL AVPlayerItem_inst_canPlaySlowForward(void *id) {
	return [(AVPlayerItem*)id
		canPlaySlowForward];
}

BOOL AVPlayerItem_inst_canPlaySlowReverse(void *id) {
	return [(AVPlayerItem*)id
		canPlaySlowReverse];
}

CMTime AVPlayerItem_inst_forwardPlaybackEndTime(void *id) {
	return [(AVPlayerItem*)id
		forwardPlaybackEndTime];
}

void AVPlayerItem_inst_setForwardPlaybackEndTime_(void *id, CMTime value) {
	[(AVPlayerItem*)id
		setForwardPlaybackEndTime: value];
}

CMTime AVPlayerItem_inst_reversePlaybackEndTime(void *id) {
	return [(AVPlayerItem*)id
		reversePlaybackEndTime];
}

void AVPlayerItem_inst_setReversePlaybackEndTime_(void *id, CMTime value) {
	[(AVPlayerItem*)id
		setReversePlaybackEndTime: value];
}

BOOL AVPlayerItem_inst_canStepForward(void *id) {
	return [(AVPlayerItem*)id
		canStepForward];
}

BOOL AVPlayerItem_inst_canStepBackward(void *id) {
	return [(AVPlayerItem*)id
		canStepBackward];
}

BOOL AVPlayerItem_inst_startsOnFirstEligibleVariant(void *id) {
	return [(AVPlayerItem*)id
		startsOnFirstEligibleVariant];
}

void AVPlayerItem_inst_setStartsOnFirstEligibleVariant_(void *id, BOOL value) {
	[(AVPlayerItem*)id
		setStartsOnFirstEligibleVariant: value];
}

CMTime AVPlayerItem_inst_duration(void *id) {
	return [(AVPlayerItem*)id
		duration];
}

void* AVPlayerItem_inst_loadedTimeRanges(void *id) {
	return [(AVPlayerItem*)id
		loadedTimeRanges];
}

void* AVPlayerItem_inst_seekableTimeRanges(void *id) {
	return [(AVPlayerItem*)id
		seekableTimeRanges];
}

BOOL AVPlayerItem_inst_isPlaybackLikelyToKeepUp(void *id) {
	return [(AVPlayerItem*)id
		isPlaybackLikelyToKeepUp];
}

BOOL AVPlayerItem_inst_isPlaybackBufferFull(void *id) {
	return [(AVPlayerItem*)id
		isPlaybackBufferFull];
}

BOOL AVPlayerItem_inst_isPlaybackBufferEmpty(void *id) {
	return [(AVPlayerItem*)id
		isPlaybackBufferEmpty];
}

void* AVPlayerItem_inst_textStyleRules(void *id) {
	return [(AVPlayerItem*)id
		textStyleRules];
}

void AVPlayerItem_inst_setTextStyleRules_(void *id, void* value) {
	[(AVPlayerItem*)id
		setTextStyleRules: value];
}

BOOL AVPlayerItem_inst_automaticallyPreservesTimeOffsetFromLive(void *id) {
	return [(AVPlayerItem*)id
		automaticallyPreservesTimeOffsetFromLive];
}

void AVPlayerItem_inst_setAutomaticallyPreservesTimeOffsetFromLive_(void *id, BOOL value) {
	[(AVPlayerItem*)id
		setAutomaticallyPreservesTimeOffsetFromLive: value];
}

CMTime AVPlayerItem_inst_recommendedTimeOffsetFromLive(void *id) {
	return [(AVPlayerItem*)id
		recommendedTimeOffsetFromLive];
}

CMTime AVPlayerItem_inst_configuredTimeOffsetFromLive(void *id) {
	return [(AVPlayerItem*)id
		configuredTimeOffsetFromLive];
}

void AVPlayerItem_inst_setConfiguredTimeOffsetFromLive_(void *id, CMTime value) {
	[(AVPlayerItem*)id
		setConfiguredTimeOffsetFromLive: value];
}

NSSize AVPlayerItem_inst_presentationSize(void *id) {
	return [(AVPlayerItem*)id
		presentationSize];
}

NSSize AVPlayerItem_inst_preferredMaximumResolution(void *id) {
	return [(AVPlayerItem*)id
		preferredMaximumResolution];
}

void AVPlayerItem_inst_setPreferredMaximumResolution_(void *id, NSSize value) {
	[(AVPlayerItem*)id
		setPreferredMaximumResolution: value];
}

void* AVPlayerItem_inst_nowPlayingInfo(void *id) {
	return [(AVPlayerItem*)id
		nowPlayingInfo];
}

void AVPlayerItem_inst_setNowPlayingInfo_(void *id, void* value) {
	[(AVPlayerItem*)id
		setNowPlayingInfo: value];
}

BOOL AVPlayerItem_inst_appliesPerFrameHDRDisplayMetadata(void *id) {
	return [(AVPlayerItem*)id
		appliesPerFrameHDRDisplayMetadata];
}

void AVPlayerItem_inst_setAppliesPerFrameHDRDisplayMetadata_(void *id, BOOL value) {
	[(AVPlayerItem*)id
		setAppliesPerFrameHDRDisplayMetadata: value];
}

void* AVPlayerItem_inst_customVideoCompositor(void *id) {
	return [(AVPlayerItem*)id
		customVideoCompositor];
}

BOOL AVPlayerItem_inst_seekingWaitsForVideoCompositionRendering(void *id) {
	return [(AVPlayerItem*)id
		seekingWaitsForVideoCompositionRendering];
}

void AVPlayerItem_inst_setSeekingWaitsForVideoCompositionRendering_(void *id, BOOL value) {
	[(AVPlayerItem*)id
		setSeekingWaitsForVideoCompositionRendering: value];
}

void* AVPlayerItem_inst_outputs(void *id) {
	return [(AVPlayerItem*)id
		outputs];
}

void* AVPlayerItem_inst_mediaDataCollectors(void *id) {
	return [(AVPlayerItem*)id
		mediaDataCollectors];
}

double AVPlayerItem_inst_preferredPeakBitRate(void *id) {
	return [(AVPlayerItem*)id
		preferredPeakBitRate];
}

void AVPlayerItem_inst_setPreferredPeakBitRate_(void *id, double value) {
	[(AVPlayerItem*)id
		setPreferredPeakBitRate: value];
}

NSTimeInterval AVPlayerItem_inst_preferredForwardBufferDuration(void *id) {
	return [(AVPlayerItem*)id
		preferredForwardBufferDuration];
}

void AVPlayerItem_inst_setPreferredForwardBufferDuration_(void *id, NSTimeInterval value) {
	[(AVPlayerItem*)id
		setPreferredForwardBufferDuration: value];
}

BOOL AVPlayerItem_inst_canUseNetworkResourcesForLiveStreamingWhilePaused(void *id) {
	return [(AVPlayerItem*)id
		canUseNetworkResourcesForLiveStreamingWhilePaused];
}

void AVPlayerItem_inst_setCanUseNetworkResourcesForLiveStreamingWhilePaused_(void *id, BOOL value) {
	[(AVPlayerItem*)id
		setCanUseNetworkResourcesForLiveStreamingWhilePaused: value];
}

BOOL AVPlayerItem_inst_isContentAuthorizedForPlayback(void *id) {
	return [(AVPlayerItem*)id
		isContentAuthorizedForPlayback];
}

BOOL AVPlayerItem_inst_isAuthorizationRequiredForPlayback(void *id) {
	return [(AVPlayerItem*)id
		isAuthorizationRequiredForPlayback];
}

BOOL AVPlayerItem_inst_isApplicationAuthorizedForPlayback(void *id) {
	return [(AVPlayerItem*)id
		isApplicationAuthorizedForPlayback];
}

void* AVPlayerItem_inst_asset(void *id) {
	return [(AVPlayerItem*)id
		asset];
}

void* AVPlayerItem_inst_automaticallyLoadedAssetKeys(void *id) {
	return [(AVPlayerItem*)id
		automaticallyLoadedAssetKeys];
}

void* AVPlayer_inst_initWithURL_(void *id, void* URL) {
	return [(AVPlayer*)id
		initWithURL: URL];
}

void* AVPlayer_inst_initWithPlayerItem_(void *id, void* item) {
	return [(AVPlayer*)id
		initWithPlayerItem: item];
}

void AVPlayer_inst_replaceCurrentItemWithPlayerItem_(void *id, void* item) {
	[(AVPlayer*)id
		replaceCurrentItemWithPlayerItem: item];
}

void AVPlayer_inst_play(void *id) {
	[(AVPlayer*)id
		play];
}

void AVPlayer_inst_pause(void *id) {
	[(AVPlayer*)id
		pause];
}

CMTime AVPlayer_inst_currentTime(void *id) {
	return [(AVPlayer*)id
		currentTime];
}

void AVPlayer_inst_removeTimeObserver_(void *id, void* observer) {
	[(AVPlayer*)id
		removeTimeObserver: observer];
}

void AVPlayer_inst_seekToTime_(void *id, CMTime time) {
	[(AVPlayer*)id
		seekToTime: time];
}

void AVPlayer_inst_seekToTime_toleranceBefore_toleranceAfter_(void *id, CMTime time, CMTime toleranceBefore, CMTime toleranceAfter) {
	[(AVPlayer*)id
		seekToTime: time
		toleranceBefore: toleranceBefore
		toleranceAfter: toleranceAfter];
}

void AVPlayer_inst_seekToDate_(void *id, void* date) {
	[(AVPlayer*)id
		seekToDate: date];
}

void AVPlayer_inst_playImmediatelyAtRate_(void *id, float rate) {
	[(AVPlayer*)id
		playImmediatelyAtRate: rate];
}

void AVPlayer_inst_setRate_time_atHostTime_(void *id, float rate, CMTime itemTime, CMTime hostClockTime) {
	[(AVPlayer*)id
		setRate: rate
		time: itemTime
		atHostTime: hostClockTime];
}

void AVPlayer_inst_cancelPendingPrerolls(void *id) {
	[(AVPlayer*)id
		cancelPendingPrerolls];
}

void* AVPlayer_inst_init(void *id) {
	return [(AVPlayer*)id
		init];
}

void* AVPlayer_inst_currentItem(void *id) {
	return [(AVPlayer*)id
		currentItem];
}

long AVPlayer_inst_status(void *id) {
	return [(AVPlayer*)id
		status];
}

void* AVPlayer_inst_error(void *id) {
	return [(AVPlayer*)id
		error];
}

float AVPlayer_inst_rate(void *id) {
	return [(AVPlayer*)id
		rate];
}

void AVPlayer_inst_setRate_(void *id, float value) {
	[(AVPlayer*)id
		setRate: value];
}

BOOL AVPlayer_inst_automaticallyWaitsToMinimizeStalling(void *id) {
	return [(AVPlayer*)id
		automaticallyWaitsToMinimizeStalling];
}

void AVPlayer_inst_setAutomaticallyWaitsToMinimizeStalling_(void *id, BOOL value) {
	[(AVPlayer*)id
		setAutomaticallyWaitsToMinimizeStalling: value];
}

long AVPlayer_inst_actionAtItemEnd(void *id) {
	return [(AVPlayer*)id
		actionAtItemEnd];
}

void AVPlayer_inst_setActionAtItemEnd_(void *id, long value) {
	[(AVPlayer*)id
		setActionAtItemEnd: value];
}

BOOL AVPlayer_inst_appliesMediaSelectionCriteriaAutomatically(void *id) {
	return [(AVPlayer*)id
		appliesMediaSelectionCriteriaAutomatically];
}

void AVPlayer_inst_setAppliesMediaSelectionCriteriaAutomatically_(void *id, BOOL value) {
	[(AVPlayer*)id
		setAppliesMediaSelectionCriteriaAutomatically: value];
}

float AVPlayer_inst_volume(void *id) {
	return [(AVPlayer*)id
		volume];
}

void AVPlayer_inst_setVolume_(void *id, float value) {
	[(AVPlayer*)id
		setVolume: value];
}

BOOL AVPlayer_inst_isMuted(void *id) {
	return [(AVPlayer*)id
		isMuted];
}

void AVPlayer_inst_setMuted_(void *id, BOOL value) {
	[(AVPlayer*)id
		setMuted: value];
}

BOOL AVPlayer_inst_allowsExternalPlayback(void *id) {
	return [(AVPlayer*)id
		allowsExternalPlayback];
}

void AVPlayer_inst_setAllowsExternalPlayback_(void *id, BOOL value) {
	[(AVPlayer*)id
		setAllowsExternalPlayback: value];
}

BOOL AVPlayer_inst_isExternalPlaybackActive(void *id) {
	return [(AVPlayer*)id
		isExternalPlaybackActive];
}

BOOL AVPlayer_inst_preventsDisplaySleepDuringVideoPlayback(void *id) {
	return [(AVPlayer*)id
		preventsDisplaySleepDuringVideoPlayback];
}

void AVPlayer_inst_setPreventsDisplaySleepDuringVideoPlayback_(void *id, BOOL value) {
	[(AVPlayer*)id
		setPreventsDisplaySleepDuringVideoPlayback: value];
}

BOOL AVPlayer_inst_outputObscuredDueToInsufficientExternalProtection(void *id) {
	return [(AVPlayer*)id
		outputObscuredDueToInsufficientExternalProtection];
}

void* AVPlayer_inst_audioOutputDeviceUniqueID(void *id) {
	return [(AVPlayer*)id
		audioOutputDeviceUniqueID];
}

void AVPlayer_inst_setAudioOutputDeviceUniqueID_(void *id, void* value) {
	[(AVPlayer*)id
		setAudioOutputDeviceUniqueID: value];
}

void* AVQueuePlayer_inst_initWithItems_(void *id, void* items) {
	return [(AVQueuePlayer*)id
		initWithItems: items];
}

void* AVQueuePlayer_inst_items(void *id) {
	return [(AVQueuePlayer*)id
		items];
}

void AVQueuePlayer_inst_advanceToNextItem(void *id) {
	[(AVQueuePlayer*)id
		advanceToNextItem];
}

BOOL AVQueuePlayer_inst_canInsertItem_afterItem_(void *id, void* item, void* afterItem) {
	return [(AVQueuePlayer*)id
		canInsertItem: item
		afterItem: afterItem];
}

void AVQueuePlayer_inst_insertItem_afterItem_(void *id, void* item, void* afterItem) {
	[(AVQueuePlayer*)id
		insertItem: item
		afterItem: afterItem];
}

void AVQueuePlayer_inst_removeItem_(void *id, void* item) {
	[(AVQueuePlayer*)id
		removeItem: item];
}

void AVQueuePlayer_inst_removeAllItems(void *id) {
	[(AVQueuePlayer*)id
		removeAllItems];
}

void* AVQueuePlayer_inst_init(void *id) {
	return [(AVQueuePlayer*)id
		init];
}


BOOL avcore_objc_bool_true = YES;
BOOL avcore_objc_bool_false = NO;

*/
import "C"

func convertObjCBoolToGo(b C.BOOL) bool {
	// NOTE: the prefix here is used to namespace these since the linker will
	// otherwise report a "duplicate symbol" because the C functions have the
	// same name.
	return bool(C.avcore_convertObjCBool(b))
}

func convertToObjCBool(b bool) C.BOOL {
	if b {
		return C.avcore_objc_bool_true
	}
	return C.avcore_objc_bool_false
}

func AVAsset_alloc() (
	r0 AVAsset,
) {
	ret := C.AVAsset_type_alloc()
	r0 = AVAsset_fromPointer(ret)
	return
}

func AVAsset_assetWithURL_(
	URL core.NSURLRef,
) (
	r0 AVAsset,
) {
	ret := C.AVAsset_type_assetWithURL_(
		objc.RefPointer(URL),
	)
	r0 = AVAsset_fromPointer(ret)
	return
}

func AVPlayerItem_alloc() (
	r0 AVPlayerItem,
) {
	ret := C.AVPlayerItem_type_alloc()
	r0 = AVPlayerItem_fromPointer(ret)
	return
}

func AVPlayerItem_playerItemWithURL_(
	URL core.NSURLRef,
) (
	r0 AVPlayerItem,
) {
	ret := C.AVPlayerItem_type_playerItemWithURL_(
		objc.RefPointer(URL),
	)
	r0 = AVPlayerItem_fromPointer(ret)
	return
}

func AVPlayerItem_playerItemWithAsset_(
	asset AVAssetRef,
) (
	r0 AVPlayerItem,
) {
	ret := C.AVPlayerItem_type_playerItemWithAsset_(
		objc.RefPointer(asset),
	)
	r0 = AVPlayerItem_fromPointer(ret)
	return
}

func AVPlayerItem_playerItemWithAsset_automaticallyLoadedAssetKeys_(
	asset AVAssetRef,
	automaticallyLoadedAssetKeys core.NSArrayRef,
) (
	r0 AVPlayerItem,
) {
	ret := C.AVPlayerItem_type_playerItemWithAsset_automaticallyLoadedAssetKeys_(
		objc.RefPointer(asset),
		objc.RefPointer(automaticallyLoadedAssetKeys),
	)
	r0 = AVPlayerItem_fromPointer(ret)
	return
}

func AVPlayer_alloc() (
	r0 AVPlayer,
) {
	ret := C.AVPlayer_type_alloc()
	r0 = AVPlayer_fromPointer(ret)
	return
}

func AVPlayer_playerWithURL_(
	URL core.NSURLRef,
) (
	r0 AVPlayer,
) {
	ret := C.AVPlayer_type_playerWithURL_(
		objc.RefPointer(URL),
	)
	r0 = AVPlayer_fromPointer(ret)
	return
}

func AVPlayer_playerWithPlayerItem_(
	item AVPlayerItemRef,
) (
	r0 AVPlayer,
) {
	ret := C.AVPlayer_type_playerWithPlayerItem_(
		objc.RefPointer(item),
	)
	r0 = AVPlayer_fromPointer(ret)
	return
}

func AVPlayer_eligibleForHDRPlayback() (
	r0 bool,
) {
	ret := C.AVPlayer_type_eligibleForHDRPlayback()
	r0 = convertObjCBoolToGo(ret)
	return
}

func AVQueuePlayer_alloc() (
	r0 AVQueuePlayer,
) {
	ret := C.AVQueuePlayer_type_alloc()
	r0 = AVQueuePlayer_fromPointer(ret)
	return
}

func AVQueuePlayer_queuePlayerWithItems_(
	items core.NSArrayRef,
) (
	r0 AVQueuePlayer,
) {
	ret := C.AVQueuePlayer_type_queuePlayerWithItems_(
		objc.RefPointer(items),
	)
	r0 = AVQueuePlayer_fromPointer(ret)
	return
}

type AVAssetRef interface {
	Pointer() uintptr
	Init_asAVAsset() AVAsset
}

type gen_AVAsset struct {
	objc.Object
}

func AVAsset_fromPointer(ptr unsafe.Pointer) AVAsset {
	return AVAsset{gen_AVAsset{
		objc.Object_fromPointer(ptr),
	}}
}

func AVAsset_fromRef(ref objc.Ref) AVAsset {
	return AVAsset_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_AVAsset) CancelLoading() {
	C.AVAsset_inst_cancelLoading(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_AVAsset) Init_asAVAsset() (
	r0 AVAsset,
) {
	ret := C.AVAsset_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = AVAsset_fromPointer(ret)
	return
}

func (x gen_AVAsset) Duration() (
	r0 core.CMTime,
) {
	ret := C.AVAsset_inst_duration(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVAsset) ProvidesPreciseDurationAndTiming() (
	r0 bool,
) {
	ret := C.AVAsset_inst_providesPreciseDurationAndTiming(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVAsset) MinimumTimeOffsetFromLive() (
	r0 core.CMTime,
) {
	ret := C.AVAsset_inst_minimumTimeOffsetFromLive(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVAsset) Tracks() (
	r0 core.NSArray,
) {
	ret := C.AVAsset_inst_tracks(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVAsset) TrackGroups() (
	r0 core.NSArray,
) {
	ret := C.AVAsset_inst_trackGroups(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVAsset) Metadata() (
	r0 core.NSArray,
) {
	ret := C.AVAsset_inst_metadata(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVAsset) CommonMetadata() (
	r0 core.NSArray,
) {
	ret := C.AVAsset_inst_commonMetadata(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVAsset) AvailableMetadataFormats() (
	r0 core.NSArray,
) {
	ret := C.AVAsset_inst_availableMetadataFormats(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVAsset) Lyrics() (
	r0 core.NSString,
) {
	ret := C.AVAsset_inst_lyrics(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSString_fromPointer(ret)
	return
}

func (x gen_AVAsset) IsPlayable() (
	r0 bool,
) {
	ret := C.AVAsset_inst_isPlayable(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVAsset) IsExportable() (
	r0 bool,
) {
	ret := C.AVAsset_inst_isExportable(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVAsset) IsReadable() (
	r0 bool,
) {
	ret := C.AVAsset_inst_isReadable(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVAsset) IsComposable() (
	r0 bool,
) {
	ret := C.AVAsset_inst_isComposable(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVAsset) IsCompatibleWithAirPlayVideo() (
	r0 bool,
) {
	ret := C.AVAsset_inst_isCompatibleWithAirPlayVideo(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVAsset) PreferredRate() (
	r0 float32,
) {
	ret := C.AVAsset_inst_preferredRate(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_AVAsset) PreferredVolume() (
	r0 float32,
) {
	ret := C.AVAsset_inst_preferredVolume(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_AVAsset) AllMediaSelections() (
	r0 core.NSArray,
) {
	ret := C.AVAsset_inst_allMediaSelections(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVAsset) AvailableMediaCharacteristicsWithMediaSelectionOptions() (
	r0 core.NSArray,
) {
	ret := C.AVAsset_inst_availableMediaCharacteristicsWithMediaSelectionOptions(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVAsset) AvailableChapterLocales() (
	r0 core.NSArray,
) {
	ret := C.AVAsset_inst_availableChapterLocales(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVAsset) HasProtectedContent() (
	r0 bool,
) {
	ret := C.AVAsset_inst_hasProtectedContent(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVAsset) CanContainFragments() (
	r0 bool,
) {
	ret := C.AVAsset_inst_canContainFragments(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVAsset) ContainsFragments() (
	r0 bool,
) {
	ret := C.AVAsset_inst_containsFragments(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVAsset) OverallDurationHint() (
	r0 core.CMTime,
) {
	ret := C.AVAsset_inst_overallDurationHint(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

type AVPlayerItemRef interface {
	Pointer() uintptr
	Init_asAVPlayerItem() AVPlayerItem
}

type gen_AVPlayerItem struct {
	objc.Object
}

func AVPlayerItem_fromPointer(ptr unsafe.Pointer) AVPlayerItem {
	return AVPlayerItem{gen_AVPlayerItem{
		objc.Object_fromPointer(ptr),
	}}
}

func AVPlayerItem_fromRef(ref objc.Ref) AVPlayerItem {
	return AVPlayerItem_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_AVPlayerItem) InitWithURL__asAVPlayerItem(
	URL core.NSURLRef,
) (
	r0 AVPlayerItem,
) {
	ret := C.AVPlayerItem_inst_initWithURL_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(URL),
	)
	r0 = AVPlayerItem_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) InitWithAsset__asAVPlayerItem(
	asset AVAssetRef,
) (
	r0 AVPlayerItem,
) {
	ret := C.AVPlayerItem_inst_initWithAsset_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(asset),
	)
	r0 = AVPlayerItem_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) InitWithAsset_automaticallyLoadedAssetKeys__asAVPlayerItem(
	asset AVAssetRef,
	automaticallyLoadedAssetKeys core.NSArrayRef,
) (
	r0 AVPlayerItem,
) {
	ret := C.AVPlayerItem_inst_initWithAsset_automaticallyLoadedAssetKeys_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(asset),
		objc.RefPointer(automaticallyLoadedAssetKeys),
	)
	r0 = AVPlayerItem_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) StepByCount_(
	stepCount core.NSInteger,
) {
	C.AVPlayerItem_inst_stepByCount_(
		unsafe.Pointer(x.Pointer()),
		C.long(stepCount),
	)
	return
}

func (x gen_AVPlayerItem) CancelPendingSeeks() {
	C.AVPlayerItem_inst_cancelPendingSeeks(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_AVPlayerItem) CurrentTime() (
	r0 core.CMTime,
) {
	ret := C.AVPlayerItem_inst_currentTime(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVPlayerItem) CurrentDate() (
	r0 core.NSDate,
) {
	ret := C.AVPlayerItem_inst_currentDate(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSDate_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) CancelContentAuthorizationRequest() {
	C.AVPlayerItem_inst_cancelContentAuthorizationRequest(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_AVPlayerItem) Init_asAVPlayerItem() (
	r0 AVPlayerItem,
) {
	//ret := C.AVPlayerItem_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = AVPlayerItem_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) Tracks() (
	r0 core.NSArray,
) {
	ret := C.AVPlayerItem_inst_tracks(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) Error() (
	r0 core.NSError,
) {
	ret := C.AVPlayerItem_inst_error(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSError_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) CanPlayReverse() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_canPlayReverse(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) CanPlayFastForward() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_canPlayFastForward(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) CanPlayFastReverse() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_canPlayFastReverse(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) CanPlaySlowForward() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_canPlaySlowForward(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) CanPlaySlowReverse() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_canPlaySlowReverse(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) ForwardPlaybackEndTime() (
	r0 core.CMTime,
) {
	ret := C.AVPlayerItem_inst_forwardPlaybackEndTime(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVPlayerItem) SetForwardPlaybackEndTime_(
	value core.CMTime,
) {
	C.AVPlayerItem_inst_setForwardPlaybackEndTime_(
		unsafe.Pointer(x.Pointer()),
		*(*C.CMTime)(unsafe.Pointer(&value)),
	)
	return
}

func (x gen_AVPlayerItem) ReversePlaybackEndTime() (
	r0 core.CMTime,
) {
	ret := C.AVPlayerItem_inst_reversePlaybackEndTime(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVPlayerItem) SetReversePlaybackEndTime_(
	value core.CMTime,
) {
	C.AVPlayerItem_inst_setReversePlaybackEndTime_(
		unsafe.Pointer(x.Pointer()),
		*(*C.CMTime)(unsafe.Pointer(&value)),
	)
	return
}

func (x gen_AVPlayerItem) CanStepForward() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_canStepForward(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) CanStepBackward() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_canStepBackward(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) StartsOnFirstEligibleVariant() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_startsOnFirstEligibleVariant(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) SetStartsOnFirstEligibleVariant_(
	value bool,
) {
	C.AVPlayerItem_inst_setStartsOnFirstEligibleVariant_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayerItem) Duration() (
	r0 core.CMTime,
) {
	ret := C.AVPlayerItem_inst_duration(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVPlayerItem) LoadedTimeRanges() (
	r0 core.NSArray,
) {
	ret := C.AVPlayerItem_inst_loadedTimeRanges(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) SeekableTimeRanges() (
	r0 core.NSArray,
) {
	ret := C.AVPlayerItem_inst_seekableTimeRanges(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) IsPlaybackLikelyToKeepUp() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_isPlaybackLikelyToKeepUp(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) IsPlaybackBufferFull() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_isPlaybackBufferFull(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) IsPlaybackBufferEmpty() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_isPlaybackBufferEmpty(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) TextStyleRules() (
	r0 core.NSArray,
) {
	ret := C.AVPlayerItem_inst_textStyleRules(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) SetTextStyleRules_(
	value core.NSArrayRef,
) {
	C.AVPlayerItem_inst_setTextStyleRules_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_AVPlayerItem) AutomaticallyPreservesTimeOffsetFromLive() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_automaticallyPreservesTimeOffsetFromLive(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) SetAutomaticallyPreservesTimeOffsetFromLive_(
	value bool,
) {
	C.AVPlayerItem_inst_setAutomaticallyPreservesTimeOffsetFromLive_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayerItem) RecommendedTimeOffsetFromLive() (
	r0 core.CMTime,
) {
	ret := C.AVPlayerItem_inst_recommendedTimeOffsetFromLive(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVPlayerItem) ConfiguredTimeOffsetFromLive() (
	r0 core.CMTime,
) {
	ret := C.AVPlayerItem_inst_configuredTimeOffsetFromLive(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVPlayerItem) SetConfiguredTimeOffsetFromLive_(
	value core.CMTime,
) {
	C.AVPlayerItem_inst_setConfiguredTimeOffsetFromLive_(
		unsafe.Pointer(x.Pointer()),
		*(*C.CMTime)(unsafe.Pointer(&value)),
	)
	return
}

func (x gen_AVPlayerItem) PresentationSize() (
	r0 core.NSSize,
) {
	ret := C.AVPlayerItem_inst_presentationSize(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.NSSize)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVPlayerItem) PreferredMaximumResolution() (
	r0 core.NSSize,
) {
	ret := C.AVPlayerItem_inst_preferredMaximumResolution(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.NSSize)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVPlayerItem) SetPreferredMaximumResolution_(
	value core.NSSize,
) {
	C.AVPlayerItem_inst_setPreferredMaximumResolution_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSSize)(unsafe.Pointer(&value)),
	)
	return
}

func (x gen_AVPlayerItem) NowPlayingInfo() (
	r0 core.NSDictionary,
) {
	ret := C.AVPlayerItem_inst_nowPlayingInfo(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSDictionary_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) SetNowPlayingInfo_(
	value core.NSDictionaryRef,
) {
	C.AVPlayerItem_inst_setNowPlayingInfo_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_AVPlayerItem) AppliesPerFrameHDRDisplayMetadata() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_appliesPerFrameHDRDisplayMetadata(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) SetAppliesPerFrameHDRDisplayMetadata_(
	value bool,
) {
	C.AVPlayerItem_inst_setAppliesPerFrameHDRDisplayMetadata_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayerItem) CustomVideoCompositor() (
	r0 objc.Object,
) {
	ret := C.AVPlayerItem_inst_customVideoCompositor(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) SeekingWaitsForVideoCompositionRendering() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_seekingWaitsForVideoCompositionRendering(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) SetSeekingWaitsForVideoCompositionRendering_(
	value bool,
) {
	C.AVPlayerItem_inst_setSeekingWaitsForVideoCompositionRendering_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayerItem) Outputs() (
	r0 core.NSArray,
) {
	ret := C.AVPlayerItem_inst_outputs(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) MediaDataCollectors() (
	r0 core.NSArray,
) {
	ret := C.AVPlayerItem_inst_mediaDataCollectors(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) PreferredPeakBitRate() (
	r0 float64,
) {
	ret := C.AVPlayerItem_inst_preferredPeakBitRate(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float64(ret)
	return
}

func (x gen_AVPlayerItem) SetPreferredPeakBitRate_(
	value float64,
) {
	C.AVPlayerItem_inst_setPreferredPeakBitRate_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	return
}

func (x gen_AVPlayerItem) PreferredForwardBufferDuration() (
	r0 float64,
) {
	ret := C.AVPlayerItem_inst_preferredForwardBufferDuration(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float64(ret)
	return
}

func (x gen_AVPlayerItem) SetPreferredForwardBufferDuration_(
	value float64,
) {
	C.AVPlayerItem_inst_setPreferredForwardBufferDuration_(
		unsafe.Pointer(x.Pointer()),
		C.NSTimeInterval(value),
	)
	return
}

func (x gen_AVPlayerItem) CanUseNetworkResourcesForLiveStreamingWhilePaused() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_canUseNetworkResourcesForLiveStreamingWhilePaused(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) SetCanUseNetworkResourcesForLiveStreamingWhilePaused_(
	value bool,
) {
	C.AVPlayerItem_inst_setCanUseNetworkResourcesForLiveStreamingWhilePaused_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayerItem) IsContentAuthorizedForPlayback() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_isContentAuthorizedForPlayback(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) IsAuthorizationRequiredForPlayback() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_isAuthorizationRequiredForPlayback(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) IsApplicationAuthorizedForPlayback() (
	r0 bool,
) {
	ret := C.AVPlayerItem_inst_isApplicationAuthorizedForPlayback(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayerItem) Asset() (
	r0 AVAsset,
) {
	ret := C.AVPlayerItem_inst_asset(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = AVAsset_fromPointer(ret)
	return
}

func (x gen_AVPlayerItem) AutomaticallyLoadedAssetKeys() (
	r0 core.NSArray,
) {
	ret := C.AVPlayerItem_inst_automaticallyLoadedAssetKeys(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

type AVPlayerRef interface {
	Pointer() uintptr
	Init_asAVPlayer() AVPlayer
}

type gen_AVPlayer struct {
	objc.Object
}

func AVPlayer_fromPointer(ptr unsafe.Pointer) AVPlayer {
	return AVPlayer{gen_AVPlayer{
		objc.Object_fromPointer(ptr),
	}}
}

func AVPlayer_fromRef(ref objc.Ref) AVPlayer {
	return AVPlayer_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_AVPlayer) InitWithURL__asAVPlayer(
	URL core.NSURLRef,
) (
	r0 AVPlayer,
) {
	ret := C.AVPlayer_inst_initWithURL_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(URL),
	)
	r0 = AVPlayer_fromPointer(ret)
	return
}

func (x gen_AVPlayer) InitWithPlayerItem__asAVPlayer(
	item AVPlayerItemRef,
) (
	r0 AVPlayer,
) {
	ret := C.AVPlayer_inst_initWithPlayerItem_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(item),
	)
	r0 = AVPlayer_fromPointer(ret)
	return
}

func (x gen_AVPlayer) ReplaceCurrentItemWithPlayerItem_(
	item AVPlayerItemRef,
) {
	C.AVPlayer_inst_replaceCurrentItemWithPlayerItem_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(item),
	)
	return
}

func (x gen_AVPlayer) Play() {
	C.AVPlayer_inst_play(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_AVPlayer) Pause() {
	C.AVPlayer_inst_pause(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_AVPlayer) CurrentTime() (
	r0 core.CMTime,
) {
	ret := C.AVPlayer_inst_currentTime(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*core.CMTime)(unsafe.Pointer(&ret))
	return
}

func (x gen_AVPlayer) RemoveTimeObserver_(
	observer objc.Ref,
) {
	C.AVPlayer_inst_removeTimeObserver_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(observer),
	)
	return
}

func (x gen_AVPlayer) SeekToTime_(
	time core.CMTime,
) {
	C.AVPlayer_inst_seekToTime_(
		unsafe.Pointer(x.Pointer()),
		*(*C.CMTime)(unsafe.Pointer(&time)),
	)
	return
}

func (x gen_AVPlayer) SeekToTime_toleranceBefore_toleranceAfter_(
	time core.CMTime,
	toleranceBefore core.CMTime,
	toleranceAfter core.CMTime,
) {
	C.AVPlayer_inst_seekToTime_toleranceBefore_toleranceAfter_(
		unsafe.Pointer(x.Pointer()),
		*(*C.CMTime)(unsafe.Pointer(&time)),
		*(*C.CMTime)(unsafe.Pointer(&toleranceBefore)),
		*(*C.CMTime)(unsafe.Pointer(&toleranceAfter)),
	)
	return
}

func (x gen_AVPlayer) SeekToDate_(
	date core.NSDateRef,
) {
	C.AVPlayer_inst_seekToDate_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(date),
	)
	return
}

func (x gen_AVPlayer) PlayImmediatelyAtRate_(
	rate float32,
) {
	C.AVPlayer_inst_playImmediatelyAtRate_(
		unsafe.Pointer(x.Pointer()),
		C.float(rate),
	)
	return
}

func (x gen_AVPlayer) SetRate_time_atHostTime_(
	rate float32,
	itemTime core.CMTime,
	hostClockTime core.CMTime,
) {
	C.AVPlayer_inst_setRate_time_atHostTime_(
		unsafe.Pointer(x.Pointer()),
		C.float(rate),
		*(*C.CMTime)(unsafe.Pointer(&itemTime)),
		*(*C.CMTime)(unsafe.Pointer(&hostClockTime)),
	)
	return
}

func (x gen_AVPlayer) CancelPendingPrerolls() {
	C.AVPlayer_inst_cancelPendingPrerolls(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_AVPlayer) Init_asAVPlayer() (
	r0 AVPlayer,
) {
	ret := C.AVPlayer_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = AVPlayer_fromPointer(ret)
	return
}

func (x gen_AVPlayer) CurrentItem() (
	r0 AVPlayerItem,
) {
	ret := C.AVPlayer_inst_currentItem(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = AVPlayerItem_fromPointer(ret)
	return
}

func (x gen_AVPlayer) Status() (
	r0 core.NSInteger,
) {
	ret := C.AVPlayer_inst_status(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSInteger(ret)
	return
}

func (x gen_AVPlayer) Error() (
	r0 core.NSError,
) {
	ret := C.AVPlayer_inst_error(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSError_fromPointer(ret)
	return
}

func (x gen_AVPlayer) Rate() (
	r0 float32,
) {
	ret := C.AVPlayer_inst_rate(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_AVPlayer) SetRate_(
	value float32,
) {
	C.AVPlayer_inst_setRate_(
		unsafe.Pointer(x.Pointer()),
		C.float(value),
	)
	return
}

func (x gen_AVPlayer) AutomaticallyWaitsToMinimizeStalling() (
	r0 bool,
) {
	ret := C.AVPlayer_inst_automaticallyWaitsToMinimizeStalling(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayer) SetAutomaticallyWaitsToMinimizeStalling_(
	value bool,
) {
	C.AVPlayer_inst_setAutomaticallyWaitsToMinimizeStalling_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayer) ActionAtItemEnd() (
	r0 core.NSInteger,
) {
	ret := C.AVPlayer_inst_actionAtItemEnd(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSInteger(ret)
	return
}

func (x gen_AVPlayer) SetActionAtItemEnd_(
	value core.NSInteger,
) {
	C.AVPlayer_inst_setActionAtItemEnd_(
		unsafe.Pointer(x.Pointer()),
		C.long(value),
	)
	return
}

func (x gen_AVPlayer) AppliesMediaSelectionCriteriaAutomatically() (
	r0 bool,
) {
	ret := C.AVPlayer_inst_appliesMediaSelectionCriteriaAutomatically(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayer) SetAppliesMediaSelectionCriteriaAutomatically_(
	value bool,
) {
	C.AVPlayer_inst_setAppliesMediaSelectionCriteriaAutomatically_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayer) Volume() (
	r0 float32,
) {
	ret := C.AVPlayer_inst_volume(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_AVPlayer) SetVolume_(
	value float32,
) {
	C.AVPlayer_inst_setVolume_(
		unsafe.Pointer(x.Pointer()),
		C.float(value),
	)
	return
}

func (x gen_AVPlayer) IsMuted() (
	r0 bool,
) {
	ret := C.AVPlayer_inst_isMuted(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayer) SetMuted_(
	value bool,
) {
	C.AVPlayer_inst_setMuted_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayer) AllowsExternalPlayback() (
	r0 bool,
) {
	ret := C.AVPlayer_inst_allowsExternalPlayback(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayer) SetAllowsExternalPlayback_(
	value bool,
) {
	C.AVPlayer_inst_setAllowsExternalPlayback_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayer) IsExternalPlaybackActive() (
	r0 bool,
) {
	ret := C.AVPlayer_inst_isExternalPlaybackActive(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayer) PreventsDisplaySleepDuringVideoPlayback() (
	r0 bool,
) {
	ret := C.AVPlayer_inst_preventsDisplaySleepDuringVideoPlayback(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayer) SetPreventsDisplaySleepDuringVideoPlayback_(
	value bool,
) {
	C.AVPlayer_inst_setPreventsDisplaySleepDuringVideoPlayback_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_AVPlayer) OutputObscuredDueToInsufficientExternalProtection() (
	r0 bool,
) {
	ret := C.AVPlayer_inst_outputObscuredDueToInsufficientExternalProtection(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVPlayer) AudioOutputDeviceUniqueID() (
	r0 core.NSString,
) {
	ret := C.AVPlayer_inst_audioOutputDeviceUniqueID(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSString_fromPointer(ret)
	return
}

func (x gen_AVPlayer) SetAudioOutputDeviceUniqueID_(
	value core.NSStringRef,
) {
	C.AVPlayer_inst_setAudioOutputDeviceUniqueID_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

type AVQueuePlayerRef interface {
	Pointer() uintptr
	Init_asAVQueuePlayer() AVQueuePlayer
}

type gen_AVQueuePlayer struct {
	AVPlayer
}

func AVQueuePlayer_fromPointer(ptr unsafe.Pointer) AVQueuePlayer {
	return AVQueuePlayer{gen_AVQueuePlayer{
		AVPlayer_fromPointer(ptr),
	}}
}

func AVQueuePlayer_fromRef(ref objc.Ref) AVQueuePlayer {
	return AVQueuePlayer_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_AVQueuePlayer) InitWithItems_(
	items core.NSArrayRef,
) (
	r0 AVQueuePlayer,
) {
	ret := C.AVQueuePlayer_inst_initWithItems_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(items),
	)
	r0 = AVQueuePlayer_fromPointer(ret)
	return
}

func (x gen_AVQueuePlayer) Items() (
	r0 core.NSArray,
) {
	ret := C.AVQueuePlayer_inst_items(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_AVQueuePlayer) AdvanceToNextItem() {
	C.AVQueuePlayer_inst_advanceToNextItem(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_AVQueuePlayer) CanInsertItem_afterItem_(
	item AVPlayerItemRef,
	afterItem AVPlayerItemRef,
) (
	r0 bool,
) {
	ret := C.AVQueuePlayer_inst_canInsertItem_afterItem_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(item),
		objc.RefPointer(afterItem),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_AVQueuePlayer) InsertItem_afterItem_(
	item AVPlayerItemRef,
	afterItem AVPlayerItemRef,
) {
	C.AVQueuePlayer_inst_insertItem_afterItem_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(item),
		objc.RefPointer(afterItem),
	)
	return
}

func (x gen_AVQueuePlayer) RemoveItem_(
	item AVPlayerItemRef,
) {
	C.AVQueuePlayer_inst_removeItem_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(item),
	)
	return
}

func (x gen_AVQueuePlayer) RemoveAllItems() {
	C.AVQueuePlayer_inst_removeAllItems(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_AVQueuePlayer) Init_asAVQueuePlayer() (
	r0 AVQueuePlayer,
) {
	ret := C.AVQueuePlayer_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = AVQueuePlayer_fromPointer(ret)
	return
}
