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
			Valve: 0b010000000000,
			Dur:   5 * sec,
			Str:   "Подогрев пробы (термостатирование)",
		},
		/*{
			Valve: 0b000001111000,
			Dur:   5 * sec,
			Str:   "Продувка азотом барбатера",
		},*/
	}
	x.tasks = append([]*task{
		{
			Valve: 0b010100101000,
			Dur:   4 * sec,
			Str:   "Барботирование и прокачивание через пробоотборную петлю",
		},
		{
			Valve: 0b001010010000,
			Dur:   5 * sec,
			Str:   "Слив остатков нефти",
		},
		{
			Valve: 0b111000100001,
			Dur:   5 * sec,
			Str:   "Очистка испарителя от остаточных паров нефти",
		},
		{
			Valve: 0b000000000111,
			Dur:   5 * sec,
			Str:   "Нефть полностью заполняет испаритель",
		},
		{
			Valve: 0b000010010001,
			Dur:   15 * sec,
			Str:   "Прекращение прокачивания, нефть вытесняется до уровня А-А",
		},
		{
			Valve: 0b000001001000,
			Dur:   5 * sec,
			Str:   "Сброс избыточного давления",
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
