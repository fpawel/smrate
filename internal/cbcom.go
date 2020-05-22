package internal

import (
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/smrate/internal/cfg"
	"github.com/fpawel/smrate/internal/pkg/must"
	"github.com/fpawel/smrate/internal/pkg/winapi"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"sort"
)

type cbComport struct {
	cb                 *walk.ComboBox
	disableTextChanged bool
	fnGet              func(cfg.Config) string
	fnSet              func(*cfg.Config, string)
}

func (x *cbComport) Combobox() ComboBox {

	ports, _ := comport.Ports()
	sort.Strings(ports)
	getCurrentIndex := func() int {
		n := -1
		c := cfg.Get()
		for i, s := range ports {
			if s == x.fnGet(c) {
				n = i
				break
			}
		}
		return n
	}
	handleChanged := func() {
		if x.disableTextChanged {
			return
		}
		x.disableTextChanged = true
		c := cfg.Get()
		x.fnSet(&c, x.cb.Text())
		must.PanicIf(cfg.Set(c))
		x.disableTextChanged = false
	}
	return ComboBox{
		Editable:              false,
		AssignTo:              &x.cb,
		MaxSize:               Size{Width: 100},
		MinSize:               Size{Width: 100},
		Model:                 ports,
		CurrentIndex:          getCurrentIndex(),
		OnCurrentIndexChanged: handleChanged,
		OnTextChanged:         handleChanged,
	}
}

func (x *cbComport) trackRegChange( synchronizeFunc func( func() ) ) {
	_ = winapi.NotifyRegChangeComport(func(ports []string) {
		synchronizeFunc(func() {
			x.disableTextChanged = true
			c := cfg.Get()
			_ = x.cb.SetModel(ports)
			_ = x.cb.SetText(x.fnGet(c))
			x.disableTextChanged = false
		})
	})
}
var (
	comboBoxComport = &cbComport{
		fnGet: func(c cfg.Config) string {
			return c.Comport
		},
		fnSet: func(c *cfg.Config, s string) {
			c.Comport = s
		},
	}
)