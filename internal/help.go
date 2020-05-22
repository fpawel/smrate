package internal

import (
	"bytes"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/smrate/internal/pkg/must"
	"github.com/lxn/walk"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

func panicWithSaveRecoveredErrorToFile() {
	x := recover()
	if x == nil {
		return
	}
	err := fmt.Errorf("panic: %+v", x)
	saveErrorToFile(err)

	dir := filepath.Dir(os.Args[0])
	msg := strings.ReplaceAll(err.Error(), dir+"\\", "")
	walk.MsgBox(nil, "Установка контроля ЧЭ", msg,
		walk.MsgBoxIconError|walk.MsgBoxOK|walk.MsgBoxSystemModal)
	panic(x)
}

func saveErrorToFile(saveErr error) {
	file, err := os.OpenFile(filepath.Join(filepath.Dir(os.Args[0]), "errors.log"), os.O_CREATE|os.O_APPEND, 0666)
	must.FatalIf(err)
	defer func() {
		must.FatalIf(file.Close())
	}()
	_, err = file.WriteString(time.Now().Format("2006.01.02 15:04:05") + " " + strings.TrimSpace(saveErr.Error()) + "\n")
	must.FatalIf(err)

	_, err = file.WriteString(formatMerryStacktrace(saveErr) + "\n")
	must.FatalIf(err)

	_, err = file.Write(append(debug.Stack(), '\n'))
	must.FatalIf(err)

}

func formatMerryStacktrace(e error) string {
	s := merry.Stack(e)
	if len(s) == 0 {
		return ""
	}
	buf := bytes.Buffer{}
	for i, fp := range s {
		fnc := runtime.FuncForPC(fp)
		if fnc != nil {
			f, l := fnc.FileLine(fp)
			name := filepath.Base(fnc.Name())
			ident := " "
			if i > 0 {
				ident = "\t"
			}
			buf.WriteString(fmt.Sprintf("%s%s:%d %s\n", ident, f, l, name))
		}
	}
	return buf.String()
}
