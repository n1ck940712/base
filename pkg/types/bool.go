package types

func (bol Bool) Ptr() *bool {
	return (*bool)(&bol)
}
