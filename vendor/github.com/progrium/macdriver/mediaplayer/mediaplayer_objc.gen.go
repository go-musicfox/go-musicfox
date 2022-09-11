//go:build darwin
// +build darwin

package mediaplayer

import (
	core "github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"unsafe"
)

/*
#cgo CFLAGS: -x objective-c -Wno-everything
#cgo LDFLAGS: -lobjc -framework MediaPlayer
#define __OBJC2__ 1
#include <objc/message.h>
#include <stdlib.h>

#include <MediaPlayer/MediaPlayer.h>

bool mediaplayer_convertObjCBool(BOOL b) {
	if (b) { return true; }
	return false;
}


void* MPNowPlayingInfoCenter_type_alloc() {
	return [MPNowPlayingInfoCenter
		alloc];
}
void* MPNowPlayingInfoCenter_type_defaultCenter() {
	return [MPNowPlayingInfoCenter
		defaultCenter];
}
void* MPRemoteCommandEvent_type_alloc() {
	return [MPRemoteCommandEvent
		alloc];
}
void* MPChangePlaybackPositionCommandEvent_type_alloc() {
	return [MPChangePlaybackPositionCommandEvent
		alloc];
}
void* MPChangeLanguageOptionCommandEvent_type_alloc() {
	return [MPChangeLanguageOptionCommandEvent
		alloc];
}
void* MPChangePlaybackRateCommandEvent_type_alloc() {
	return [MPChangePlaybackRateCommandEvent
		alloc];
}
void* MPChangeRepeatModeCommandEvent_type_alloc() {
	return [MPChangeRepeatModeCommandEvent
		alloc];
}
void* MPChangeShuffleModeCommandEvent_type_alloc() {
	return [MPChangeShuffleModeCommandEvent
		alloc];
}
void* MPSeekCommandEvent_type_alloc() {
	return [MPSeekCommandEvent
		alloc];
}
void* MPSkipIntervalCommandEvent_type_alloc() {
	return [MPSkipIntervalCommandEvent
		alloc];
}
void* MPRemoteCommand_type_alloc() {
	return [MPRemoteCommand
		alloc];
}
void* MPChangeShuffleModeCommand_type_alloc() {
	return [MPChangeShuffleModeCommand
		alloc];
}
void* MPChangeRepeatModeCommand_type_alloc() {
	return [MPChangeRepeatModeCommand
		alloc];
}
void* MPChangePlaybackPositionCommand_type_alloc() {
	return [MPChangePlaybackPositionCommand
		alloc];
}
void* MPChangePlaybackRateCommand_type_alloc() {
	return [MPChangePlaybackRateCommand
		alloc];
}
void* MPSkipIntervalCommand_type_alloc() {
	return [MPSkipIntervalCommand
		alloc];
}
void* MPFeedbackCommand_type_alloc() {
	return [MPFeedbackCommand
		alloc];
}
void* MPRatingCommand_type_alloc() {
	return [MPRatingCommand
		alloc];
}
void* MPRemoteCommandCenter_type_alloc() {
	return [MPRemoteCommandCenter
		alloc];
}
void* MPRemoteCommandCenter_type_sharedCommandCenter() {
	return [MPRemoteCommandCenter
		sharedCommandCenter];
}


//void* MPNowPlayingInfoCenter_inst_init(void *id) {
//	return [(MPNowPlayingInfoCenter*)id
//		init];
//}

void* MPNowPlayingInfoCenter_inst_nowPlayingInfo(void *id) {
	return [(MPNowPlayingInfoCenter*)id
		nowPlayingInfo];
}

void MPNowPlayingInfoCenter_inst_setNowPlayingInfo_(void *id, void* value) {
	[(MPNowPlayingInfoCenter*)id
		setNowPlayingInfo: value];
}

unsigned long MPNowPlayingInfoCenter_inst_playbackState(void *id) {
	return [(MPNowPlayingInfoCenter*)id
		playbackState];
}

void MPNowPlayingInfoCenter_inst_setPlaybackState_(void *id, unsigned long value) {
	[(MPNowPlayingInfoCenter*)id
		setPlaybackState: value];
}

void* MPRemoteCommandEvent_inst_init(void *id) {
	return [(MPRemoteCommandEvent*)id
		init];
}

void* MPRemoteCommandEvent_inst_command(void *id) {
	return [(MPRemoteCommandEvent*)id
		command];
}

NSTimeInterval MPRemoteCommandEvent_inst_timestamp(void *id) {
	return [(MPRemoteCommandEvent*)id
		timestamp];
}

void* MPChangePlaybackPositionCommandEvent_inst_init(void *id) {
	return [(MPChangePlaybackPositionCommandEvent*)id
		init];
}

NSTimeInterval MPChangePlaybackPositionCommandEvent_inst_positionTime(void *id) {
	return [(MPChangePlaybackPositionCommandEvent*)id
		positionTime];
}

void* MPChangeLanguageOptionCommandEvent_inst_init(void *id) {
	return [(MPChangeLanguageOptionCommandEvent*)id
		init];
}

void* MPChangePlaybackRateCommandEvent_inst_init(void *id) {
	return [(MPChangePlaybackRateCommandEvent*)id
		init];
}

float MPChangePlaybackRateCommandEvent_inst_playbackRate(void *id) {
	return [(MPChangePlaybackRateCommandEvent*)id
		playbackRate];
}

void* MPChangeRepeatModeCommandEvent_inst_init(void *id) {
	return [(MPChangeRepeatModeCommandEvent*)id
		init];
}

long MPChangeRepeatModeCommandEvent_inst_repeatType(void *id) {
	return [(MPChangeRepeatModeCommandEvent*)id
		repeatType];
}

BOOL MPChangeRepeatModeCommandEvent_inst_preservesRepeatMode(void *id) {
	return [(MPChangeRepeatModeCommandEvent*)id
		preservesRepeatMode];
}

void* MPChangeShuffleModeCommandEvent_inst_init(void *id) {
	return [(MPChangeShuffleModeCommandEvent*)id
		init];
}

long MPChangeShuffleModeCommandEvent_inst_shuffleType(void *id) {
	return [(MPChangeShuffleModeCommandEvent*)id
		shuffleType];
}

BOOL MPChangeShuffleModeCommandEvent_inst_preservesShuffleMode(void *id) {
	return [(MPChangeShuffleModeCommandEvent*)id
		preservesShuffleMode];
}

void* MPSeekCommandEvent_inst_init(void *id) {
	return [(MPSeekCommandEvent*)id
		init];
}

void* MPSkipIntervalCommandEvent_inst_init(void *id) {
	return [(MPSkipIntervalCommandEvent*)id
		init];
}

NSTimeInterval MPSkipIntervalCommandEvent_inst_interval(void *id) {
	return [(MPSkipIntervalCommandEvent*)id
		interval];
}

void MPRemoteCommand_inst_addTarget_action_(void *id, void* target, void* action) {
	[(MPRemoteCommand*)id
		addTarget: target
		action: action];
}

void MPRemoteCommand_inst_removeTarget_(void *id, void* target) {
	[(MPRemoteCommand*)id
		removeTarget: target];
}

void MPRemoteCommand_inst_removeTarget_action_(void *id, void* target, void* action) {
	[(MPRemoteCommand*)id
		removeTarget: target
		action: action];
}

//void* MPRemoteCommand_inst_init(void *id) {
//	return [(MPRemoteCommand*)id
//		init];
//}

BOOL MPRemoteCommand_inst_isEnabled(void *id) {
	return [(MPRemoteCommand*)id
		isEnabled];
}

void MPRemoteCommand_inst_setEnabled_(void *id, BOOL value) {
	[(MPRemoteCommand*)id
		setEnabled: value];
}

//void* MPChangeShuffleModeCommand_inst_init(void *id) {
//	return [(MPChangeShuffleModeCommand*)id
//		init];
//}

long MPChangeShuffleModeCommand_inst_currentShuffleType(void *id) {
	return [(MPChangeShuffleModeCommand*)id
		currentShuffleType];
}

void MPChangeShuffleModeCommand_inst_setCurrentShuffleType_(void *id, long value) {
	[(MPChangeShuffleModeCommand*)id
		setCurrentShuffleType: value];
}

//void* MPChangeRepeatModeCommand_inst_init(void *id) {
//	return [(MPChangeRepeatModeCommand*)id
//		init];
//}

long MPChangeRepeatModeCommand_inst_currentRepeatType(void *id) {
	return [(MPChangeRepeatModeCommand*)id
		currentRepeatType];
}

void MPChangeRepeatModeCommand_inst_setCurrentRepeatType_(void *id, long value) {
	[(MPChangeRepeatModeCommand*)id
		setCurrentRepeatType: value];
}

//void* MPChangePlaybackPositionCommand_inst_init(void *id) {
//	return [(MPChangePlaybackPositionCommand*)id
//		init];
//}

//void* MPChangePlaybackRateCommand_inst_init(void *id) {
//	return [(MPChangePlaybackRateCommand*)id
//		init];
//}

void* MPChangePlaybackRateCommand_inst_supportedPlaybackRates(void *id) {
	return [(MPChangePlaybackRateCommand*)id
		supportedPlaybackRates];
}

void MPChangePlaybackRateCommand_inst_setSupportedPlaybackRates_(void *id, void* value) {
	[(MPChangePlaybackRateCommand*)id
		setSupportedPlaybackRates: value];
}

//void* MPSkipIntervalCommand_inst_init(void *id) {
//	return [(MPSkipIntervalCommand*)id
//		init];
//}

void* MPSkipIntervalCommand_inst_preferredIntervals(void *id) {
	return [(MPSkipIntervalCommand*)id
		preferredIntervals];
}

void MPSkipIntervalCommand_inst_setPreferredIntervals_(void *id, void* value) {
	[(MPSkipIntervalCommand*)id
		setPreferredIntervals: value];
}

//void* MPFeedbackCommand_inst_init(void *id) {
//	return [(MPFeedbackCommand*)id
//		init];
//}

BOOL MPFeedbackCommand_inst_isActive(void *id) {
	return [(MPFeedbackCommand*)id
		isActive];
}

void MPFeedbackCommand_inst_setActive_(void *id, BOOL value) {
	[(MPFeedbackCommand*)id
		setActive: value];
}

void* MPFeedbackCommand_inst_localizedTitle(void *id) {
	return [(MPFeedbackCommand*)id
		localizedTitle];
}

void MPFeedbackCommand_inst_setLocalizedTitle_(void *id, void* value) {
	[(MPFeedbackCommand*)id
		setLocalizedTitle: value];
}

void* MPFeedbackCommand_inst_localizedShortTitle(void *id) {
	return [(MPFeedbackCommand*)id
		localizedShortTitle];
}

void MPFeedbackCommand_inst_setLocalizedShortTitle_(void *id, void* value) {
	[(MPFeedbackCommand*)id
		setLocalizedShortTitle: value];
}

//void* MPRatingCommand_inst_init(void *id) {
//	return [(MPRatingCommand*)id
//		init];
//}

float MPRatingCommand_inst_maximumRating(void *id) {
	return [(MPRatingCommand*)id
		maximumRating];
}

void MPRatingCommand_inst_setMaximumRating_(void *id, float value) {
	[(MPRatingCommand*)id
		setMaximumRating: value];
}

float MPRatingCommand_inst_minimumRating(void *id) {
	return [(MPRatingCommand*)id
		minimumRating];
}

void MPRatingCommand_inst_setMinimumRating_(void *id, float value) {
	[(MPRatingCommand*)id
		setMinimumRating: value];
}

//void* MPRemoteCommandCenter_inst_init(void *id) {
//	return [(MPRemoteCommandCenter*)id
//		init];
//}

void* MPRemoteCommandCenter_inst_pauseCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		pauseCommand];
}

void* MPRemoteCommandCenter_inst_playCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		playCommand];
}

void* MPRemoteCommandCenter_inst_stopCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		stopCommand];
}

void* MPRemoteCommandCenter_inst_togglePlayPauseCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		togglePlayPauseCommand];
}

void* MPRemoteCommandCenter_inst_nextTrackCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		nextTrackCommand];
}

void* MPRemoteCommandCenter_inst_previousTrackCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		previousTrackCommand];
}

void* MPRemoteCommandCenter_inst_changeRepeatModeCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		changeRepeatModeCommand];
}

void* MPRemoteCommandCenter_inst_changeShuffleModeCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		changeShuffleModeCommand];
}

void* MPRemoteCommandCenter_inst_changePlaybackRateCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		changePlaybackRateCommand];
}

void* MPRemoteCommandCenter_inst_seekBackwardCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		seekBackwardCommand];
}

void* MPRemoteCommandCenter_inst_seekForwardCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		seekForwardCommand];
}

void* MPRemoteCommandCenter_inst_skipBackwardCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		skipBackwardCommand];
}

void* MPRemoteCommandCenter_inst_skipForwardCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		skipForwardCommand];
}

void* MPRemoteCommandCenter_inst_changePlaybackPositionCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		changePlaybackPositionCommand];
}

void* MPRemoteCommandCenter_inst_ratingCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		ratingCommand];
}

void* MPRemoteCommandCenter_inst_likeCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		likeCommand];
}

void* MPRemoteCommandCenter_inst_dislikeCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		dislikeCommand];
}

void* MPRemoteCommandCenter_inst_bookmarkCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		bookmarkCommand];
}

void* MPRemoteCommandCenter_inst_enableLanguageOptionCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		enableLanguageOptionCommand];
}

void* MPRemoteCommandCenter_inst_disableLanguageOptionCommand(void *id) {
	return [(MPRemoteCommandCenter*)id
		disableLanguageOptionCommand];
}


BOOL mediaplayer_objc_bool_true = YES;
BOOL mediaplayer_objc_bool_false = NO;

*/
import "C"

