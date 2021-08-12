package install

import (
	"strconv"
	"strings"
	"time"
)

type NetTop struct {
	delta *NetStat
	last *NetStat
	t0 time.Time
	dt time.Duration
	Interface string
}

func NewNetTop() *NetTop {
	nt := &NetTop{
		delta: NewNetStat(),
		last: NewNetStat(),
		t0: time.Now(),
		dt: 1500 * time.Millisecond,
		Interface: "*",
	}
	return nt
}

func (nt *NetTop) Update() (*NetStat, time.Duration) {
	stat1 := nt.getNetInfo()
	nt.dt = time.Since(nt.t0)

	for _, value := range stat1.Dev {
		t0, ok := nt.last.Stat[value]
		if !ok {
			continue
		}

		dev, ok := nt.delta.Stat[value]
		if !ok {
			nt.delta.Stat[value] = new(DevStat)
			dev = nt.delta.Stat[value]
			nt.delta.Dev = append(nt.delta.Dev, value)
		}
		t1 := stat1.Stat[value]
		dev.Rx = t1.Rx - t0.Rx
		dev.Tx = t1.Tx - t0.Tx
	}
	nt.last = &stat1
	nt.t0 = time.Now()

	return nt.delta, nt.dt
}

func (nt *NetTop) getNetInfo() (ret NetStat) {
	lines, _ := ReadLines("/proc/net/dev")

	ret.Dev = make([]string, 0)
	ret.Stat = make(map[string]*DevStat)

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.Fields(strings.TrimSpace(fields[1]))

		if nt.Interface != "*" && nt.Interface != key {
			continue
		}

		c := new(DevStat)
		c.Name = key
		r, err := strconv.ParseInt(value[0], 10, 64)
		if err != nil {
			//logger.Error(key, "Rx", value[0], err)
			break
		}
		c.Rx = uint64(r)

		t, err := strconv.ParseInt(value[8], 10, 64)
		if err != nil {
			//logger.Error(key, "Tx", value[8], err)
			break
		}
		c.Tx = uint64(t)

		ret.Dev = append(ret.Dev, key)
		ret.Stat[key] = c
	}
	return
}

type NetStat struct {
	Dev  []string
	Stat map[string]*DevStat
}

func NewNetStat() *NetStat {
	return &NetStat{
		Dev: make([]string, 0),
		Stat: make(map[string]*DevStat),
	}
}

type DevStat struct {
	Name string
	Rx   uint64
	Tx   uint64
}