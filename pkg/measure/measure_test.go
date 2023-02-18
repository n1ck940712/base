package measure

import (
	"testing"
	"time"
)

func TestExecution(t *testing.T) {
	exec := NewExecution()

	defer func() {
		println("execution: ", exec.Done())
	}()
	time.Sleep(5000 * time.Millisecond)
}
