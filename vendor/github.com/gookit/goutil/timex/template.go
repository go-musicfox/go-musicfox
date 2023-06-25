package timex

import (
	"time"

	"github.com/gookit/goutil/strutil"
)

// some common datetime templates
const (
	DefaultTemplate = "Y-m-d H:i:s"
	TemplateWithMs3 = "Y-m-d H:i:s.v" // end with ".000"
	TemplateWithMs6 = "Y-m-d H:i:s.u" // end with ".000000"
)

// char to Go date layout
// eg: "Y-m-d H:i:s" => "2006-01-02 15:04:05",
//
// # More see time.stdLongMonth
//
// Char sheet from https://www.php.net/manual/en/datetime.format.php
var charMap = map[byte][]byte{
	// Year
	'Y': []byte("2006"), // long year. eg: 1999, 2003
	'y': []byte("06"),   // short year. eg: 99 or 03
	// Month
	'm': {'0', '1'},        // 01 through 12
	'n': {'1'},             // 1 through 12
	'M': []byte("Jan"),     // Jan through Dec
	'F': []byte("January"), // January through December(full month, php)
	// Day
	'j': {'2'},            // day of the month, 1 to 31
	'd': []byte("02"),     // day of the month, 01 to 31
	'D': []byte("Mon"),    // weekday. Mon through Sun(php)
	'w': []byte("Mon"),    // weekday. Mon through Sun
	'W': []byte("Monday"), // long weekday. Sunday through Saturday
	'l': []byte("Monday"), // long weekday. Sunday through Saturday(php)
	'z': []byte("002"),    // day of the year, 0 through 365
	// Hour
	'H': []byte("15"), // 00 through 23
	'h': []byte("03"), // 01 through 12
	'g': {'3'},        // 1 through 12
	'G': []byte("15"), // go not support 0-23, use 00-23 instead
	// Minutes - 'i' is second char of 'minutes'
	'I': []byte("04"), // 00 to 59
	'i': []byte("4"),  // 0 to 59
	// Seconds
	'S': []byte("05"), // 00 to 59
	's': []byte("5"),  // 0 to 59
	// Time
	'a': []byte("pm"),     // am or pm
	'A': []byte("PM"),     // AM or PM
	'v': []byte("000"),    // Milliseconds eg: 654
	'u': []byte("000000"), // Microseconds eg: 654321
	// Timezone
	'e': []byte("MST"),    // Timezone identifier. eg: UTC, GMT, Atlantic/Azores
	'Z': []byte("Z07"),    // Timezone abbreviation, if known; otherwise the GMT offset. Examples: EST, MDT, +05
	'O': []byte("Z0700"),  // Difference to Greenwich time (GMT) without colon between hours and minutes. Example: +0200
	'P': []byte("Z07:00"), // Difference to Greenwich time (GMT) with colon between hours and minutes. Example: +02:00
	// Full Date/Time
	'c': []byte(time.RFC3339),  // ISO 8601 date. eg: 2004-02-12T15:19:21+00:00
	'r': []byte(time.RFC1123Z), // » RFC 2822/» RFC 5322 formatted date. eg: Thu, 21 Dec 2000 16:01:07 +0200
}

// ToLayout convert chars date template to Go date layout.
//
// template chars see timex.charMap
func ToLayout(template string) string {
	if template == "" {
		return DefaultLayout
	}

	// layout eg: "2006-01-02 15:04:05"
	bts := make([]byte, 0, 24)
	for _, c := range strutil.ToBytes(template) {
		if bs, ok := charMap[c]; ok {
			bts = append(bts, bs...)
		} else {
			bts = append(bts, c)
		}
	}

	return strutil.Byte2str(bts)
}
