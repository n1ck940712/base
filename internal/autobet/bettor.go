package autobet

type Bettor interface {
	Bet(elapsedTime int64) //let bettors bet
}

type BettorConfig interface {
	GetBetOnElapsed() (min int64, max int64)
}

type bettor struct {
}

func NewBettor() Bettor {
	return &bettor{}
}

func (b *bettor) Bet(elapsedTime int64) {

}
