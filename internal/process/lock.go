package process

type Lock struct {
	isLocked bool
}

func NewLock() *Lock {
	return &Lock{isLocked: false}
}

func (lock *Lock) Lock() bool {
	if lock.isLocked {
		return true
	}
	lock.isLocked = true
	return false
}

func (lock *Lock) Unlock() {
	lock.isLocked = false
}
