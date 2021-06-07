package types

import "strconv"

type Message struct {
	Identifier string
	Round      int
	Block      Hash
	Peer       string
	Signature  []byte
}

func (m *Message) ToString() string {
	return m.Identifier + strconv.Itoa(m.Round) + string(m.Block[:]) + m.Peer
}

func (m *Message) ToBytes() []byte {
	s := []byte(m.Identifier + strconv.Itoa(m.Round))
	s = append(s, m.Block[:]...)
	peerByte := []byte(m.Peer)
	s = append(s, peerByte[:]...)
	return s
}
