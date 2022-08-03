package message

import (
	"time"

	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type Post struct {
	Received chan string      // received messages from web
	ToSend   chan interface{} // messages to send to web
}

type Req struct {
	Message string `json:"message" form:"message"`
}

func (p *Post) GetHandle(c *gin.Context) {
	select {
	case message := <-p.ToSend:
		common.SuccessResp(c, message)
	default:
		common.ErrorStrResp(c, "no message", 404)
	}
}

func (p *Post) SendHandle(c *gin.Context) {
	var req Req
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	select {
	case p.Received <- req.Message:
		common.SuccessResp(c)
	default:
		common.ErrorStrResp(c, "send failed", 500)
	}
}

func (p *Post) Send(data interface{}) error {
	select {
	case p.ToSend <- data:
		return nil
	default:
		return errors.New("send failed")
	}
}

func (p *Post) Receive() (string, error) {
	select {
	case message := <-p.Received:
		return message, nil
	default:
		return "", errors.New("receive failed")
	}
}

func (p *Post) WaitSend(data interface{}, d int) error {
	select {
	case p.ToSend <- data:
		return nil
	case <-time.After(time.Duration(d) * time.Second):
		return errors.New("send timeout")
	}
}

func (p *Post) WaitReceive(d int) (string, error) {
	select {
	case message := <-p.Received:
		return message, nil
	case <-time.After(time.Duration(d) * time.Second):
		return "", errors.New("receive timeout")
	}
}

var PostInstance = &Post{
	Received: make(chan string),
	ToSend:   make(chan interface{}),
}
