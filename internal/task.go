package internal

import (
	"github.com/fpawel/smrate/internal/cfg"
	"github.com/lxn/walk"
	"time"
)

type task struct {
	Valve uint16
	Dur   time.Duration
	Str   string

	edHour, edMin, edSec *walk.NumberEdit
	pb                   *walk.ProgressBar
}

func (x *app) initTasks() {
	x.drainTasks = []*task{
		{
			Valve: 0b011010010000,
			Dur:   5 * sec,
			Str:   "Слив остатков нефти",
		},
		{
			Valve: 0b000001111000,
			Dur:   5 * sec,
			Str:   "Продувка азотом барбатера",
		},
	}
	x.tasks = append([]*task{
		{
			Valve: 0b010001001110,
			Dur:   4 * sec,
			Str:   "Стадия заполнения",
		},
		{
			Valve: 0b010010010001,
			Dur:   5 * sec,
			Str:   "Избыток нефти вытесняется в дренажную емкость",
		},
		{
			Valve: 0b010010001001,
			Dur:   5 * sec,
			Str:   "Продувка трубки",
		},
		{
			Valve: 0b010001001000,
			Dur:   5 * sec,
			Str:   "Сброс избыточного давления",
		},
		{
			Valve: 0b010000000000,
			Dur:   15 * sec,
			Str:   "Подогрев нефти",
		},
		{
			Valve: 0b000100101000,
			Dur:   5 * sec,
			Str:   "Барботирование азота",
		},
	}, x.drainTasks...)
	m := cfg.Get().Dur
	for _, t := range x.tasks {
		if d, f := m[t.Str]; f {
			t.Dur = d
		}
	}
}

const sec = time.Second
