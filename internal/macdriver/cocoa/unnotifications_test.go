//go:build darwin

package cocoa

import (
	"testing"
)

func TestUNClassesRegistered(t *testing.T) {
	if class_UNUserNotificationCenter == 0 {
		t.Fatal("UNUserNotificationCenter class not registered")
	}
	if class_UNMutableNotificationContent == 0 {
		t.Fatal("UNMutableNotificationContent class not registered")
	}
	if class_UNNotificationRequest == 0 {
		t.Fatal("UNNotificationRequest class not registered")
	}
	if class_UNTimeIntervalNotificationTrigger == 0 {
		t.Fatal("UNTimeIntervalNotificationTrigger class not registered")
	}
	if class_UNNotificationAttachment == 0 {
		t.Fatal("UNNotificationAttachment class not registered")
	}
}

func TestUNSelectorsRegistered(t *testing.T) {
	if sel_currentNotificationCenter == 0 {
		t.Fatal("sel_currentNotificationCenter not registered")
	}
	if sel_addNotificationRequest == 0 {
		t.Fatal("sel_addNotificationRequest not registered")
	}
	if sel_triggerWithTimeIntervalRepeats == 0 {
		t.Fatal("sel_triggerWithTimeIntervalRepeats not registered")
	}
	if sel_requestWithIdentifierContentTrigger == 0 {
		t.Fatal("sel_requestWithIdentifierContentTrigger not registered")
	}
	if sel_setValueForKey == 0 {
		t.Fatal("sel_setValueForKey not registered")
	}
	if sel_alloc == 0 {
		t.Fatal("sel_alloc not registered")
	}
	if sel_init == 0 {
		t.Fatal("sel_init not registered")
	}
	if sel_attachmentWithIdentifierURLOptions == 0 {
		t.Fatal("sel_attachmentWithIdentifierURLOptions not registered")
	}
}

func TestUNMutableNotificationContentCreation(t *testing.T) {
	content := UNMutableNotificationContent_new()
	if content.ID == 0 {
		t.Fatal("UNMutableNotificationContent.new() returned nil")
	}
	defer content.Release()

	content.SetTitle("Test Title")
	content.SetBody("Test Body")
	content.SetSound("")
}

func TestUNTimeIntervalNotificationTriggerCreation(t *testing.T) {
	trigger := UNTimeIntervalNotificationTrigger_triggerWithTimeInterval(1.0, false)
	if trigger.ID == 0 {
		t.Fatal("UNTimeIntervalNotificationTrigger.triggerWithTimeInterval() returned nil")
	}
	defer trigger.Release()
}
