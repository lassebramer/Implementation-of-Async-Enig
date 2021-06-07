package blockchain

import (
	"math/rand"
)

type Lottery interface {
	TryWin()
	VerifyDraw() bool
	RegisterBlockchain(bc *Blockchain)
}

// lottery stub, to fake lotteries
type FakeLottery struct {
	blockchain *Protocol
}

func MakeFakeLottery(p *Protocol) Lottery {
	lottery := FakeLottery{blockchain: p}
	return &lottery
}

func (l *FakeLottery) TryWin() {
	draw := rand.Intn(1000)
	won := draw > (*l.blockchain).GetHardness()
	if won {
		(*l.blockchain).MakeBlock()
	}
}

func (l *FakeLottery) VerifyDraw() bool {
	return true
}

func (l *FakeLottery) RegisterBlockchain(bc *Blockchain) {
	// l.blockchain = bc
}
