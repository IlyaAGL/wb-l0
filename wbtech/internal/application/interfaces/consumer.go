package interfaces

type Consumer interface {
	Consume(msgChan chan<- []byte)
}
