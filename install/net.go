package install

import (
	"fmt"
	"time"
)

//BuildNet is
func BuildNet() {
	nettop := NewNetTop()
	nettop.Interface = NetInterface
	nettop.Update()

	count := NetCount
	for {
		time.Sleep(time.Duration(NetUpdateTime*1000) * time.Millisecond)

		delta, dt := nettop.Update()
		dtf := dt.Seconds()

		for _, iface := range delta.Dev {
			stat := delta.Stat[iface]
			bytesR := uint64(float64(stat.Rx) / dtf)
			bytesT := uint64(float64(stat.Tx) / dtf)
			fmt.Printf("%v\t%v\t%v\n", iface, bytesR, bytesT)
		}

		count -= 1
		if count == 0 {
			break
		}
	}
}