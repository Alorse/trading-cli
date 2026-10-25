package phemex

import (
	"io"
	"os"
)

func printToStderr() io.Writer { return os.Stderr }
