package timex

import (
	"fmt"
	"strings"
	"time"

	"github.com/gookit/goutil/basefn"
	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/strutil"
)

// NowUnix is short of time.Now().Unix()
func NowUnix() int64 {
	return time.Now().Unix()
}

// Format convert time to string use default layout
func Format(t time.Time) string { return t.Format(DefaultLayout) }

// FormatBy given default layout
func FormatBy(t time.Time, layout string) string { return t.Format(layout) }

// Date format time by given date template. see ToLayout() for template parse.
func Date(t time.Time, template ...string) string { return Datetime(t, template...) }

// Datetime convert time to string use template. see ToLayout() for template parse.
func Datetime(t time.Time, template ...string) string {
	return FormatByTpl(t, basefn.FirstOr(template, DefaultTemplate))
}

// DateFormat format time by given date template. see ToLayout()
func DateFormat(t time.Time, template string) string {
	return FormatByTpl(t, template)
}

// FormatByTpl format time by given date template. see ToLayout()
func FormatByTpl(t time.Time, template string) string {
	return t.Format(ToLayout(template))
}

// FormatUnix time seconds use default layout
func FormatUnix(sec int64, layout ...string) string {
	return time.Unix(sec, 0).Format(basefn.FirstOr(layout, DefaultLayout))
}

// FormatUnixBy format time seconds use given layout
func FormatUnixBy(sec int64, layout string) string {
	return time.Unix(sec, 0).Format(layout)
}

// FormatUnixByTpl format time seconds use given date template.
// see ToLayout()
func FormatUnixByTpl(sec int64, template ...string) string {
	layout := ToLayout(basefn.FirstOr(template, DefaultTemplate))
	return time.Unix(sec, 0).Format(layout)
}

// HowLongAgo format given timestamp to string.
func HowLongAgo(sec int64) string {
	return basefn.HowLongAgo(sec)
}

// ToTime parse a datetime string. alias of strutil.ToTime()
func ToTime(s string, layouts ...string) (time.Time, error) {
	return strutil.ToTime(s, layouts...)
}

// ToDur parse a duration string. alias of ToDuration()
func ToDur(s string) (time.Duration, error) { return ToDuration(s) }

// ToDuration parses a duration string. such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func ToDuration(s string) (time.Duration, error) {
	return comfunc.ToDuration(s)
}

// IsDuration check the string is a valid duration string. alias of basefn.IsDuration()
func IsDuration(s string) bool { return comfunc.IsDuration(s) }

// TryToTime parse a date string or duration string to time.Time.
//
// if s is empty, return zero time.
func TryToTime(s string, bt time.Time) (time.Time, error) {
	if s == "" {
		return ZeroTime, nil
	}
	if s == "now" {
		return time.Now(), nil
	}

	// if s is a duration string, add it to bt(base time)
	if IsDuration(s) {
		dur, err := ToDuration(s)
		if err != nil {
			return ZeroTime, err
		}
		return bt.Add(dur), nil
	}

	// as a date string, parse it to time.Time
	return ToTime(s)
}

// InRange check the dst time is in the range of start and end.
//
// if start is zero, only check dst < end,
// if end is zero, only check dst > start.
func InRange(dst, start, end time.Time) bool {
	if start.IsZero() && end.IsZero() {
		return false
	}

	if start.IsZero() {
		return dst.Before(end)
	}
	if end.IsZero() {
		return dst.After(start)
	}

	return dst.After(start) && dst.Before(end)
}

// ParseRangeOpt is the option for ParseRange
type ParseRangeOpt struct {
	// BaseTime is the base time for relative time string.
	// if is zero, use time.Now() as base time.
	BaseTime time.Time
	// OneAsEnd is the option for one time range.
	//  - False: "-1h" => "-1h,0"; "1h" => "+1h, feature"
	//  - True:  "-1h" => "zero,-1h"; "1h" => "zero,1h"
	OneAsEnd bool
	// AutoSort is the option for sort the time range.
	AutoSort bool
	// SepChar is the separator char for time range string. default is '~'
	SepChar byte
	// BeforeFn hook for before parse time string.
	BeforeFn func(string) string
	// KeywordFn is the function for parse keyword time string.
	KeywordFn func(string) (time.Time, time.Time, error)
}

