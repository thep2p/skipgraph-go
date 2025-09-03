package modules

import "context"

type ThrowableContext interface {
	context.Context
	ThrowIrrecoverable(err error)
}
