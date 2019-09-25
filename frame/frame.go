package frame

import (
	"github.com/dynata/stomp/frame/command"
	"github.com/dynata/stomp/frame/header"
)

type Frame struct {
	Command command.Command
	Headers header.Map
}
