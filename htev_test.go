package htev_test

import (
	"net/http"
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/byter"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/htev"
	cv "github.com/smartystreets/goconvey/convey"
)

var (
	eventSecretID = "myAccessToken"
)

func TestMain(m *testing.M) {
	ev := htev.NewHub(byter.NewByter("")).SetSecret(eventSecretID).SetTimeout(1 * time.Minute)
	defer ev.Close()

	sp := kaos.NewService().SetBasePoint("/event/v1")
	sp.Log().LogToStdOut = false
	sp.RegisterModel(new(htevModel), "model").SetDeployer(htev.DeployerName)

	mux := http.NewServeMux()
	htev.NewDeployer(nil, eventSecretID).Deploy(sp, mux)

	go func() {
		http.ListenAndServe("localhost:18080", mux)
		//cv.Printf("l&s: %s", err.Error())
	}()
	time.Sleep(1 * time.Millisecond)

	m.Run()
}

func TestHtevValid(t *testing.T) {
	cv.Convey("htev valid", t, func() {
		ev2 := htev.NewCaller("http://localhost:18080", byter.NewByter(""), 15*time.Second).
			SetSecret(eventSecretID).
			SetDefaultOpts(&kaos.PublishOpts{
				Config: codekit.M{
					"Prefix": "/event/v1",
				},
			})
		resp := ""
		data := codekit.M{}.Set("ID", "User01")
		err := ev2.Publish("/model/Register", data, &resp, &kaos.PublishOpts{})

		cv.So(err, cv.ShouldBeNil)
		cv.So(resp, cv.ShouldEqual, data.GetString("ID"))
	})
}

func TestHtevInvalid(t *testing.T) {
	cv.Convey("htev invalid", t, func() {
		ev2 := htev.NewCaller("http://localhost:18080", byter.NewByter(""), 15*time.Second).
			SetDefaultOpts(&kaos.PublishOpts{
				Config: codekit.M{
					"Prefix": "/event/v1",
				},
			})
		resp := ""
		data := codekit.M{}.Set("ID", "User01")
		err := ev2.Publish("/model/Register", data, &resp, &kaos.PublishOpts{})

		cv.So(err, cv.ShouldNotBeNil)
	})
}

type htevModel struct {
}

func (o *htevModel) Register(ctx *kaos.Context, payload codekit.M) (string, error) {
	res := payload.GetString("ID")
	return res, nil
}
