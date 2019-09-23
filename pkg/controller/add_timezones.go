package controller

import (
	"github.com/giorgosdi/wallclock-operator/pkg/controller/timezones"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, timezones.Add)
}
