package mutex

import (
	"sync"
	"testing"

	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func TestMutex(t *testing.T) {
	dataMu := sync.Mutex{}
	data := 0
	data1 := NewData(0)
	data2 := NewData(types.String(""))

	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer dataMu.Unlock()
			dataMu.Lock()
			data++
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			data1.Lock()
			defer data1.Unlock()
			data1.Data += 1

		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			data2.Lock()
			defer data2.Unlock()
			data2.Data = (data2.Data.Int() + 1).String()
		}()

	}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer dataMu.Unlock()
			dataMu.Lock()
			data++
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			data1.Lock()
			defer data1.Unlock()
			data1.Data += 1
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			data2.Lock()
			defer data2.Unlock()
			data2.Data = (data2.Data.Int() + 1).String()
		}()
	}
	wg.Wait()
	println("data: ", data, " data1: ", data1.Data, " data2: ", data2.Data)

}
