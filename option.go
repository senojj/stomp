package stomp

import (
	"github.com/dynata/stomp/proto"
	"strconv"
)

type Option proto.Header

func (o Option) Set(name string, value ...string) {
	o[name] = value
}

func WithCredentials(login, passcode string) func(Option) {
	return func(option Option) {
		option.Set(proto.HdrLogin, login)
		option.Set(proto.HdrPasscode, passcode)
	}
}

func WithHeartBeat(tx, rx int) func(Option) {
	return func(option Option) {
		option.Set(proto.HdrHeartBeat, strconv.Itoa(tx)+","+strconv.Itoa(rx))
	}
}

func WithContentType(typ string) func(Option) {
	return func(option Option) {
		option.Set(proto.HdrContentType, typ)
	}
}

func WithReceipt(id string) func(Option) {
	return func(option Option) {
		option.Set(proto.HdrReceipt, id)
	}
}