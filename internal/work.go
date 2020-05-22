package internal

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/smrate/internal/cfg"
	"github.com/fpawel/smrate/internal/pkg/must"
	"os"
	"time"
)

func (x *app) runWork(workName string, work func(context.Context) error)  {

	x.port.SetConfig(log, comportConfig())
	var ctx context.Context
	ctx, x.interruptWorkFunc = context.WithCancel(x.ctx)
	x.wgWork.Add(1)

	x.setupWidgetsRun(true, workName)

	go func() {
		defer panicWithSaveRecoveredErrorToFile()
		if err := work(ctx); err != nil && !merry.Is(err, context.Canceled){
			log.PrintErr(err, "stack", formatMerryStacktrace(err))
			saveErrorToFile(err)
			x.w.Synchronize(func() {
				x.panelError.SetVisible(true)
				must.PanicIf(  x.labelError.SetText(fmt.Sprintf("%s %s", time.Now().Format("15:04:05"), err)) )
			})
		}
		_ = x.port.Close()
		x.wgWork.Done()
		x.setupWidgetsRunSafeUI(false, workName)
	}()
}

func (x *app) performTasks(ctx context.Context, xs []*task) error {

	x.w.Synchronize(func() {
		for _,t := range x.tasks{
			t.pb.SetValue(0)
		}
	})

	for _,t := range xs {
		if err := x.performTask(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (x *app) performTask(ctx context.Context, t *task) error {
	if err := x.setupValves(t.Valve); err != nil {
		return merry.Prependf(err, "%s", t.Str)
	}

	const tickTime = time.Millisecond * 100
	x.w.Synchronize(func() {
		t.pb.SetRange(0, int(t.Dur / tickTime) )
		t.pb.SetValue(0 )
	})

	ctxDelay,_ := context.WithTimeout(ctx, t.Dur)

	tck := time.NewTicker(tickTime)
	defer tck.Stop()
	for {
		select {
		case <-ctxDelay.Done():
			return ctx.Err()
		case <-tck.C:
			x.w.Synchronize(func() {
				t.pb.SetValue(t.pb.Value() + 1)
			})
		}
	}
}


func (x *app) setupValves(value uint16) error {
	_, err := modbus.Request{
		Addr:     0x10,
		ProtoCmd: 0x10,
		Data:     []byte{0x00, 0x32, 0x00, 0x01, 0x02, byte(value >> 8), byte(value)},
	}.GetResponse(log, context.Background(), x.comm())
	if err != nil {
		err = merry.Prepend(err, cfg.Get().Comport).Prependf("%010b", value)
	}
	return err
}


func comportConfig() comport.Config {
	return comport.Config{
		Name:        cfg.Get().Comport,
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	}
}

func (x *app) comm() comm.T {
	if os.Getenv("SMRATE_NO_HARDWARE") == "true" {
		return comport.NewMock(func (req []byte) []byte {
			if len(req) == 11  {
				b := []byte{req[0], req[1], 0x00, 0x00}
				b[2], b[3] = modbus.CRC16(b[:2])
				return b
			}
			return nil
		})
	}
	return comm.New(x.port, cfg.Get().Comm.Comm())
}

