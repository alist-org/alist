package message

type Messenger interface {
	Send(string, interface{}) error
	WaitSend(string, interface{}) error
	Receive(string) (string, error)
	WaitReceive(string) (string, error)
}
