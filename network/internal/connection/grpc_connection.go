package connection

import "github/thep2p/skipgraph-go/network/internal"

// GRPCConnection represents a connection to a remote peer using gRPC.
type GRPCConnection struct {
}

var _ internal.Connection = (*GRPCConnection)(nil)

func (G GRPCConnection) RemoteAddr() string {
	//TODO implement me
	panic("implement me")
}

func (G GRPCConnection) Send(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (G GRPCConnection) Next() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (G GRPCConnection) Close() error {
	//TODO implement me
	panic("implement me")
}
