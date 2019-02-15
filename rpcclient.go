package yiyirpc

import (
	"bufio"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io"
	"net"
	"net/rpc"
	"time"
)

type yiyiClientCodec struct {
	rwc    io.ReadWriteCloser
	dec    *msgpack.Decoder
	enc    *msgpack.Encoder
	encBuf *bufio.Writer
}

func (c *yiyiClientCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {
	if err = TimeoutCoder(c.enc.Encode, r, "client write request"); err != nil {
		return
	}
	if err = TimeoutCoder(c.enc.Encode, body, "client write request body"); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *yiyiClientCodec) ReadResponseHeader(r *rpc.Response) error {
	return c.dec.Decode(r)
}

func (c *yiyiClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *yiyiClientCodec) Close() error {
	return c.rwc.Close()
}

type RpcClient struct {
	DialTimeout time.Duration
}

func NewRpcClient(DialTimeout int) *RpcClient {
	return &RpcClient{
		DialTimeout: time.Duration(DialTimeout) * time.Millisecond,
	}
}

func (r *RpcClient) Call(srv string, rpcname string, args interface{}, reply interface{}) error {
	conn, err := net.DialTimeout("tcp", srv, r.DialTimeout)
	if err != nil {
		return fmt.Errorf("ConnectError: %s", err.Error())
	}
	encBuf := bufio.NewWriter(conn)
	codec := &yiyiClientCodec{conn, msgpack.NewDecoder(conn), msgpack.NewEncoder(encBuf), encBuf}
	c := rpc.NewClientWithCodec(codec)
	err = c.Call(rpcname, args, reply)
	errc := c.Close()
	if err != nil && errc != nil {
		return fmt.Errorf("%s %s", err, errc)
	}
	if err != nil {
		return err
	} else {
		return errc
	}
}
