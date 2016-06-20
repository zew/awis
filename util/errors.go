package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/zew/awis/logx"
)

func CheckErr(err error) {
	defer logx.SL().Incr().Decr()
	if err != nil {
		logx.Printf("%v", err)
		str := strings.Join(logx.StackTrace(2, 3, 2), "\n")
		logx.Printf("\nStacktrace: \n%s", str)
		os.Exit(1)
	}
}

func IndentedDump(v interface{}) string {

	firstColLeftMostPrefix := " "
	byts, err := json.MarshalIndent(v, firstColLeftMostPrefix, "\t")
	if err != nil {
		s := fmt.Sprintf("error indent: %v\n", err)
		return s
	}

	byts = bytes.Replace(byts, []byte(`\u003c`), []byte("<"), -1)
	byts = bytes.Replace(byts, []byte(`\u003e`), []byte(">"), -1)
	byts = bytes.Replace(byts, []byte(`\n`), []byte("\n"), -1)

	return string(byts)
}
