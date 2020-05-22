package pkg

import (
	"fmt"
	"time"
)

func EncodeDuration(d time.Duration) (int,int,int) {

	d = d.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second

	return int(h), int(m), int(s)
}

func FormatDuration(d time.Duration) string {

	d = d.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second

	return fmt.Sprintf("%02d:%02d", m, s)
}
