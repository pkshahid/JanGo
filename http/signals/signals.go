package signals

import (
	"github.com/pkshahid/JanGo/core/signals"
)

var (
	// RequestStarted is sent when a request begins.
	RequestStarted = signals.NewSignal("request_started")

	// RequestFinished is sent when a request has finished processing.
	RequestFinished = signals.NewSignal("request_finished")

	// GotRequestException is sent whenever an exception is caught while processing a request.
	GotRequestException = signals.NewSignal("got_request_exception")
)

// Connect is a shortcut to attach a receiver to an HTTP signal.
func Connect(sig *signals.Signal, receiver signals.ReceiverFunc, sender any) error {
	return sig.Connect(receiver, sender, false) // Default sync
}
