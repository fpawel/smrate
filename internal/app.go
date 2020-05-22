package internal

import (
	"context"
	"fmt"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/smrate/internal/cfg"
	"github.com/fpawel/smrate/internal/pkg"
	"github.com/fpawel/smrate/internal/pkg/must"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/powerman/structlog"
	"sync"
	"time"
)

type app struct {
	w *walk.MainWindow
	btnRun,
	btnLoop,
	btnDrain,
	btnCloseAllValves,
	btnStop *walk.PushButton
	labelStatusWork *walk.LineEdit
	labelError *walk.TextEdit
	panelError *walk.Composite

	ctx    context.Context
	cancel context.CancelFunc

	shown             bool
	tasks             []*task
	drainTasks        []*task
	port              *comport.Port
	wgWork            sync.WaitGroup
	interruptWorkFunc func()
}

func (x *app) mainWindow() MainWindow {

	return MainWindow{
		AssignTo:   &x.w,
		Title:      "Пробозаборник",
		Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		Layout:    VBox{},
		MenuItems: []MenuItem{},
		Children: []Widget{
			ScrollView{
				VerticalFixed: true,
				MaxSize:       Size{Height: 80, Width: 0},
				MinSize:       Size{Height: 80, Width: 0},
				Layout: HBox{
					Alignment:   AlignHCenterVCenter,
					MarginsZero: true,
				},
				Children: []Widget{
					ImageView{
						Image: "assets/img/rs232_25.png",
					},
					comboBoxComport.Combobox(),
					PushButton{
						AssignTo:  &x.btnRun,
						Text:      " Пробоотбор",
						MaxSize:   Size{Width: 140},
						Image: "assets/img/work25.png",
						OnClicked: func() {
							x.runWork("Отбор пробы", func(ctx context.Context) error {
								return x.performTasks(ctx, x.tasks)
							})
						},
					},
					PushButton{
						AssignTo:  &x.btnLoop,
						Text:      " Цикл",
						Image: "assets/img/loop25.png",
						MaxSize:   Size{Width: 120},
						OnClicked: func() {
							x.runWork("Цикл отбора пробы", func(ctx context.Context) error {
								for {
									if err := x.performTasks(ctx, x.tasks); err != nil {
										return err
									}
								}
							})
						},
					},
					PushButton{
						AssignTo:  &x.btnDrain,
						Text:      " Слив",
						MaxSize:   Size{Width: 120},
						Image: "assets/img/drain25.png",
						OnClicked: func() {
							x.runWork("Слив", func(ctx context.Context) error {
								return x.performTasks(ctx, x.drainTasks)
							})
						},
					},
					PushButton{
						AssignTo:  &x.btnCloseAllValves,
						Text:      " Закрытие",
						MaxSize:   Size{Width: 140},
						Image: "assets/img/close25.png",
						OnClicked: func () {
							x.runWork("Закрытие клапанов", func(context.Context) error {
								return x.setupValves(0)
							})
						},
					},
					PushButton{
						Visible:  false,
						AssignTo: &x.btnStop,
						Text:     " Прервать",
						MaxSize:  Size{Width: 120},
						OnClicked: func() {
							x.interruptWorkFunc()
						},
						Image:     "assets/img/cancel25.png",
					},
					LineEdit{
						AssignTo:  &x.labelStatusWork,
						TextColor: walk.RGB(0, 0, 128),
						ReadOnly:  true,
						Visible: false,
					},
				},
			},
			Composite{
				Layout:   Grid{},
				Children: x.widgetsTasks(),
			},
			Composite{
				Visible:  false,
				AssignTo: &x.panelError,
				Layout:   HBox{MarginsZero: true},
				MaxSize:  Size{Height: 80},
				MinSize:  Size{Height: 80},
				Children: []Widget{
					ImageView{
						Image: "assets/img/error_80.png",
					},
					TextEdit{
						AssignTo:  &x.labelError,
						TextColor: walk.RGB(255, 0, 0),
						ReadOnly:  true,
					},
				},
			},
			TableView{},
		},
	}
}

func (x *app) setupWidgetsRunSafeUI (run bool, workName string) {
	x.w.Synchronize(func() {
		x.setupWidgetsRun(run, workName)
	})
}

func (x *app) setupWidgetsRun (run bool, workName string) {
	for _,btn := range []*walk.PushButton{x.btnRun, x.btnCloseAllValves, x.btnDrain, x.btnLoop}{
		btn.SetVisible(!run)
	}
	x.btnStop.SetVisible(run)
	_ = x.labelStatusWork.SetText( fmt.Sprintf("%s %s: ", time.Now().Format("15:04:05"), workName) )
	s := "выполнение окончено"
	if run {
		x.panelError.SetVisible(false)
		s = "выполняется"
	}
	_ = x.labelStatusWork.SetText( x.labelStatusWork.Text() + s )
	for _,t := range x.tasks{
		t.edSec.SetEnabled(!run)
		t.edMin.SetEnabled(!run)
		t.edHour.SetEnabled(!run)
		if run{
			t.pb.SetValue(0)
		}
	}
	x.labelStatusWork.SetVisible(true)

}

func (x *app) widgetsTasks() []Widget{
	xs := []Widget{
		Label{
			Text:   "час",
			Column: 0,
			Font:   fontHeader,
		},
		Label{
			Text:   "мин",
			Column: 1,
			Font:   fontHeader,
		},
		Label{
			Text:   "сек",
			Column: 2,
			Font:   fontHeader,
		},
		Label{
			Text:   "Наименование",
			Column: 3,
			Font:   fontHeader,
		},
		Label{
			Text:   "Выполнение",
			Column: 4,
			Font:   fontHeader,
		},
	}

	for row, task := range x.tasks {
		row++
		xs = append(xs,
			x.newNe(0, row),
			x.newNe(1, row),
			x.newNe(2, row),
			Label{
				Text:   task.Str,
				Column: 3,
				Row:    row,
			}, ProgressBar{
				AssignTo: &task.pb,
				Column:   4,
				Row:      row,
			})
	}
	return xs
}

func (x *app) newNe(col, row int) NumberEdit {

	neDur := func(ne *walk.NumberEdit) time.Duration {
		return time.Duration(ne.Value())
	}

	sz := Size{Width: 30}
	t := x.tasks[row-1]

	ne := NumberEdit{
		Column:   col,
		Row:      row,
		MinValue: 0,
		MaxValue: 59,
		Decimals: 0,
		MinSize:  sz,
		MaxSize:  sz,
		OnValueChanged: func() {
			if !x.shown {
				return
			}
			t.Dur = time.Hour*neDur(t.edHour) + time.Minute*neDur(t.edMin) + time.Second*neDur(t.edSec)
			c := cfg.Get()
			c.Dur[t.Str] = t.Dur
			must.PanicIf(cfg.Set(c))
		},
	}

	h, m, s := pkg.EncodeDuration(t.Dur)
	switch col {
	case 0:
		ne.Value = float64(h)
		ne.MaxValue = 23
		ne.AssignTo = &t.edHour
	case 1:
		ne.Value = float64(m)
		ne.AssignTo = &t.edMin
	case 2:
		ne.Value = float64(s)
		ne.AssignTo = &t.edSec
	}
	return ne
}

var (
	log = structlog.New()

	fontHeader = Font{
		Family:    "Segoe UI",
		PointSize: 12,
		Bold:      true,
	}
)
