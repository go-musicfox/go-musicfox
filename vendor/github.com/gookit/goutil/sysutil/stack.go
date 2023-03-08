package sysutil

import (
	"runtime"
	"strconv"
)

// CallerInfo struct
type CallerInfo struct {
	PC   uintptr
	Fc   *runtime.Func
	File string
	Line int
}

// String convert
func (ci *CallerInfo) String() string {
	return ci.File + ":" + strconv.Itoa(ci.Line)
}

// CallersInfos returns an array of the CallerInfo.
//
// Usage:
//
//		cs := sysutil.CallersInfos(3, 2)
//	 for _, ci := range cs {
//			fc := runtime.FuncForPC(pc)
//			// maybe need check fc = nil
//			fnName = fc.Name()
//		}
func CallersInfos(skip, num int, filters ...func(file string, fc *runtime.Func) bool) []*CallerInfo {
	filterLn := len(filters)
	callers := make([]*CallerInfo, 0, num)
	for i := skip; i < skip+num; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			// The breaks below failed to terminate the loop, and we ran off the
			// end of the call stack.
			break
		}

		fc := runtime.FuncForPC(pc)
		if fc == nil {
			continue
		}

		if filterLn > 0 && filters[0] != nil {
			// filter - return false for skip
			if !filters[0](file, fc) {
				continue
			}
		}

		// collecting
		callers = append(callers, &CallerInfo{
			PC:   pc,
			Fc:   fc,
			File: file,
			Line: line,
		})
	}

	return callers
}
