package trace

import (
	"fmt"
	"runtime"
	"strings"
)

// Trace 报错时显示堆栈信息
func Trace(errMessage string) string {
	//存pc值
	var pcstack [32]uintptr

	/*
		Callers fills the slice pc with the return program counters of function invocations
		on the calling goroutine's stack. The argument skip is the number of stack frames
		to skip before recording in pc, with 0 identifying the frame for Callers itself and
		1 identifying the caller of Callers.
		It returns the number of entries written to pc.
	*/
	n := runtime.Callers(3, pcstack[:])

	var str strings.Builder
	str.WriteString(errMessage + "\nTraceback")
	for _, pc := range pcstack[:n] {
		function := runtime.FuncForPC(pc)
		file, line := function.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