func ensureOpt(opt *ParseRangeOpt) *ParseRangeOpt {
	if opt == nil {
		opt = &ParseRangeOpt{BaseTime: time.Now(), SepChar: '~'}
	} else {
		if opt.BaseTime.IsZero() {
			opt.BaseTime = time.Now()
		}
		if opt.SepChar == 0 {
			opt.SepChar = '~'
		}
	}

	return opt
}

// ParseRange parse time range expression string to time.Time range.
//   - "0" is alias of "now"
//
// Expression format:
//
//	"-5h~-1h"       	=> 5 hours ago to 1 hour ago
//	"1h~5h"         	=> 1 hour after to 5 hours after
//	"-1h~1h"        	=> 1 hour ago to 1 hour after
//	"-1h"            	=> 1 hour ago to feature. eq "-1h,"
//	"-1h~0"          	=> 1 hour ago to now.
//	"< -1h" OR "~-1h"   => 1 hour ago. eq ",-1h"
//	"> 1h" OR "1h"     	=> 1 hour after to feature
//	// keyword: now, today, yesterday, tomorrow
//	"today"          => today start to today end
//	"yesterday"      => yesterday start to yesterday end
//	"tomorrow"       => tomorrow start to tomorrow end
//
// Usage:
//
//	start, end, err := ParseRange("-1h~1h", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(start, end)
func ParseRange(expr string, opt *ParseRangeOpt) (start, end time.Time, err error) {
	opt = ensureOpt(opt)
	expr = strings.TrimSpace(expr)
	if expr == "" {
		err = fmt.Errorf("invalid time range expr %q", expr)
		return
	}

	// parse time range. eg: "5h~1h"
	if strings.IndexByte(expr, opt.SepChar) > -1 {
		s1, s2 := strutil.TrimCut(expr, string(opt.SepChar))
		if s1 == "" && s2 == "" {
			err = fmt.Errorf("invalid time range expr: %s", expr)
			return
		}

		if s1 != "" {
			start, err = TryToTime(s1, opt.BaseTime)
			if err != nil {
				return
			}
		}

		if s2 != "" {
			end, err = TryToTime(s2, opt.BaseTime)
			// auto sort range time
			if opt.AutoSort && err == nil {
				if !start.IsZero() && start.After(end) {
					start, end = end, start
				}
			}
		}

		return
	}

	// single time. eg: "5h", "1h", "-1h"
	if IsDuration(expr) {
		tt, err1 := TryToTime(expr, opt.BaseTime)
		if err1 != nil {
			err = err1
			return
		}

		if opt.OneAsEnd {
			end = tt
		} else {
			start = tt
		}
		return
	}

	// with compare operator. eg: "<1h", ">1h"
	if expr[0] == '<' || expr[0] == '>' {
		tt, err1 := TryToTime(strings.Trim(expr[1:], " ="), opt.BaseTime)
		if err1 != nil {
			err = err1
			return
		}

		if expr[0] == '<' {
			end = tt
		} else {
			start = tt
		}
		return
	}

	// parse keyword time string
	switch expr {
	case "0":
		if opt.OneAsEnd {
			end = opt.BaseTime
		} else {
			start = opt.BaseTime
		}
	case "now":
		if opt.OneAsEnd {
			end = time.Now()
		} else {
			start = time.Now()
		}
	case "today":
		start = DayStart(opt.BaseTime)
		end = DayEnd(opt.BaseTime)
	case "yesterday":
		yd := opt.BaseTime.AddDate(0, 0, -1)
		start = DayStart(yd)
		end = DayEnd(yd)
	case "tomorrow":
		td := opt.BaseTime.AddDate(0, 0, 1)
		start = DayStart(td)
		end = DayEnd(td)
	default:
		// single datetime. eg: "2019-01-01"
		tt, err1 := TryToTime(expr, opt.BaseTime)
		if err1 != nil {
			if opt.KeywordFn == nil {
				err = fmt.Errorf("invalid keyword time string: %s", expr)
				return
			}

			start, end, err = opt.KeywordFn(expr)
			return
		}

		if opt.OneAsEnd {
			end = tt
		} else {
			start = tt
		}
	}

	return
}