func convertObjCBoolToGo(b C.BOOL) bool {
	// NOTE: the prefix here is used to namespace these since the linker will
	// otherwise report a "duplicate symbol" because the C functions have the
	// same name.
	return bool(C.mediaplayer_convertObjCBool(b))
}

func convertToObjCBool(b bool) C.BOOL {
	if b {
		return C.mediaplayer_objc_bool_true
	}
	return C.mediaplayer_objc_bool_false
}

func MPNowPlayingInfoCenter_alloc() (
	r0 MPNowPlayingInfoCenter,
) {
	ret := C.MPNowPlayingInfoCenter_type_alloc()
	r0 = MPNowPlayingInfoCenter_fromPointer(ret)
	return
}

func MPNowPlayingInfoCenter_defaultCenter() (
	r0 MPNowPlayingInfoCenter,
) {
	ret := C.MPNowPlayingInfoCenter_type_defaultCenter()
	r0 = MPNowPlayingInfoCenter_fromPointer(ret)
	return
}

func MPRemoteCommandEvent_alloc() (
	r0 MPRemoteCommandEvent,
) {
	ret := C.MPRemoteCommandEvent_type_alloc()
	r0 = MPRemoteCommandEvent_fromPointer(ret)
	return
}

func MPChangePlaybackPositionCommandEvent_alloc() (
	r0 MPChangePlaybackPositionCommandEvent,
) {
	ret := C.MPChangePlaybackPositionCommandEvent_type_alloc()
	r0 = MPChangePlaybackPositionCommandEvent_fromPointer(ret)
	return
}

