package client

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/mahjongRecordSummaryWebtool/message"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type ClientConn struct {
	ctx context.Context
	*utils.WSClient
	msgIndex uint8
	replys   sync.Map // 回复消息 map[uint8]*Reply
	notify   chan proto.Message
}

type Reply struct {
	out  proto.Message
	wait chan struct{}
}

func NewClientConn(ctx context.Context, addr string) (*ClientConn, error) {
	cConn := &ClientConn{
		ctx:      ctx,
		WSClient: utils.NewWSClient(addr),
		notify:   make(chan proto.Message, 32),
	}
	retry := 3
	duration := 100 * time.Millisecond
	var err error

	// add retry
	for i := 0; i < retry; i++ {
		err = cConn.WSClient.Connect()
		if err != nil {
			time.Sleep(duration)
		} else {
			break
		}
	}

	if err != nil {
		log.Println("ws connect fail", err)
		return nil, err
	}
	go cConn.loop()
	return cConn, nil
}

func (c *ClientConn) loop() {
receive:
	for {
		msg, err := c.WSClient.Read()
		if err != nil {
			log.Println("websocket read fail", err)
			break receive
		}

		if len(msg) > 0 {
			switch msg[0] {
			case MsgTypeNotify:
				//c.handleNotify(msg)	//test
				break
			case MsgTypeResponse:
				c.handleResponse(msg)
				break
			default:
				log.Printf("ClientConn.loop unknown msg type: %d \n", msg[0])
			}
		}

		select {
		case <-c.ctx.Done():
			break receive
		default:
		}
	}
}

// disable 1022
func (c *ClientConn) handleNotify(msg []byte) {
	/*
		wrapper := new(message.Wrapper)
		err := proto.Unmarshal(msg[1:], wrapper)
		if err != nil {
			log.Printf("ClientConn.handleNotify unmarshal error: %v \n", err)
			return
		}
		pm := message.GetNotifyType(wrapper.Name)
		if pm == nil {
			log.Printf("ClientConn.handleNotify unknown notify type: %s \n", wrapper.Name)
			return
		}
		err = proto.Unmarshal(wrapper.Data, pm)
		if err != nil {
			log.Printf("ClientConn.handleNotify unmarshal error: %v \n", err)
			return
		}
		c.notify <- pm

	*/
}

func (c *ClientConn) handleResponse(msg []byte) {
	key := (msg[2] << 7) + msg[1]
	v, ok := c.replys.Load(key)
	if !ok {
		log.Printf("ClientConn.handleResponse not found key: %d \n", key)
		return
	}
	reply, ok := v.(*Reply)
	if !ok {
		log.Printf("ClientConn.handleResponse rv not proto.Message: %v \n", reply)
		return
	}
	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[3:], wrapper)
	if err != nil {
		log.Printf("ClientConn.handleResponse unmarshal error: %v \n", err)
		return
	}
	err = proto.Unmarshal(wrapper.Data, reply.out)
	if err != nil {
		log.Printf("ClientConn.handleResponse unmarshal error: %v \n", err)
		return
	}
	close(reply.wait)
}

func (c *ClientConn) Receive() <-chan proto.Message {
	return c.notify
}

func (c *ClientConn) Invoke(ctx context.Context, method string, in interface{}, out interface{}, opts ...grpc.CallOption) error {
	tokens := strings.Split(method, "/")
	api := strings.Join(tokens, ".")
	return c.Send(ctx, api, in.(proto.Message), out.(proto.Message))
}

func (c *ClientConn) Send(ctx context.Context, api string, in proto.Message, out proto.Message) error {
	body, err := proto.Marshal(in.(proto.Message))
	if err != nil {
		return err
	}

	wrapper := &message.Wrapper{
		Name: api,
		Data: body,
	}

	body, err = proto.Marshal(wrapper)
	if err != nil {
		return err
	}

	buff := new(bytes.Buffer)
	c.msgIndex %= 255
	buff.WriteByte(MsgTypeRequest)
	buff.WriteByte(c.msgIndex - (c.msgIndex >> 7 << 7))
	buff.WriteByte(c.msgIndex >> 7)
	buff.Write(body)

	err = c.WSClient.Send(buff.Bytes())
	if err != nil {
		return err
	}

	reply := &Reply{
		out:  out.(proto.Message),
		wait: make(chan struct{}),
	}
	if _, ok := c.replys.LoadOrStore(c.msgIndex, reply); ok {
		return fmt.Errorf("index exists %d", c.msgIndex)
	}
	defer c.replys.Delete(c.msgIndex)

	c.msgIndex++

	select {
	case <-reply.wait:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (c *ClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("implement me")
}
