package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/powerman/structlog"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	structlog.DefaultLogger.
		SetPrefixKeys(structlog.KeyLevel).
		SetSuffixKeys(structlog.KeySource).
		SetDefaultKeyvals( structlog.KeySource, structlog.Auto ).
		SetKeysFormat(map[string]string{
			modbus.LogKeyData:   " %[1]s=`% [2]X`",
			structlog.KeySource: " %6[2]s",
			structlog.KeyLevel:  "%3[2]s",
		})

	if len(os.Args) < 2 {
		runScript(parseFile(), false)
		return
	}

	switch os.Args[1] {

	case "run":
		if len(os.Args) > 2 && os.Args[2] == "loop"{
			runScript(parseStrings(os.Args[3:]), true)
		} else {
			runScript(parseStrings(os.Args[2:]), false)
		}

	case "loop":
		runScript(parseFile(), true)

	case "ports":
		printComports()

	case "close":
		openComport()
		setupValves(0)

	case "valve":
		if len(os.Args) < 2 {
			log.Fatal("valve: не задано значение")
		}
		x, err := strconv.ParseUint(os.Args[2], 2, 10)
		if err != nil {
			log.Fatalf("valve: %s", err)
		}
		openComport()
		setupValves(uint16(x))

	default:
		fmt.Print(strUsage)
	}
}

func printComports() {
	ports, err := comport.Ports()
	if err != nil {
		log.Fatal(err)
	}
	for _, port := range ports {
		fmt.Print(port, " ")
	}
	fmt.Println("")
}

func openComport() {
	comm.SetEnableLog(os.Getenv("COMPORT_LOG") == "true")
	Comport := comport.NewPort(comport.Config{
		Name:        os.Getenv("COMPORT"),
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	})
	if err := Comport.Open(); err != nil {
		log.Fatal(err)
	}

	Comm = comm.New(Comport, comm.Config{
		TimeoutGetResponse: 300 * time.Millisecond,
		TimeoutEndResponse: 10 * time.Millisecond,
	})
}

func runScript(actions []uint16DurStr, loop bool) {
	openComport()
labelLoop:
	for n, act := range actions {
		if len(act.Str) > 0 {
			act.Str += ": "
		}
		log.Printf("%d из %d: %s%010b %s", n+1, len(actions), act.Str, act.UInt16, act.Dur)
		setupValves(act.UInt16)
		time.Sleep(act.Dur)
	}
	if loop {
		goto labelLoop
	}
}

func setupValves(value uint16) {
	_, err := modbus.Request{
		Addr:     0x10,
		ProtoCmd: 0x10,
		Data:     []byte{0x00, 0x32, 0x00, 0x01, 0x02, byte(value >> 8), byte(value)},
	}.GetResponse(log, context.Background(), Comm)
	if err != nil {
		log.Fatal(err)
	}
}

type uint16DurStr struct {
	UInt16 uint16
	Dur    time.Duration
	Str    string
}

func parseStrings(Strings []string) (xs []uint16DurStr) {

	if len(Strings) == 0 {
		log.Fatal("аргументы не заданы")
	}

	if len(Strings)%2 != 0 {
		log.Fatal("нечётное количество аргументов")
	}
	for i := 0; i < len(Strings); i += 2 {
		n, dur, err := parseFields(Strings[i:])
		if err != nil {
			log.Fatalf("аргумент в позиции %d: %q,%q: %s", i+1, Strings[i], Strings[i+1], err)
		}
		xs = append(xs, uint16DurStr{
			UInt16: n,
			Dur:    dur,
		})
	}
	return xs
}

func parseFile() (xs []uint16DurStr) {
	f, err := os.Open("smrate.txt")
	if err != nil {
		log.Fatal("smrate.txt", err)
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	line := 0
	for scanner.Scan() {
		str := scanner.Text()
		if len(strings.TrimSpace(str)) == 0 {
			line++
			continue
		}
		fields := strings.Fields(str)
		n, dur, err := parseFields(fields)
		if len(fields) < 2 {
			log.Fatalf("сценарии: строка %d: %q: %s", line+1, str, err)
		}
		var s string
		if len(fields) > 2 {
			s = strings.Join(fields[2:], " ")
		}
		xs = append(xs, uint16DurStr{
			UInt16: n,
			Dur:    dur,
			Str:    s,
		})
		line++
	}
	if len(xs) == 0 {
		log.Fatal("сценарии: не задан")
	}
	return
}

func parseFields(fields []string) (uint16, time.Duration, error) {
	if len(fields) < 2 {
		return 0, 0, errors.New("ожидалось два слова - 10-битное число в двоичном формате и длительность паузы", )
	}
	n, err := strconv.ParseInt(fields[0], 2, 11)
	if err != nil {
		return 0, 0, err
	}
	dur, err := time.ParseDuration(fields[1])
	if err != nil {
		return 0, 0, err
	}
	return uint16(n), dur, nil
}

const strUsage = `Аргументы командной строки smrate.exe:
	- (без аргументов) : однократно выполниить сценарий smrate.txt
	- loop : многократно выполниить сценарий smrate.txt
	- close : закрыть все клапаны
	- ports - распечатать список доступных СОМ портов
	- valve [N] : установить состояния клапанов, где N - 10 бит в двоичном формате
	- run [loop - опционально, повторять многократно] [N1] [P1] [N2] [P2]...[Nn] [Pn] :
		для i от 1 до n:
			- установить состояния клапанов Ni, где Ni - 10 бит в двоичном формате
			- выдержать паузу Pi, где Pi - длительность паузы, знаковая последовательность десятичных чисел,
				каждое из которых содержит необязательную дробь и суффикс единицы времени,
				например "300ms", "-1.5h" или "2h45m".
				Допустимые единицы времени: "ns", "us" (или "µs"), "ms", "s", "m", "h".
Переменные окружения, используемые smrate.exe:
	- COMPORT : имя СОМ порта связи
	- COMPORT_LOG=true : показывать в консоли посылки СОМ порта`

var (
	log  = structlog.New()
	Comm comm.T
)
