package message

import (
	"time"

	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type Http struct {
	Received chan string  // received messages from web
	ToSend   chan Message // messages to send to web
}

type Req struct {
	Message string `json:"message" form:"message"`
}

func (p *Http) GetHandle(c *gin.Context) {
	select {
	case message := <-p.ToSend:
		common.SuccessResp(c, message)
	default:
		common.ErrorStrResp(c, "no message", 404)
	}
}

func (p *Http) SendHandle(c *gin.Context) {
	var req Req
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	select {
	case p.Received <- req.Message:
		common.SuccessResp(c)
	default:
		common.ErrorStrResp(c, "nowhere needed", 500)
	}
}

func (p *Http) Send(message Message) error {
	select {
	case p.ToSend <- message:
		return nil
	default:
		return errors.New("send failed")
	}
}

func (p *Http) Receive() (string, error) {
	select {
	case message := <-p.Received:
		return message, nil
	default:
		return "", errors.New("receive failed")
	}
}

func (p *Http) WaitSend(message Message, d int) error {
	select {
	case p.ToSend <- message:
		return nil
	case <-time.After(time.Duration(d) * time.Second):
		return errors.New("send timeout")
	}
}

func (p *Http) WaitReceive(d int) (string, error) {
	select {
	case message := <-p.Received:
		return message, nil
	case <-time.After(time.Duration(d) * time.Second):
		return "", errors.New("receive timeout")
	}
}

var HttpInstance = &Http{
	Received: make(chan string),
	ToSend:   make(chan Message),
}
