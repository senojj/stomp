package stomp

import (
	"github.com/dynata/stomp/frame"
	"strconv"
)

type Option frame.Header

func (o Option) Set(name string, value ...string) {
	o[name] = value
}

func WithCredentials(login, passcode string) func(Option) {
	return func(option Option) {
		option.Set(frame.HdrLogin, login)
		option.Set(frame.HdrPasscode, passcode)
	}
}

func WithHeartBeat(tx, rx int) func(Option) {
	return func(option Option) {
		option.Set(frame.HdrHeartBeat, strconv.Itoa(tx)+","+strconv.Itoa(rx))
	}
}

func WithContentType(typ string) func(Option) {
	return func(option Option) {
		option.Set(frame.HdrContentType, typ)
	}
}

func WithReceipt() func(Option) {
	return func(option Option) {
		option.Set(frame.HdrReceipt, nextId())
	}
}

func WithTransaction(trn *Transaction) func(Option) {
	return func(option Option) {
		option.Set(frame.HdrTransaction, trn.id)
	}
}

func WithAck(ack string) func(Option) {
	return func(option Option) {
		option.Set(frame.HdrAck, ack)
	}
}