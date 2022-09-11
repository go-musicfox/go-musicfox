// Copyright 2020 The Oto Authors
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

//go:build darwin && !ios && !js
// +build darwin,!ios,!js

#import <AppKit/AppKit.h>

#include "_cgo_export.h"

@interface OtoNotificationObserver : NSObject {
}

@end

@implementation OtoNotificationObserver {
}

- (void)receiveSleepNote:(NSNotification *)note {
  oto_setGlobalPause();
}

- (void)receiveWakeNote:(NSNotification *)note {
  oto_setGlobalResume();
}

@end

// oto_setNotificationHandler sets a handler for sleep/wake notifications.
void oto_setNotificationHandler() {
  OtoNotificationObserver *observer = [[OtoNotificationObserver alloc] init];

  [[[NSWorkspace sharedWorkspace] notificationCenter]
      addObserver:observer
         selector:@selector(receiveSleepNote:)
             name:NSWorkspaceWillSleepNotification
           object:NULL];
  [[[NSWorkspace sharedWorkspace] notificationCenter]
      addObserver:observer
         selector:@selector(receiveWakeNote:)
             name:NSWorkspaceDidWakeNotification
           object:NULL];
}

bool oto_isBackground(void) {
  // TODO: Should this be implemented?
  return false;
}
