package errors

import "fmt"

var ErrTooManyMessageToDelete = fmt.Errorf("too many message in receiptHandlerMap(should be less that 10)")
