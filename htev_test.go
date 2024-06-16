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

func TestHtev(t *testing.T) {
	cv.Convey("prepare", t, func() {
		ev := htev.NewHub(byter.NewByter("")).SetSecret(eventSecretID).SetTimeout(1 * time.Minute)
		cv.So(ev.Error(), cv.ShouldBeNil)
		defer ev.Close()

		sp := kaos.NewService().SetBasePoint("/event/v1")
		sp.Log().LogToStdOut = false
		sp.RegisterModel(new(htevModel), "model").SetDeployer(htev.DeployerName)

		mux := http.NewServeMux()
		e := htev.NewDeployer(nil, eventSecretID).Deploy(sp, mux)
		cv.So(e, cv.ShouldBeNil)

		go func() {
			http.ListenAndServe("localhost:18080", mux)
			//cv.Printf("l&s: %s", err.Error())
		}()
		time.Sleep(1 * time.Millisecond)

		cv.Convey("htev valid", func() {
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

			cv.Convey("htev invalid", func() {
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
		})
	})
}

type htevModel struct {
}

func (o *htevModel) Register(ctx *kaos.Context, payload codekit.M) (string, error) {
	res := payload.GetString("ID")
	return res, nil
}
