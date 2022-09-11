package strutil

import "strings"

/*************************************************************
 * String split operation
 *************************************************************/

// Cut same of the strings.Cut
func Cut(s, sep string) (before string, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

// MustCut always returns two substring.
func MustCut(s, sep string) (before string, after string) {
	before, after, _ = Cut(s, sep)
	return
}

// SplitValid string to slice. will filter empty string node.
func SplitValid(s, sep string) (ss []string) { return Split(s, sep) }

// Split string to slice. will filter empty string node.
func Split(s, sep string) (ss []string) {
	if s = strings.TrimSpace(s); s == "" {
		return
	}

	for _, val := range strings.Split(s, sep) {
		if val = strings.TrimSpace(val); val != "" {
			ss = append(ss, val)
		}
	}
	return
}

// SplitNValid string to slice. will filter empty string node.
func SplitNValid(s, sep string, n int) (ss []string) { return SplitN(s, sep, n) }

// SplitN string to slice. will filter empty string node.
func SplitN(s, sep string, n int) (ss []string) {
	if s = strings.TrimSpace(s); s == "" {
		return
	}

	rawList := strings.Split(s, sep)
	for i, val := range rawList {
		if val = strings.TrimSpace(val); val != "" {
			if len(ss) == n-1 {
				ss = append(ss, strings.TrimSpace(strings.Join(rawList[i:], sep)))
				break
			}

			ss = append(ss, val)
		}
	}
	return
}

// SplitTrimmed split string to slice.
// will trim space for each node, but not filter empty
func SplitTrimmed(s, sep string) (ss []string) {
	if s = strings.TrimSpace(s); s == "" {
		return
	}

	for _, val := range strings.Split(s, sep) {
		ss = append(ss, strings.TrimSpace(val))
	}
	return
}

// SplitNTrimmed split string to slice.
// will trim space for each node, but not filter empty
func SplitNTrimmed(s, sep string, n int) (ss []string) {
	if s = strings.TrimSpace(s); s == "" {
		return
	}

	for _, val := range strings.SplitN(s, sep, n) {
		ss = append(ss, strings.TrimSpace(val))
	}
	return
}

// Substr for a string.
// if length <= 0, return pos to end.
func Substr(s string, pos, length int) string {
	runes := []rune(s)
	strLn := len(runes)

	// pos is too large
	if pos >= strLn {
		return ""
	}

	stopIdx := pos + length
	if length == 0 || stopIdx > strLn {
		stopIdx = strLn
	} else if length < 0 {
		stopIdx = strLn + length
	}

	return string(runes[pos:stopIdx])
}
