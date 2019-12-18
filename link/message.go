package link

// all flags
const (
	FlagRetain = 0x01
	FlagAck    = 0x02
)

// Retain checks whether the message is need to retain
func (m *Message) Retain() bool {
	return m.Context.Flags&FlagRetain == FlagRetain
}

// Ack checks whether the message is a ack message
func (m *Message) Ack() bool {
	return m.Context.Flags&FlagAck == FlagAck
}
