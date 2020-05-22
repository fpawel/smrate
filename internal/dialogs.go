package internal

import (
	"github.com/fpawel/smrate/internal/pkg/must"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func errorDialog( owner walk.Form,  err error) {
	var (
		dlg *walk.Dialog
		pb  *walk.PushButton
	)

	Dlg := Dialog{
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		AssignTo: &dlg,
		Title:    "Произошла ошибка",
		Layout:   HBox{},
		MinSize:  Size{Width: 600, Height: 300},
		MaxSize:  Size{Width: 600, Height: 300},

		CancelButton:  &pb,
		DefaultButton: &pb,

		Children: []Widget{

			TextEdit{
				TextColor: walk.RGB(255, 0, 0),
				ReadOnly:  true,
				Text:      err.Error(),
			},
			ScrollView{
				Layout:          VBox{},
				HorizontalFixed: true,
				Children: []Widget{
					PushButton{
						AssignTo: &pb,
						Text:     "Продолжить",
						OnClicked: func() {
							dlg.Accept()
						},
					},
					ImageView{
						Image: "assets/img/error_80.png",
					},
				},
			},
		},
	}
	must.PanicIf(Dlg.Create(owner))
	_ = pb.SetFocus()
	dlg.Run()
}
