// Copyright 2019 The Oto Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build darwin && ios
// +build darwin,ios

#import <AVFoundation/AVFoundation.h>
#import <AudioToolbox/AudioToolbox.h>
#import <UIKit/UIKit.h>

#include "_cgo_export.h"

@interface OtoNotificationObserver : NSObject {
}

- (void)onAudioSessionInterruption:(NSNotification *)notification;

@end

@implementation OtoNotificationObserver {
}

- (void)onAudioSessionInterruption:(NSNotification *)notification {
  if (![notification.name isEqualToString:AVAudioSessionInterruptionNotification]) {
    return;
  }

  NSObject* value = [notification.userInfo valueForKey:AVAudioSessionInterruptionTypeKey];
  AVAudioSessionInterruptionType interruptionType = [(NSNumber*)value intValue];
  switch (interruptionType) {
  case AVAudioSessionInterruptionTypeBegan: {
    oto_setGlobalPause();
    break;
  }
  case AVAudioSessionInterruptionTypeEnded: {
    // AVAudioSessionInterruptionTypeBegan and Ended might not be paired when
    // Siri is used. Then, incrementing and decrementing a counter with this
    // notification doesn't work.
    oto_setGlobalResume();
    break;
  }
  default:
    NSAssert(NO, @"unexpected AVAudioSessionInterruptionType: %lu",
             (unsigned long)(interruptionType));
    break;
  }
}

@end

// oto_setNotificationHandler sets a handler for interruption events.
// Without the handler, Siri would stop the audio (#80).
void oto_setNotificationHandler() {
  AVAudioSession* session = [AVAudioSession sharedInstance];
  OtoNotificationObserver *observer = [[OtoNotificationObserver alloc] init];
  [[NSNotificationCenter defaultCenter]
      addObserver:observer
         selector:@selector(onAudioSessionInterruption:)
             name:AVAudioSessionInterruptionNotification
           object:session];

  // The notifications UIApplicationDidEnterBackgroundNotification and
  // UIApplicationWillEnterForegroundNotification were not reliable: at least,
  // they were not notified at iPod touch A2178.
  //
  // Instead, check the background state via UIApplication actively.
}

bool oto_isBackground(void) {
  if ([NSThread isMainThread]) {
    return [[UIApplication sharedApplication] applicationState] ==
           UIApplicationStateBackground;
  }

  __block bool background = false;
  dispatch_sync(dispatch_get_main_queue(), ^{
    background = oto_isBackground();
  });
  return background;
}