func MPChangeLanguageOptionCommandEvent_alloc() (
	r0 MPChangeLanguageOptionCommandEvent,
) {
	ret := C.MPChangeLanguageOptionCommandEvent_type_alloc()
	r0 = MPChangeLanguageOptionCommandEvent_fromPointer(ret)
	return
}

func MPChangePlaybackRateCommandEvent_alloc() (
	r0 MPChangePlaybackRateCommandEvent,
) {
	ret := C.MPChangePlaybackRateCommandEvent_type_alloc()
	r0 = MPChangePlaybackRateCommandEvent_fromPointer(ret)
	return
}

func MPChangeRepeatModeCommandEvent_alloc() (
	r0 MPChangeRepeatModeCommandEvent,
) {
	ret := C.MPChangeRepeatModeCommandEvent_type_alloc()
	r0 = MPChangeRepeatModeCommandEvent_fromPointer(ret)
	return
}

func MPChangeShuffleModeCommandEvent_alloc() (
	r0 MPChangeShuffleModeCommandEvent,
) {
	ret := C.MPChangeShuffleModeCommandEvent_type_alloc()
	r0 = MPChangeShuffleModeCommandEvent_fromPointer(ret)
	return
}

func MPSeekCommandEvent_alloc() (
	r0 MPSeekCommandEvent,
) {
	ret := C.MPSeekCommandEvent_type_alloc()
	r0 = MPSeekCommandEvent_fromPointer(ret)
	return
}

func MPSkipIntervalCommandEvent_alloc() (
	r0 MPSkipIntervalCommandEvent,
) {
	ret := C.MPSkipIntervalCommandEvent_type_alloc()
	r0 = MPSkipIntervalCommandEvent_fromPointer(ret)
	return
}

