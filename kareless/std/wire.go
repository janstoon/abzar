package std

import (
	"encoding/json"

	"github.com/janstoon/toolbox/kareless"
)

func IdentityRouter(medium int) kareless.Router {
	return kareless.RouterFunc(func(addr string) kareless.Route {
		return kareless.Route{
			Medium:  medium,
			Address: addr,
		}
	})
}

var JsonMarshaler = kareless.MarshalerFunc(func(payload any) []byte {
	bb, _ := json.Marshal(payload)

	return bb
})

var JsonUnmarshaler = kareless.UnmarshalerFunc(json.Unmarshal)

func NoopEncapsulator[M any]() kareless.Encapsulator[M] {
	return kareless.EncapsulatorFunc[M](func(route kareless.Route, data []byte) (m M) {
		return m
	})
}

func NoopDecapsulator[M any]() kareless.Decapsulator[M] {
	return kareless.DecapsulatorFunc[M](func(msg M) ([]byte, error) {
		return nil, nil
	})
}

// NewMuldem creates a simple kareless.Muldem[M] with IdentityRouter(0), Json Encoding (JsonMarshaler, JsonUnmarshaler)
// and Noop Encapsulation (NoopEncapsulator[M], NoopDecapsulator[M]).
func NewMuldem[M any]() kareless.Muldem[M] {
	return kareless.Muldem[M]{
		Router: IdentityRouter(0),

		Marshaler:   JsonMarshaler,
		Unmarshaler: JsonUnmarshaler,

		Encapsulator: NoopEncapsulator[M](),
		Decapsulator: NoopDecapsulator[M](),
	}
}
