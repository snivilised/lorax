package i18n

import (
	"github.com/snivilised/extendio/i18n"
)

// ❌ OutputChTimeout

// OutputChTimeoutTemplData - occurs when the worker pool times out, attempting
// to send a result to the output channel. This can occur if the client has
// set up te worker pool with a non nil output channel but has failed to also
// provide a consumer to read from this output channel, that would ultimately
// end up in a deadlock had it not been for the occurrence of this error.
type OutputChTimeoutTemplData struct {
	loraxTemplData
}

func (td OutputChTimeoutTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "output-channel-timeout.lorax.nav",
		Description: "Timeout occurred trying to send to output channel (consumer not reading results)",
		Other:       "Timeout occurred workers trying to send to channel. Please ensure the consumer is reading the results",
	}
}

// OutputChTimeoutErrorBehaviourQuery used to query if an error is:
// "Timeout occurred trying to send to output channel"
type OutputChTimeoutErrorBehaviourQuery interface {
	OutputChTimeout() bool
}

type OutputChTimeoutError struct {
	i18n.LocalisableError
}

// OutputChTimeout enables the client to check if error is OutputChTimeoutError
// via OutputChTimeoutErrorBehaviourQuery
func (e OutputChTimeoutError) OutputChTimeout() bool {
	return true
}

// NewOutputChTimeoutError creates a OutputChTimeoutError
func NewOutputChTimeoutError() OutputChTimeoutError {
	return OutputChTimeoutError{
		LocalisableError: i18n.LocalisableError{
			Data: OutputChTimeoutTemplData{},
		},
	}
}
