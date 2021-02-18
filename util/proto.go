package util

import (
	"github.com/tristoney/xl_chord/proto"
	"github.com/tristoney/xl_chord/util/chorderr"
	"time"
)

func GetTimeForNow() int64 {
	return time.Now().UnixNano()
}

func NewErrorResp(err error) *proto.BaseResp {
	code, msg := chorderr.GetFields(err)
	return &proto.BaseResp{
		ErrNo:   code,
		ErrTips: msg,
		Ts:      GetTimeForNow(),
	}
}

func NewBaseResp() *proto.BaseResp {
	return &proto.BaseResp{
		ErrNo:   0,
		ErrTips: "",
		Ts:      GetTimeForNow(),
	}
}

func NewBaseReq() *proto.BaseReq {
	return &proto.BaseReq{Ts: GetTimeForNow()}
}
