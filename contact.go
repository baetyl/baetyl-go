package baetyl

import (
	"fmt"
)

// ContactAPI Contact API
type ContactAPI interface {
	// async mode
	Send(req *Message) error
	Receive(talk Talk) error
	// sync mode
	SendSync(req *Message) (*Message, error)
	RespSync(call Callback) error
}

type Contact struct {
	scfg ContactServerConfig
	ccfg ContactClientConfig
	cli  *CClient
	ser  *CServer
}

// NewContact for send/Receive sync/async message
func NewContact(scfg ContactServerConfig, ccfg ContactClientConfig) (*Contact, error) {
	ser, err := NewCServer(scfg, nil, nil)
	if err != nil {
		return nil, err
	}
	cli, err := NewCClient(ccfg)
	if err != nil {
		return nil, err
	}
	return &Contact{
		scfg: scfg,
		ccfg: ccfg,
		ser:  ser,
		cli:  cli,
	}, nil
}

// Receive implement Talk for receive async message
func (c *Contact) Receive(talk Talk) error {
	if c.ser != nil {
		c.ser.talk = talk
	} else {
		return fmt.Errorf("ContactServer not init")
	}
	return nil
}

// Send for send async message
func (c *Contact) Send(req *Message) error {
	stream, err := c.cli.Talk()
	if err != nil {
		return err
	}
	err = stream.Send(req)
	if err != nil {
		return err
	}
	return nil
}

// SendSync send a sync message
func (c *Contact) SendSync(req *Message) (*Message, error) {
	return c.cli.Call(req)
}

// RespSync implement callback for response sync message
func (c *Contact) RespSync(call Callback) error {
	if c.ser != nil {
		c.ser.call = call
	} else {
		return fmt.Errorf("ContactServer not init")
	}
	return nil
}

// Close closes CServer and CClient
func (c *Contact) Close() {
	if c.ser != nil {
		c.ser.Close()
	}
	if c.cli != nil {
		c.cli.Close()
	}
}
