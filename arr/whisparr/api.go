package whisparr

import (
	"github.com/golusoris/goenvoy/arr/v2"
)

// api holds the shared BaseClient embedded by Client and ClientV3.
type api struct {
	base *arr.BaseClient
}
