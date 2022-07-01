package message

type Messenger interface {
	Send(interface{}) error
	Receive() (string, error)
	WaitSend(interface{}, int) error
	WaitReceive(int) (string, error)
}

func GetMessenger() Messenger {
	return PostInstance
}
