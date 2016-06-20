package logx

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const NumLastDirs = 2 // how many directories to display

const ColWidth = 6 // default column width
const minCX = 2    // minimum spaces towards next columns

var l = log.New(os.Stdout, "lx", log.Lshortfile)

func init() {
	// l := log.New(os.Stdout, "lx", log.Lshortfile)
	l.SetFlags(log.Ltime | log.Lshortfile)
	l.SetFlags(log.Lshortfile)
	l.SetFlags(0)
}

// For giving tracing
func Get() *log.Logger {
	return l
}

func Enable() {
	l.SetOutput(os.Stdout)
}
func Disable() {
	l.SetOutput(ioutil.Discard)
}

func Fatalf(format string, v ...interface{}) {
	defer SL().Incr().Decr()
	Printf(format, v...)
	os.Exit(1)
}

func Fatal(v ...interface{}) {
	defer SL().Incr().Decr()
	l.Print(v...)
	os.Exit(1)
}

func Printf(format string, args ...interface{}) {
	payload := fmt.Sprintf(format, args...)
	payload = strings.TrimRight(payload, "\n")
	payload = Columnify(payload, 56, 4)

	line := fmt.Sprintf("%s%s", logPrefix(), payload)
	if sl.appendStacktrace {
		linesUp := strings.Join(StackTrace(3, 3, 1), "\t")
		line = fmt.Sprintf("%s%s", line, linesUp)
	}
	l.Print(line)
}

func logPrefix() string {
	line, file := StackLine(sl.lvl, NumLastDirs)
	prfx := fmt.Sprintf("%s:%d", file, line)
	prfx = Columnify(prfx, 8, ColWidth)
	return prfx
}

func Columnify(arg string, minWidth, colWidth int) string {
	if len(arg) < minWidth {
		padd := minWidth + minCX - len(arg)
		arg = fmt.Sprintf("%s%s", arg, strings.Repeat(" ", padd))
	} else {
		largestFraction := (len(arg)+minCX)/colWidth + 1
		padd := largestFraction*colWidth - len(arg) // columns of colWidth chars
		arg = fmt.Sprintf("%s%s", arg, strings.Repeat(" ", padd))
	}
	return arg
}

func LastXDirs(path string, numTrailingDirs int) string {

	rump := path // init
	dirs := make([]string, 0, numTrailingDirs)

	for i := 0; i < numTrailingDirs; i++ {
		rump = filepath.Dir(rump)
		dir := filepath.Base(rump)
		dirs = append([]string{dir}, dirs...)
	}

	lastDirs := filepath.Join(dirs...)
	lastDirs = filepath.Join(lastDirs, filepath.Base(path))

	return lastDirs

}

func StackLine(levelUp, numTrailingDirs int) (int, string) {
	_, file, line, _ := runtime.Caller(levelUp + 1) // plus one for myself-func
	return line, LastXDirs(file, numTrailingDirs)
}

func StackTrace(lvlInit, lvlsUp, numLastDirs int) []string {
	ret := make([]string, lvlsUp)
	for i := 0; i < lvlsUp; i++ {
		line, file := StackLine(lvlInit+i, numLastDirs)
		if line == 0 && file == "." {
			break
		}
		ret[i] = fmt.Sprintf("%s:%d", file, line)
		ret[i] = Columnify(ret[i], 12, 12)
	}
	return ret
}
