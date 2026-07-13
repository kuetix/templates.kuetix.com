package modules

import (
	di "github.com/kuetix/container"
	StdAuthModule "github.com/kuetix/std-auth/modules"
	StdHttpModule "github.com/kuetix/std-http/modules"
)

func init() {
	di.Boot()
}

func Enable() {
	StdAuthModule.Enable()
	StdHttpModule.Enable()
}
