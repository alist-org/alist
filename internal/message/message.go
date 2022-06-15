package message

type Messager interface {
	Send(string, interface{}) error
	Receive(string) (string, error)
}
