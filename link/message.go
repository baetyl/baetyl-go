package link

// Retain checks whether the message is need to retain
func (m *Message) Retain() bool {
	return m.Context.Type == MsgRtn
}
