package htev

import (
	"reflect"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/byter"
	"github.com/sebarcode/codekit"
)

type EventRequest struct {
	Headers codekit.M
	Payload []byte
}

// EventResponse knats will use this for default event response for Kaos,
// kaos will throw 2 of them,  data and error
type EventResponse struct {
	Data  interface{}
	Error string
}

type Hub struct {
	prefix    string
	secret    string
	service   *kaos.Service
	signature string
	btr       byter.Byter
	timeout   time.Duration

	err error
}

func (obj *Hub) SetPrefix(p string) kaos.EventHub {
	obj.prefix = p
	return obj
}

func (obj *Hub) Prefix() string {
	return obj.prefix
}

func (obj *Hub) SetSecret(secret string) kaos.EventHub {
	obj.secret = secret
	return obj
}

func (obj *Hub) Secret() string {
	return obj.secret
}

func (obj *Hub) SetService(svc *kaos.Service) {
	obj.service = svc
}

func (obj *Hub) Service() *kaos.Service {
	return obj.service
}

func (obj *Hub) SetSignature(sign string) kaos.EventHub {
	obj.signature = sign
	return obj
}

func (obj *Hub) Signature() string {
	return obj.signature
}

func (o *Hub) Timeout() time.Duration {
	if int(o.timeout) == 0 {
		o.timeout = 5 * time.Second
	}
	return o.timeout
}

func (o *Hub) SetTimeout(d time.Duration) kaos.EventHub {
	o.timeout = d
	return o
}

func (o *Hub) Byter() byter.Byter {
	return o.btr
}

func (o *Hub) SetByter(b byter.Byter) kaos.EventHub {
	return o
}

func (obj *Hub) Publish(name string, data interface{}, reply interface{}, opts *kaos.PublishOpts) error {
	return nil
}

func (obj *Hub) Unsubscribe(name string, model *kaos.ServiceModel) {
}

func (obj *Hub) Subscribe(name string, model *kaos.ServiceModel, fn interface{}) error {
	return nil
}

func (obj *Hub) SubscribeEx(name string, model *kaos.ServiceModel, fn interface{}) error {
	return nil
}

func (obj *Hub) SubscribeExWithType(name string, model *kaos.ServiceModel, fn interface{}, reqType reflect.Type) error {
	return nil
}

func (obj *Hub) Close() {
}

func (obj *Hub) Error() error {
	return obj.err
}
