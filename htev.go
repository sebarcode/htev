package htev

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"strings"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/kano/kaos/deployer"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
)

type EventRequest struct {
	Headers codekit.M
	Payload []byte
}

type EventResponse struct {
	Data  interface{}
	Error string
}

type Hub struct {
	prefix             string
	secret             string
	service            *kaos.Service
	signature          string
	btr                byter.Byter
	timeout            time.Duration
	defaultPublishOpts *kaos.PublishOpts
	setupOpts          *SetupOpts

	err error
}

type SetupOpts struct {
	BasePath string `json:"base_path"`
}

func init() {
	deployer.RegisterDeployer(DeployerName, func(obj interface{}) (deployer.Deployer, error) {
		m, ok := obj.(codekit.M)
		if !ok {
			m = codekit.M{}
		}
		reqValidation := m.GetBool("require_validation")
		deployerSecret := m.GetString("secret")

		htevDep := new(HtevDeployer)
		if reqValidation {
			htevDep.validateRequest = func(validateObj interface{}) bool {
				secret, ok := validateObj.(string)
				if !ok {
					return false
				}
				if deployerSecret != "" && secret != deployerSecret {
					return false
				}
				return true
			}
		}
		return htevDep, nil
	})

	kaos.RegisterEventHubSetup(DeployerName, func(cfg kaos.EventServerConfig) (kaos.EventHub, error) {
		setupOpts := new(SetupOpts)
		err := serde.Serde(cfg.Data, setupOpts)
		if err != nil {
			return nil, err
		}

		ev := NewHub(byter.NewByter(cfg.ByterName)).(*Hub)
		if cfg.Timeout != 0 {
			ev.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
		}
		ev.SetSecret(cfg.Secret)

		ev.setupOpts = setupOpts
		ev.defaultPublishOpts = &kaos.PublishOpts{Headers: codekit.M{}, Config: codekit.M{}.Set("prefix", setupOpts.BasePath)}
		return ev, nil
	})
}

func NewHub(btr byter.Byter) kaos.EventHub {
	h := new(Hub)
	h.btr = btr
	h.timeout = 5 * time.Second
	return h
}

func (obj *Hub) EventType() string {
	return DeployerName
}

func (obj *Hub) SetPrefix(p string) kaos.EventHub {
	obj.prefix = p
	return obj
}

func (obj *Hub) Prefix() string {
	return obj.prefix
}

// SetDefaultOpts sets the default publishing options for the Hub.
// If the provided options are nil, it initializes a new PublishOpts instance.
// It returns the updated EventHub instance.
//
// Parameters:
//
//	opts - a pointer to kaos.PublishOpts which contains the publishing options.
//
// Returns:
//
//	kaos.EventHub - the updated EventHub instance with the default options set.
func (obj *Hub) SetDefaultOpts(opts *kaos.PublishOpts) kaos.EventHub {
	if opts == nil {
		opts = new(kaos.PublishOpts)
	}
	obj.defaultPublishOpts = opts
	return obj
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
	if opts == nil {
		opts = &kaos.PublishOpts{Headers: codekit.M{}}
	}

	var callOpts kaos.PublishOpts
	if obj.defaultPublishOpts != nil {
		callOpts = *obj.defaultPublishOpts
	}
	opts = kaos.MergePublishOpts(&callOpts, opts)

	routePath := name
	prefix := opts.Config.GetString("prefix")
	if prefix != "" {
		routePath = path.Join(prefix, name)
		routePath = strings.ReplaceAll(routePath, "http:/", "http://")
		routePath = strings.ReplaceAll(routePath, "https:/", "https://")
	}

	bs, err := obj.btr.Encode(data)
	if err != nil {
		return fmt.Errorf("htev encode: %s", err.Error())
	}

	byteReader := bytes.NewReader(bs)
	req, err := http.NewRequest(http.MethodPost, routePath, byteReader)
	if err != nil {
		return fmt.Errorf("htev prepare request: %s", err.Error())
	}
	for k, v := range opts.Headers {
		str, ok := v.(string)
		if !ok {
			continue
		}
		req.Header.Set(k, str)
	}
	req.Header.Set(fmt.Sprintf("x-%s-secret", DeployerName), obj.Secret())

	cl := new(http.Client)
	if obj.Timeout() != time.Duration(0) {
		cl.Timeout = obj.timeout
	}
	resp, err := cl.Do(req)
	if err != nil {
		return fmt.Errorf("htev invoke request: %s", err.Error())
	}
	bsResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("htev read respond: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
		return fmt.Errorf("htev invalid respond: %s: %s", routePath, string(bsResp))
	}

	err = obj.btr.DecodeTo(bsResp, reply, nil)
	if err != nil {
		return fmt.Errorf("htev respond decode: %s", err.Error())
	}

	return nil
}

func (obj *Hub) Unsubscribe(name string, model *kaos.ServiceModel) {
}

func (obj *Hub) Subscribe(name string, model *kaos.ServiceModel, fn interface{}) error {
	return fmt.Errorf("this eventhub does not support Subcribe")
}

func (obj *Hub) SubscribeEx(name string, model *kaos.ServiceModel, fn interface{}) error {
	return obj.SubscribeExWithType(name, model, fn, nil)
}

func (obj *Hub) SubscribeExWithType(name string, model *kaos.ServiceModel, fn interface{}, reqType reflect.Type) error {
	return fmt.Errorf("this eventhub does not support SubcribeEx or SubscribeExWithType")
}

func (obj *Hub) Close() {
}

func (obj *Hub) Error() error {
	return obj.err
}
