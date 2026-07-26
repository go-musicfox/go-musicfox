package model

import (
	"fmt"
	"strings"
	"sync"
)

// MessageID is a stable key for a translatable string.
type MessageID string

const (
	MsgLoading        MessageID = "loading"
	MsgHintNavigate   MessageID = "hint.navigate"
	MsgHintConfirm    MessageID = "hint.confirm"
	MsgHintBack       MessageID = "hint.back"
	MsgHintQuit       MessageID = "hint.quit"
	MsgHintSearch     MessageID = "hint.search"
	MsgNoData         MessageID = "no_data"
	MsgNoColumns      MessageID = "no_columns"
	MsgEmptyDirectory MessageID = "empty_directory"
	MsgReadError      MessageID = "read_error"
	MsgYes            MessageID = "yes"
	MsgNo             MessageID = "no"
	MsgConfirm        MessageID = "confirm"
	MsgCancel         MessageID = "cancel"
	MsgFieldRequired  MessageID = "field_required"
)

// Catalog stores localized message tables and the currently selected locale.
type Catalog struct {
	mu             sync.RWMutex
	messages       map[string]map[MessageID]string
	locale         string
	fallbackLocale string
}

// NewCatalog returns an empty catalog with English as its active and fallback locale.
func NewCatalog() *Catalog {
	return &Catalog{
		messages:       make(map[string]map[MessageID]string),
		locale:         "en",
		fallbackLocale: "en",
	}
}

// Register adds or overrides messages for locale. The supplied map is copied.
func (c *Catalog) Register(locale string, messages map[MessageID]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.messages[locale] == nil {
		c.messages[locale] = make(map[MessageID]string)
	}
	for id, message := range messages {
		c.messages[locale][id] = message
	}
}

// SetLocale sets the active locale used by T and Tf.
func (c *Catalog) SetLocale(locale string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.locale = locale
}

// SetFallbackLocale sets the locale used when an active-locale message is missing.
func (c *Catalog) SetFallbackLocale(locale string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fallbackLocale = locale
}

// Locale returns the active locale.
func (c *Catalog) Locale() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.locale
}

// T returns the message for id. It looks up the exact active locale, then the
// active locale's language-only form (for example, "zh-CN" then "zh"), then
// the fallback locale, and finally returns id itself when no message is found.
func (c *Catalog) T(id MessageID) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if message, ok := c.lookup(c.locale, id); ok {
		return message
	}
	if language := languageOnly(c.locale); language != c.locale {
		if message, ok := c.lookup(language, id); ok {
			return message
		}
	}
	if message, ok := c.lookup(c.fallbackLocale, id); ok {
		return message
	}
	return string(id)
}

// Tf returns T(id) formatted with fmt.Sprintf and args.
func (c *Catalog) Tf(id MessageID, args ...any) string {
	return fmt.Sprintf(c.T(id), args...)
}

func (c *Catalog) lookup(locale string, id MessageID) (string, bool) {
	message, ok := c.messages[locale][id]
	return message, ok
}

func languageOnly(locale string) string {
	if index := strings.IndexAny(locale, "-_"); index >= 0 {
		return locale[:index]
	}
	return locale
}

var defaultCatalog = newDefaultCatalog()

func newDefaultCatalog() *Catalog {
	catalog := NewCatalog()
	catalog.Register("en", map[MessageID]string{
		MsgLoading:        "Loading...",
		MsgHintNavigate:   "Navigate",
		MsgHintConfirm:    "Confirm",
		MsgHintBack:       "Back",
		MsgHintQuit:       "Quit",
		MsgHintSearch:     "Search",
		MsgNoData:         "No data",
		MsgNoColumns:      "No columns",
		MsgEmptyDirectory: "(empty directory)",
		MsgReadError:      "Error: %s",
		MsgYes:            "Yes",
		MsgNo:             "No",
		MsgConfirm:        "Confirm",
		MsgCancel:         "Cancel",
		MsgFieldRequired:  "This field is required",
	})
	return catalog
}

// DefaultCatalog returns the package-level catalog.
func DefaultCatalog() *Catalog {
	return defaultCatalog
}

// T returns a message from the package-level catalog.
func T(id MessageID) string {
	return defaultCatalog.T(id)
}

// Tf returns a formatted message from the package-level catalog.
func Tf(id MessageID, args ...any) string {
	return defaultCatalog.Tf(id, args...)
}

// SetLocale sets the active locale of the package-level catalog.
func SetLocale(locale string) {
	defaultCatalog.SetLocale(locale)
}
