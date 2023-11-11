package message

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type Messenger interface {
	Send(Message) error
	Receive() (string, error)
	WaitSend(Message, int) error
	WaitReceive(int) (string, error)
}

func GetMessenger() Messenger {
	return HttpInstance
}