func MPRemoteCommand_alloc() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommand_type_alloc()
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func MPChangeShuffleModeCommand_alloc() (
	r0 MPChangeShuffleModeCommand,
) {
	ret := C.MPChangeShuffleModeCommand_type_alloc()
	r0 = MPChangeShuffleModeCommand_fromPointer(ret)
	return
}

func MPChangeRepeatModeCommand_alloc() (
	r0 MPChangeRepeatModeCommand,
) {
	ret := C.MPChangeRepeatModeCommand_type_alloc()
	r0 = MPChangeRepeatModeCommand_fromPointer(ret)
	return
}

func MPChangePlaybackPositionCommand_alloc() (
	r0 MPChangePlaybackPositionCommand,
) {
	ret := C.MPChangePlaybackPositionCommand_type_alloc()
	r0 = MPChangePlaybackPositionCommand_fromPointer(ret)
	return
}

func MPChangePlaybackRateCommand_alloc() (
	r0 MPChangePlaybackRateCommand,
) {
	ret := C.MPChangePlaybackRateCommand_type_alloc()
	r0 = MPChangePlaybackRateCommand_fromPointer(ret)
	return
}

func MPSkipIntervalCommand_alloc() (
	r0 MPSkipIntervalCommand,
) {
	ret := C.MPSkipIntervalCommand_type_alloc()
	r0 = MPSkipIntervalCommand_fromPointer(ret)
	return
}

func MPFeedbackCommand_alloc() (
	r0 MPFeedbackCommand,
) {
	ret := C.MPFeedbackCommand_type_alloc()
	r0 = MPFeedbackCommand_fromPointer(ret)
	return
}

func MPRatingCommand_alloc() (
	r0 MPRatingCommand,
) {
	ret := C.MPRatingCommand_type_alloc()
	r0 = MPRatingCommand_fromPointer(ret)
	return
}

func MPRemoteCommandCenter_alloc() (
	r0 MPRemoteCommandCenter,
) {
	ret := C.MPRemoteCommandCenter_type_alloc()
	r0 = MPRemoteCommandCenter_fromPointer(ret)
	return
}

func MPRemoteCommandCenter_sharedCommandCenter() (
	r0 MPRemoteCommandCenter,
) {
	ret := C.MPRemoteCommandCenter_type_sharedCommandCenter()
	r0 = MPRemoteCommandCenter_fromPointer(ret)
	return
}

type MPNowPlayingInfoCenterRef interface {
	Pointer() uintptr
	Init_asMPNowPlayingInfoCenter() MPNowPlayingInfoCenter
}

type gen_MPNowPlayingInfoCenter struct {
	objc.Object
}

func MPNowPlayingInfoCenter_fromPointer(ptr unsafe.Pointer) MPNowPlayingInfoCenter {
	return MPNowPlayingInfoCenter{gen_MPNowPlayingInfoCenter{
		objc.Object_fromPointer(ptr),
	}}
}

func MPNowPlayingInfoCenter_fromRef(ref objc.Ref) MPNowPlayingInfoCenter {
	return MPNowPlayingInfoCenter_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPNowPlayingInfoCenter) Init_asMPNowPlayingInfoCenter() (
	r0 MPNowPlayingInfoCenter,
) {
	//ret := C.MPNowPlayingInfoCenter_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPNowPlayingInfoCenter_fromPointer(ret)
	return
}

