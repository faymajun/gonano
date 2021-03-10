package anet

type Protocol interface {
	Encode(api uint8, data interface{}) ([]byte, error)
	Decode(data []byte) (uint8, interface{}, error)
}
