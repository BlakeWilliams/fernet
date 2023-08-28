package monte

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/blakewilliams/fernet"
)

// // ControllerHandler is the type of function that can be registered as a route handler
type ControllerHandler[AppData, ControllerData any] func(fernet.Response, *fernet.Request[AppData])

// RoutableController is the interface that controllers must implement
// in order to be registered with a router and instantiated per-request.
type RoutableController[AppData any] interface {
	// Init is called before any of the controller's handlers are called
	Init(fernet.Response, *fernet.Request[AppData], fernet.Handler[AppData])
	Routes(fernet.Routable[AppData])
}

type Controller[AppData any] struct {
	routableController RoutableController[AppData]
}

var _ fernet.Registerable[int] = (*Controller[int])(nil)

func NewController[AppData any](c RoutableController[AppData]) *Controller[AppData] {
	return &Controller[AppData]{
		routableController: c,
	}
}

func (c *Controller[AppData]) Register(parentRouter fernet.Routable[AppData]) {
	router := &controllerRouter[AppData]{
		parent:     parentRouter,
		controller: c.routableController,
	}

	c.routableController.Routes(router)
}

type controllerRouter[AppData any] struct {
	parent     fernet.Routable[AppData]
	controller RoutableController[AppData]
	// Hack until I write all the needed functions
	*fernet.RouteGroup[AppData]
}

var _ fernet.Routable[int] = (*controllerRouter[int])(nil)

func (r *controllerRouter[AppData]) Match(method string, path string, fn fernet.Handler[AppData]) {
	// This is the hackiest go I've ever written, or maybe ever. sorry not sorry
	// also there's probably a better way to do this, but this is what I landed
	// on before calling it quits for the day.
	//
	// This could probably loop over all methods on `c.controller` and compare
	// pointers, though, maybe.
	rawName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	nameParts := strings.Split(rawName, ".")
	funcName := strings.TrimSuffix(nameParts[len(nameParts)-1], "-fm")
	controllerType := reflect.TypeOf(r.controller)
	if controllerType.Kind() == reflect.Pointer {
		controllerType = controllerType.Elem()
	}

	r.parent.Match(method, path, func(req fernet.Response, res *fernet.Request[AppData]) {
		instance := reflect.New(controllerType)
		handler := instance.MethodByName(funcName).Interface().(func(req fernet.Response, res *fernet.Request[AppData]))

		controller := instance.Interface().(RoutableController[AppData])
		controller.Init(req, res, handler)
	})
}

func (c *controllerRouter[AppData]) Get(path string, fn fernet.Handler[AppData]) {
	c.Match("GET", path, fn)
}