func (x gen_MPNowPlayingInfoCenter) NowPlayingInfo() (
	r0 core.NSDictionary,
) {
	ret := C.MPNowPlayingInfoCenter_inst_nowPlayingInfo(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSDictionary_fromPointer(ret)
	return
}

func (x gen_MPNowPlayingInfoCenter) SetNowPlayingInfo_(
	value core.NSDictionaryRef,
) {
	C.MPNowPlayingInfoCenter_inst_setNowPlayingInfo_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_MPNowPlayingInfoCenter) PlaybackState() (
	r0 core.NSUInteger,
) {
	ret := C.MPNowPlayingInfoCenter_inst_playbackState(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSUInteger(ret)
	return
}

func (x gen_MPNowPlayingInfoCenter) SetPlaybackState_(
	value core.NSUInteger,
) {
	C.MPNowPlayingInfoCenter_inst_setPlaybackState_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(value),
	)
	return
}

type MPRemoteCommandEventRef interface {
	Pointer() uintptr
	Init_asMPRemoteCommandEvent() MPRemoteCommandEvent
}

type gen_MPRemoteCommandEvent struct {
	objc.Object
}

func MPRemoteCommandEvent_fromPointer(ptr unsafe.Pointer) MPRemoteCommandEvent {
	return MPRemoteCommandEvent{gen_MPRemoteCommandEvent{
		objc.Object_fromPointer(ptr),
	}}
}

func MPRemoteCommandEvent_fromRef(ref objc.Ref) MPRemoteCommandEvent {
	return MPRemoteCommandEvent_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPRemoteCommandEvent) Init_asMPRemoteCommandEvent() (
	r0 MPRemoteCommandEvent,
) {
	ret := C.MPRemoteCommandEvent_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommandEvent_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandEvent) Command() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandEvent_inst_command(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandEvent) Timestamp() (
	r0 float64,
) {
	ret := C.MPRemoteCommandEvent_inst_timestamp(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float64(ret)
	return
}

type MPChangePlaybackPositionCommandEventRef interface {
	Pointer() uintptr
	Init_asMPChangePlaybackPositionCommandEvent() MPChangePlaybackPositionCommandEvent
}

type gen_MPChangePlaybackPositionCommandEvent struct {
	MPRemoteCommandEvent
}

func MPChangePlaybackPositionCommandEvent_fromPointer(ptr unsafe.Pointer) MPChangePlaybackPositionCommandEvent {
	return MPChangePlaybackPositionCommandEvent{gen_MPChangePlaybackPositionCommandEvent{
		MPRemoteCommandEvent_fromPointer(ptr),
	}}
}

func MPChangePlaybackPositionCommandEvent_fromRef(ref objc.Ref) MPChangePlaybackPositionCommandEvent {
	return MPChangePlaybackPositionCommandEvent_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPChangePlaybackPositionCommandEvent) Init_asMPChangePlaybackPositionCommandEvent() (
	r0 MPChangePlaybackPositionCommandEvent,
) {
	ret := C.MPChangePlaybackPositionCommandEvent_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPChangePlaybackPositionCommandEvent_fromPointer(ret)
	return
}

func (x gen_MPChangePlaybackPositionCommandEvent) PositionTime() (
	r0 float64,
) {
	ret := C.MPChangePlaybackPositionCommandEvent_inst_positionTime(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float64(ret)
	return
}

type MPChangeLanguageOptionCommandEventRef interface {
	Pointer() uintptr
	Init_asMPChangeLanguageOptionCommandEvent() MPChangeLanguageOptionCommandEvent
}

type gen_MPChangeLanguageOptionCommandEvent struct {
	MPRemoteCommandEvent
}

func MPChangeLanguageOptionCommandEvent_fromPointer(ptr unsafe.Pointer) MPChangeLanguageOptionCommandEvent {
	return MPChangeLanguageOptionCommandEvent{gen_MPChangeLanguageOptionCommandEvent{
		MPRemoteCommandEvent_fromPointer(ptr),
	}}
}

func MPChangeLanguageOptionCommandEvent_fromRef(ref objc.Ref) MPChangeLanguageOptionCommandEvent {
	return MPChangeLanguageOptionCommandEvent_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPChangeLanguageOptionCommandEvent) Init_asMPChangeLanguageOptionCommandEvent() (
	r0 MPChangeLanguageOptionCommandEvent,
) {
	ret := C.MPChangeLanguageOptionCommandEvent_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPChangeLanguageOptionCommandEvent_fromPointer(ret)
	return
}

type MPChangePlaybackRateCommandEventRef interface {
	Pointer() uintptr
	Init_asMPChangePlaybackRateCommandEvent() MPChangePlaybackRateCommandEvent
}

type gen_MPChangePlaybackRateCommandEvent struct {
	MPRemoteCommandEvent
}

func MPChangePlaybackRateCommandEvent_fromPointer(ptr unsafe.Pointer) MPChangePlaybackRateCommandEvent {
	return MPChangePlaybackRateCommandEvent{gen_MPChangePlaybackRateCommandEvent{
		MPRemoteCommandEvent_fromPointer(ptr),
	}}
}

func MPChangePlaybackRateCommandEvent_fromRef(ref objc.Ref) MPChangePlaybackRateCommandEvent {
	return MPChangePlaybackRateCommandEvent_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPChangePlaybackRateCommandEvent) Init_asMPChangePlaybackRateCommandEvent() (
	r0 MPChangePlaybackRateCommandEvent,
) {
	ret := C.MPChangePlaybackRateCommandEvent_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPChangePlaybackRateCommandEvent_fromPointer(ret)
	return
}

func (x gen_MPChangePlaybackRateCommandEvent) PlaybackRate() (
	r0 float32,
) {
	ret := C.MPChangePlaybackRateCommandEvent_inst_playbackRate(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

type MPChangeRepeatModeCommandEventRef interface {
	Pointer() uintptr
	Init_asMPChangeRepeatModeCommandEvent() MPChangeRepeatModeCommandEvent
}

type gen_MPChangeRepeatModeCommandEvent struct {
	MPRemoteCommandEvent
}

func MPChangeRepeatModeCommandEvent_fromPointer(ptr unsafe.Pointer) MPChangeRepeatModeCommandEvent {
	return MPChangeRepeatModeCommandEvent{gen_MPChangeRepeatModeCommandEvent{
		MPRemoteCommandEvent_fromPointer(ptr),
	}}
}

func MPChangeRepeatModeCommandEvent_fromRef(ref objc.Ref) MPChangeRepeatModeCommandEvent {
	return MPChangeRepeatModeCommandEvent_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPChangeRepeatModeCommandEvent) Init_asMPChangeRepeatModeCommandEvent() (
	r0 MPChangeRepeatModeCommandEvent,
) {
	ret := C.MPChangeRepeatModeCommandEvent_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPChangeRepeatModeCommandEvent_fromPointer(ret)
	return
}

func (x gen_MPChangeRepeatModeCommandEvent) RepeatType() (
	r0 core.NSInteger,
) {
	ret := C.MPChangeRepeatModeCommandEvent_inst_repeatType(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSInteger(ret)
	return
}

func (x gen_MPChangeRepeatModeCommandEvent) PreservesRepeatMode() (
	r0 bool,
) {
	ret := C.MPChangeRepeatModeCommandEvent_inst_preservesRepeatMode(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

type MPChangeShuffleModeCommandEventRef interface {
	Pointer() uintptr
	Init_asMPChangeShuffleModeCommandEvent() MPChangeShuffleModeCommandEvent
}

type gen_MPChangeShuffleModeCommandEvent struct {
	MPRemoteCommandEvent
}

func MPChangeShuffleModeCommandEvent_fromPointer(ptr unsafe.Pointer) MPChangeShuffleModeCommandEvent {
	return MPChangeShuffleModeCommandEvent{gen_MPChangeShuffleModeCommandEvent{
		MPRemoteCommandEvent_fromPointer(ptr),
	}}
}

func MPChangeShuffleModeCommandEvent_fromRef(ref objc.Ref) MPChangeShuffleModeCommandEvent {
	return MPChangeShuffleModeCommandEvent_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPChangeShuffleModeCommandEvent) Init_asMPChangeShuffleModeCommandEvent() (
	r0 MPChangeShuffleModeCommandEvent,
) {
	ret := C.MPChangeShuffleModeCommandEvent_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPChangeShuffleModeCommandEvent_fromPointer(ret)
	return
}

func (x gen_MPChangeShuffleModeCommandEvent) ShuffleType() (
	r0 core.NSInteger,
) {
	ret := C.MPChangeShuffleModeCommandEvent_inst_shuffleType(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSInteger(ret)
	return
}

func (x gen_MPChangeShuffleModeCommandEvent) PreservesShuffleMode() (
	r0 bool,
) {
	ret := C.MPChangeShuffleModeCommandEvent_inst_preservesShuffleMode(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

type MPSeekCommandEventRef interface {
	Pointer() uintptr
	Init_asMPSeekCommandEvent() MPSeekCommandEvent
}

type gen_MPSeekCommandEvent struct {
	MPRemoteCommandEvent
}

func MPSeekCommandEvent_fromPointer(ptr unsafe.Pointer) MPSeekCommandEvent {
	return MPSeekCommandEvent{gen_MPSeekCommandEvent{
		MPRemoteCommandEvent_fromPointer(ptr),
	}}
}

func MPSeekCommandEvent_fromRef(ref objc.Ref) MPSeekCommandEvent {
	return MPSeekCommandEvent_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPSeekCommandEvent) Init_asMPSeekCommandEvent() (
	r0 MPSeekCommandEvent,
) {
	ret := C.MPSeekCommandEvent_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPSeekCommandEvent_fromPointer(ret)
	return
}

type MPSkipIntervalCommandEventRef interface {
	Pointer() uintptr
	Init_asMPSkipIntervalCommandEvent() MPSkipIntervalCommandEvent
}

type gen_MPSkipIntervalCommandEvent struct {
	MPRemoteCommandEvent
}

func MPSkipIntervalCommandEvent_fromPointer(ptr unsafe.Pointer) MPSkipIntervalCommandEvent {
	return MPSkipIntervalCommandEvent{gen_MPSkipIntervalCommandEvent{
		MPRemoteCommandEvent_fromPointer(ptr),
	}}
}

func MPSkipIntervalCommandEvent_fromRef(ref objc.Ref) MPSkipIntervalCommandEvent {
	return MPSkipIntervalCommandEvent_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPSkipIntervalCommandEvent) Init_asMPSkipIntervalCommandEvent() (
	r0 MPSkipIntervalCommandEvent,
) {
	ret := C.MPSkipIntervalCommandEvent_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPSkipIntervalCommandEvent_fromPointer(ret)
	return
}

func (x gen_MPSkipIntervalCommandEvent) Interval() (
	r0 float64,
) {
	ret := C.MPSkipIntervalCommandEvent_inst_interval(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float64(ret)
	return
}

type MPRemoteCommandRef interface {
	Pointer() uintptr
	Init_asMPRemoteCommand() MPRemoteCommand
}

type gen_MPRemoteCommand struct {
	objc.Object
}

func MPRemoteCommand_fromPointer(ptr unsafe.Pointer) MPRemoteCommand {
	return MPRemoteCommand{gen_MPRemoteCommand{
		objc.Object_fromPointer(ptr),
	}}
}

func MPRemoteCommand_fromRef(ref objc.Ref) MPRemoteCommand {
	return MPRemoteCommand_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPRemoteCommand) AddTarget_action_(
	target objc.Ref,
	action objc.Selector,
) {
	C.MPRemoteCommand_inst_addTarget_action_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(target),
		action.SelectorAddress(),
	)
	return
}

func (x gen_MPRemoteCommand) RemoveTarget_(
	target objc.Ref,
) {
	C.MPRemoteCommand_inst_removeTarget_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(target),
	)
	return
}

func (x gen_MPRemoteCommand) RemoveTarget_action_(
	target objc.Ref,
	action objc.Selector,
) {
	C.MPRemoteCommand_inst_removeTarget_action_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(target),
		action.SelectorAddress(),
	)
	return
}

func (x gen_MPRemoteCommand) Init_asMPRemoteCommand() (
	r0 MPRemoteCommand,
) {
	//ret := C.MPRemoteCommand_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommand) IsEnabled() (
	r0 bool,
) {
	ret := C.MPRemoteCommand_inst_isEnabled(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_MPRemoteCommand) SetEnabled_(
	value bool,
) {
	C.MPRemoteCommand_inst_setEnabled_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

type MPChangeShuffleModeCommandRef interface {
	Pointer() uintptr
	Init_asMPChangeShuffleModeCommand() MPChangeShuffleModeCommand
}

type gen_MPChangeShuffleModeCommand struct {
	MPRemoteCommand
}

func MPChangeShuffleModeCommand_fromPointer(ptr unsafe.Pointer) MPChangeShuffleModeCommand {
	return MPChangeShuffleModeCommand{gen_MPChangeShuffleModeCommand{
		MPRemoteCommand_fromPointer(ptr),
	}}
}

func MPChangeShuffleModeCommand_fromRef(ref objc.Ref) MPChangeShuffleModeCommand {
	return MPChangeShuffleModeCommand_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPChangeShuffleModeCommand) Init_asMPChangeShuffleModeCommand() (
	r0 MPChangeShuffleModeCommand,
) {
	//ret := C.MPChangeShuffleModeCommand_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPChangeShuffleModeCommand_fromPointer(ret)
	return
}

func (x gen_MPChangeShuffleModeCommand) CurrentShuffleType() (
	r0 core.NSInteger,
) {
	ret := C.MPChangeShuffleModeCommand_inst_currentShuffleType(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSInteger(ret)
	return
}

func (x gen_MPChangeShuffleModeCommand) SetCurrentShuffleType_(
	value core.NSInteger,
) {
	C.MPChangeShuffleModeCommand_inst_setCurrentShuffleType_(
		unsafe.Pointer(x.Pointer()),
		C.long(value),
	)
	return
}

type MPChangeRepeatModeCommandRef interface {
	Pointer() uintptr
	Init_asMPChangeRepeatModeCommand() MPChangeRepeatModeCommand
}

type gen_MPChangeRepeatModeCommand struct {
	MPRemoteCommand
}

func MPChangeRepeatModeCommand_fromPointer(ptr unsafe.Pointer) MPChangeRepeatModeCommand {
	return MPChangeRepeatModeCommand{gen_MPChangeRepeatModeCommand{
		MPRemoteCommand_fromPointer(ptr),
	}}
}

func MPChangeRepeatModeCommand_fromRef(ref objc.Ref) MPChangeRepeatModeCommand {
	return MPChangeRepeatModeCommand_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPChangeRepeatModeCommand) Init_asMPChangeRepeatModeCommand() (
	r0 MPChangeRepeatModeCommand,
) {
	//ret := C.MPChangeRepeatModeCommand_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPChangeRepeatModeCommand_fromPointer(ret)
	return
}

func (x gen_MPChangeRepeatModeCommand) CurrentRepeatType() (
	r0 core.NSInteger,
) {
	ret := C.MPChangeRepeatModeCommand_inst_currentRepeatType(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSInteger(ret)
	return
}

func (x gen_MPChangeRepeatModeCommand) SetCurrentRepeatType_(
	value core.NSInteger,
) {
	C.MPChangeRepeatModeCommand_inst_setCurrentRepeatType_(
		unsafe.Pointer(x.Pointer()),
		C.long(value),
	)
	return
}

type MPChangePlaybackPositionCommandRef interface {
	Pointer() uintptr
	Init_asMPChangePlaybackPositionCommand() MPChangePlaybackPositionCommand
}

type gen_MPChangePlaybackPositionCommand struct {
	MPRemoteCommand
}

func MPChangePlaybackPositionCommand_fromPointer(ptr unsafe.Pointer) MPChangePlaybackPositionCommand {
	return MPChangePlaybackPositionCommand{gen_MPChangePlaybackPositionCommand{
		MPRemoteCommand_fromPointer(ptr),
	}}
}

func MPChangePlaybackPositionCommand_fromRef(ref objc.Ref) MPChangePlaybackPositionCommand {
	return MPChangePlaybackPositionCommand_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPChangePlaybackPositionCommand) Init_asMPChangePlaybackPositionCommand() (
	r0 MPChangePlaybackPositionCommand,
) {
	//ret := C.MPChangePlaybackPositionCommand_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPChangePlaybackPositionCommand_fromPointer(ret)
	return
}

type MPChangePlaybackRateCommandRef interface {
	Pointer() uintptr
	Init_asMPChangePlaybackRateCommand() MPChangePlaybackRateCommand
}

type gen_MPChangePlaybackRateCommand struct {
	MPRemoteCommand
}

func MPChangePlaybackRateCommand_fromPointer(ptr unsafe.Pointer) MPChangePlaybackRateCommand {
	return MPChangePlaybackRateCommand{gen_MPChangePlaybackRateCommand{
		MPRemoteCommand_fromPointer(ptr),
	}}
}

func MPChangePlaybackRateCommand_fromRef(ref objc.Ref) MPChangePlaybackRateCommand {
	return MPChangePlaybackRateCommand_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPChangePlaybackRateCommand) Init_asMPChangePlaybackRateCommand() (
	r0 MPChangePlaybackRateCommand,
) {
	//ret := C.MPChangePlaybackRateCommand_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPChangePlaybackRateCommand_fromPointer(ret)
	return
}

func (x gen_MPChangePlaybackRateCommand) SupportedPlaybackRates() (
	r0 core.NSArray,
) {
	ret := C.MPChangePlaybackRateCommand_inst_supportedPlaybackRates(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_MPChangePlaybackRateCommand) SetSupportedPlaybackRates_(
	value core.NSArrayRef,
) {
	C.MPChangePlaybackRateCommand_inst_setSupportedPlaybackRates_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

type MPSkipIntervalCommandRef interface {
	Pointer() uintptr
	Init_asMPSkipIntervalCommand() MPSkipIntervalCommand
}

type gen_MPSkipIntervalCommand struct {
	MPRemoteCommand
}

func MPSkipIntervalCommand_fromPointer(ptr unsafe.Pointer) MPSkipIntervalCommand {
	return MPSkipIntervalCommand{gen_MPSkipIntervalCommand{
		MPRemoteCommand_fromPointer(ptr),
	}}
}

func MPSkipIntervalCommand_fromRef(ref objc.Ref) MPSkipIntervalCommand {
	return MPSkipIntervalCommand_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPSkipIntervalCommand) Init_asMPSkipIntervalCommand() (
	r0 MPSkipIntervalCommand,
) {
	//ret := C.MPSkipIntervalCommand_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPSkipIntervalCommand_fromPointer(ret)
	return
}

func (x gen_MPSkipIntervalCommand) PreferredIntervals() (
	r0 core.NSArray,
) {
	ret := C.MPSkipIntervalCommand_inst_preferredIntervals(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSArray_fromPointer(ret)
	return
}

func (x gen_MPSkipIntervalCommand) SetPreferredIntervals_(
	value core.NSArrayRef,
) {
	C.MPSkipIntervalCommand_inst_setPreferredIntervals_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

type MPFeedbackCommandRef interface {
	Pointer() uintptr
	Init_asMPFeedbackCommand() MPFeedbackCommand
}

type gen_MPFeedbackCommand struct {
	MPRemoteCommand
}

func MPFeedbackCommand_fromPointer(ptr unsafe.Pointer) MPFeedbackCommand {
	return MPFeedbackCommand{gen_MPFeedbackCommand{
		MPRemoteCommand_fromPointer(ptr),
	}}
}

func MPFeedbackCommand_fromRef(ref objc.Ref) MPFeedbackCommand {
	return MPFeedbackCommand_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPFeedbackCommand) Init_asMPFeedbackCommand() (
	r0 MPFeedbackCommand,
) {
	//ret := C.MPFeedbackCommand_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPFeedbackCommand_fromPointer(ret)
	return
}

func (x gen_MPFeedbackCommand) IsActive() (
	r0 bool,
) {
	ret := C.MPFeedbackCommand_inst_isActive(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_MPFeedbackCommand) SetActive_(
	value bool,
) {
	C.MPFeedbackCommand_inst_setActive_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_MPFeedbackCommand) LocalizedTitle() (
	r0 core.NSString,
) {
	ret := C.MPFeedbackCommand_inst_localizedTitle(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSString_fromPointer(ret)
	return
}

func (x gen_MPFeedbackCommand) SetLocalizedTitle_(
	value core.NSStringRef,
) {
	C.MPFeedbackCommand_inst_setLocalizedTitle_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_MPFeedbackCommand) LocalizedShortTitle() (
	r0 core.NSString,
) {
	ret := C.MPFeedbackCommand_inst_localizedShortTitle(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = core.NSString_fromPointer(ret)
	return
}

func (x gen_MPFeedbackCommand) SetLocalizedShortTitle_(
	value core.NSStringRef,
) {
	C.MPFeedbackCommand_inst_setLocalizedShortTitle_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

type MPRatingCommandRef interface {
	Pointer() uintptr
	Init_asMPRatingCommand() MPRatingCommand
}

type gen_MPRatingCommand struct {
	MPRemoteCommand
}

func MPRatingCommand_fromPointer(ptr unsafe.Pointer) MPRatingCommand {
	return MPRatingCommand{gen_MPRatingCommand{
		MPRemoteCommand_fromPointer(ptr),
	}}
}

func MPRatingCommand_fromRef(ref objc.Ref) MPRatingCommand {
	return MPRatingCommand_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPRatingCommand) Init_asMPRatingCommand() (
	r0 MPRatingCommand,
) {
	//ret := C.MPRatingCommand_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPRatingCommand_fromPointer(ret)
	return
}

func (x gen_MPRatingCommand) MaximumRating() (
	r0 float32,
) {
	ret := C.MPRatingCommand_inst_maximumRating(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_MPRatingCommand) SetMaximumRating_(
	value float32,
) {
	C.MPRatingCommand_inst_setMaximumRating_(
		unsafe.Pointer(x.Pointer()),
		C.float(value),
	)
	return
}

func (x gen_MPRatingCommand) MinimumRating() (
	r0 float32,
) {
	ret := C.MPRatingCommand_inst_minimumRating(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_MPRatingCommand) SetMinimumRating_(
	value float32,
) {
	C.MPRatingCommand_inst_setMinimumRating_(
		unsafe.Pointer(x.Pointer()),
		C.float(value),
	)
	return
}

type MPRemoteCommandCenterRef interface {
	Pointer() uintptr
	Init_asMPRemoteCommandCenter() MPRemoteCommandCenter
}

type gen_MPRemoteCommandCenter struct {
	objc.Object
}

func MPRemoteCommandCenter_fromPointer(ptr unsafe.Pointer) MPRemoteCommandCenter {
	return MPRemoteCommandCenter{gen_MPRemoteCommandCenter{
		objc.Object_fromPointer(ptr),
	}}
}

func MPRemoteCommandCenter_fromRef(ref objc.Ref) MPRemoteCommandCenter {
	return MPRemoteCommandCenter_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_MPRemoteCommandCenter) Init_asMPRemoteCommandCenter() (
	r0 MPRemoteCommandCenter,
) {
	//ret := C.MPRemoteCommandCenter_inst_init(
	//	unsafe.Pointer(x.Pointer()),
	//)
	//r0 = MPRemoteCommandCenter_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) PauseCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_pauseCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) PlayCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_playCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) StopCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_stopCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) TogglePlayPauseCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_togglePlayPauseCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) NextTrackCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_nextTrackCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) PreviousTrackCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_previousTrackCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) ChangeRepeatModeCommand() (
	r0 MPChangeRepeatModeCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_changeRepeatModeCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPChangeRepeatModeCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) ChangeShuffleModeCommand() (
	r0 MPChangeShuffleModeCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_changeShuffleModeCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPChangeShuffleModeCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) ChangePlaybackRateCommand() (
	r0 MPChangePlaybackRateCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_changePlaybackRateCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPChangePlaybackRateCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) SeekBackwardCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_seekBackwardCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) SeekForwardCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_seekForwardCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) SkipBackwardCommand() (
	r0 MPSkipIntervalCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_skipBackwardCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPSkipIntervalCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) SkipForwardCommand() (
	r0 MPSkipIntervalCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_skipForwardCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPSkipIntervalCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) ChangePlaybackPositionCommand() (
	r0 MPChangePlaybackPositionCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_changePlaybackPositionCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPChangePlaybackPositionCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) RatingCommand() (
	r0 MPRatingCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_ratingCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRatingCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) LikeCommand() (
	r0 MPFeedbackCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_likeCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPFeedbackCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) DislikeCommand() (
	r0 MPFeedbackCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_dislikeCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPFeedbackCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) BookmarkCommand() (
	r0 MPFeedbackCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_bookmarkCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPFeedbackCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) EnableLanguageOptionCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_enableLanguageOptionCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}

func (x gen_MPRemoteCommandCenter) DisableLanguageOptionCommand() (
	r0 MPRemoteCommand,
) {
	ret := C.MPRemoteCommandCenter_inst_disableLanguageOptionCommand(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = MPRemoteCommand_fromPointer(ret)
	return
}
