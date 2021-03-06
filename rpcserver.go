package yiyirpc

import (
	"bufio"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"time"
)

func TimeoutCoder(f func(...interface{}) error, e interface{}, msg string) error {
	echan := make(chan error, 1)
	go func() { echan <- f(e) }()
	select {
	case e := <-echan:
		return e
	case <-time.After(time.Hour):
		return fmt.Errorf("Timeout %s", msg)
	}
}

type yiyiServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *msgpack.Decoder
	enc    *msgpack.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *yiyiServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return TimeoutCoder(c.dec.Decode, r, "server read request header")
}

func (c *yiyiServerCodec) ReadRequestBody(body interface{}) error {
	return TimeoutCoder(c.dec.Decode, body, "server read request body")
}

func (c *yiyiServerCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	if err = TimeoutCoder(c.enc.Encode, r, "server write response"); err != nil {
		if c.encBuf.Flush() == nil {
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = TimeoutCoder(c.enc.Encode, body, "server write response body"); err != nil {
		if c.encBuf.Flush() == nil {
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *yiyiServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

type RpcServer struct {
}

func NewRpcServer() *RpcServer {
	return &RpcServer{}
}

func (f *RpcServer) Register(wk interface{}) {
	rpc.Register(wk)
}

func (f *RpcServer) ListenRPC(port int) {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Println("Error: listen error:", err)
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println("Error: accept rpc connection", err.Error())
				continue
			}
			go func(conn net.Conn) {
				buf := bufio.NewWriter(conn)
				srv := &yiyiServerCodec{
					rwc:    conn,
					dec:    msgpack.NewDecoder(conn),
					enc:    msgpack.NewEncoder(buf),
					encBuf: buf,
				}
				err = rpc.ServeRequest(srv)
				if err != nil {
					log.Println("Error: server rpc request", err.Error())
				}
				srv.Close()
			}(conn)
		}
	}()
}

func (f *RpcServer) ListenRPCFullUrl(url string) {
	l, err := net.Listen("tcp", url)
	if err != nil {
		log.Println("Error: listen error:", err)
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println("Error: accept rpc connection", err.Error())
				continue
			}
			go func(conn net.Conn) {
				buf := bufio.NewWriter(conn)
				srv := &yiyiServerCodec{
					rwc:    conn,
					dec:    msgpack.NewDecoder(conn),
					enc:    msgpack.NewEncoder(buf),
					encBuf: buf,
				}
				err = rpc.ServeRequest(srv)
				if err != nil {
					log.Println("Error: server rpc request", err.Error())
				}
				srv.Close()
			}(conn)
		}
	}()
}
