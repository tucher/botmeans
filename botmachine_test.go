package botmeans

import (
	"sync"
	"testing"
	"time"
)

type TestExecuter struct {
	id      int64
	payload int64
	do      func(int64, int64)
}

func (exec *TestExecuter) Execute() {
	exec.do(exec.id, exec.payload)
}

func (exec *TestExecuter) Id() int64 {
	return exec.id
}

func TestBotMachine(t *testing.T) {
	execChan := make(chan Executer)
	stopChan := RunMachine(execChan, 60*time.Second)

	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	ids := []int64{0, 4, 8, 1, 5, 9, 2, 6, 10, 3, 7, 11}

	for b := 0; b < 1000; b++ {
		testMap := make(map[int64][]int64)
		for _, i := range ids {
			wg.Add(1)
			execChan <- &TestExecuter{
				i / 4,
				i,
				func(id int64, p int64) {
					mutex.Lock()
					testMap[id] = append(testMap[id], p)
					mutex.Unlock()
					wg.Done()
				},
			}
		}
		wg.Wait()

		for k, v := range testMap {
			for i := int64(0); i < 4; i++ {

				if v[i] != k*4+i {
					t.Logf("%+v", testMap)
					t.Fail()
				}
			}
		}
	}
	stopChan <- true
}
