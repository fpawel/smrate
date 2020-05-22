package cfg

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/smrate/internal/pkg/must"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	Comm        `yaml:"comm"`
	Dur MapStrDur `yaml:"dur"`
}

type MapStrDur = map[string]time.Duration

type Comm struct {
	Log                bool          `yaml:"log"`
	Comport            string        `yaml:"comport"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
}

func (x Comm) Comm() comm.Config {
	return comm.Config{
		TimeoutGetResponse: x.TimeoutGetResponse,
		TimeoutEndResponse: x.TimeoutEndResponse,
		MaxAttemptsRead:    x.MaxAttemptsRead,
	}
}

func Get() Config {
	mu.Lock()
	defer mu.Unlock()
	return getGob()
}

func Set(c Config) error {
	b := must.MarshalYaml(c)
	mu.Lock()
	defer mu.Unlock()
	if err := writeFile(b); err != nil {
		return err
	}
	cfg = c
	comm.SetEnableLog(c.Comm.Log)
	if cfg.Dur == nil {
		cfg.Dur = make(MapStrDur)
	}
	return nil
}

func writeFile(b []byte) error {
	return ioutil.WriteFile(filename(), b, 0666)
}

func filename() string {
	return filepath.Join(filepath.Dir(os.Args[0]), "config.yaml")
}

func readFile() (Config, error) {
	var c Config
	data, err := ioutil.ReadFile(filename())
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(data, &c)
	return c, err
}

func init() {
	var err error
	cfg, err = readFile()
	if err == nil {
		return
	}
	fmt.Println(err, "file:", filename())
	cfg = Config{
		Comm: Comm{
			TimeoutGetResponse: 300 * time.Millisecond,
			TimeoutEndResponse: 10 * time.Millisecond,
			MaxAttemptsRead:    3,
		},
		Dur: make(MapStrDur),
	}
	must.PanicIf(writeFile(must.MarshalYaml(cfg)))
}

func getGob() (r Config) {
	must.PanicIf(enc.Encode(cfg))
	must.PanicIf(dec.Decode(&r))
	return
}

var (
	mu  sync.Mutex
	cfg Config

	buff = new(bytes.Buffer)
	enc  = gob.NewEncoder(buff)
	dec  = gob.NewDecoder(buff)
)
