package stomp

import "strconv"

type Option Header

func (o Option) Set(name string, value ...string) {
	o[name] = value
}

func WithCredentials(login, passcode string) func(Option) {
	return func(option Option) {
		option.Set(HdrLogin, login)
		option.Set(HdrPasscode, passcode)
	}
}

func WithHeartBeat(tx, rx int) func(Option) {
	return func(option Option) {
		option.Set(HdrHeartBeat, strconv.Itoa(tx)+","+strconv.Itoa(rx))
	}
}

func WithContentType(typ string) func(Option) {
	return func(option Option) {
		option.Set(HdrContentType, typ)
	}
}

func WithReceipt(id string) func(Option) {
	return func(option Option) {
		option.Set(HdrReceipt, id)
	}
}

/*
func WithTransaction(trn *Transaction) func(Option) {
	return func(option Option) {
		option.Set(HdrTransaction, trn.id)
	}
}
 */

func WithAck(ack string) func(Option) {
	return func(option Option) {
		option.Set(HdrAck, ack)
	}
}
