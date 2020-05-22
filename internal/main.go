package internal

import (
	"context"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/smrate/internal/pkg/must"
	"github.com/lxn/win"
)

func Main() {
	defer panicWithSaveRecoveredErrorToFile()

	x := new(app)
	x.initTasks()
	x.ctx, x.cancel = context.WithCancel(context.Background())
	x.interruptWorkFunc = func() {}
	x.port = comport.NewPort(comportConfig())

	must.PanicIf(x.mainWindow().Create())

	x.shown = true

	if !win.ShowWindow(x.w.Handle(), win.SW_SHOWMAXIMIZED) {
		panic("can`t show window")
	}
	go comboBoxComport.trackRegChange(x.w.Synchronize)

	x.w.Run()
	x.interruptWorkFunc()
	x.wgWork.Wait()
}

