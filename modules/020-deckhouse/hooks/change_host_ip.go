package hooks

import (
	"github.com/deckhouse/deckhouse/go_lib/hooks/change_host_address"
)

var _ = change_host_address.RegisterHook("deckhouse", "d8-system")