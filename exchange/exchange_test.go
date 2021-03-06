// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package exchange

import (
	"bytes"
	"testing"

	"time"

	"github.com/TheThingsNetwork/gateway-connector-bridge/auth"
	"github.com/TheThingsNetwork/gateway-connector-bridge/backend/dummy"
	"github.com/TheThingsNetwork/gateway-connector-bridge/types"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	. "github.com/smartystreets/goconvey/convey"
)

func TestExchange(t *testing.T) {
	Convey("Given a new Context and Backends", t, func(c C) {

		var logs bytes.Buffer
		ctx := &log.Logger{
			Handler: text.New(&logs),
			Level:   log.DebugLevel,
		}
		defer func() {
			if logs.Len() > 0 {
				c.Printf("\n%s", logs.String())
			}
		}()

		ttn := dummy.New(ctx.WithField("Direction", "TTN"))
		gateway := dummy.New(ctx.WithField("Direction", "Gateway"))

		auth := auth.NewMemory()

		Convey("When creating a new Exchange", func() {
			b := New(ctx)
			b.SetAuth(auth)

			Convey("When adding a Northbound and Southbound backend", func() {
				b.AddNorthbound(ttn)
				b.AddSouthbound(gateway)

				Convey("When starting the Exchange", func() {
					b.Start(1, 10*time.Millisecond)

					Convey("When stopping the Exchange", func() {
						b.Stop()
					})

					Convey("When sending a connect message with a Key", func() {
						err := gateway.PublishConnect(&types.ConnectMessage{
							GatewayID: "dev",
							Key:       "key",
						})
						time.Sleep(10 * time.Millisecond)
						Convey("There should be no error", func() {
							So(err, ShouldBeNil)
						})
						Convey("The gateway should be connected", func() {
							So(b.gateways.Contains("dev"), ShouldBeTrue)
						})

						Convey("When sending a disconnect message with the same Key", func() {
							err := gateway.PublishDisconnect(&types.DisconnectMessage{
								GatewayID: "dev",
								Key:       "key",
							})
							time.Sleep(10 * time.Millisecond)
							Convey("There should be no error", func() {
								So(err, ShouldBeNil)
							})
							Convey("The gateway should be disconnected", func() {
								So(b.gateways.Contains("dev"), ShouldBeFalse)
							})
						})

						Convey("When sending a disconnect message with a different Key", func() {
							err := gateway.PublishDisconnect(&types.DisconnectMessage{
								GatewayID: "dev",
								Key:       "other-key",
							})
							time.Sleep(10 * time.Millisecond)
							Convey("There should be no error", func() {
								So(err, ShouldBeNil)
							})
							Convey("The gateway should not be disconnected", func() {
								So(b.gateways.Contains("dev"), ShouldBeTrue)
							})
						})
					})

					Convey("When sending a connect message", func() {
						err := gateway.PublishConnect(&types.ConnectMessage{
							GatewayID: "dev",
						})
						time.Sleep(10 * time.Millisecond)
						Convey("There should be no error", func() {
							So(err, ShouldBeNil)
						})
						Convey("The gateway should be connected", func() {
							So(b.gateways.Contains("dev"), ShouldBeTrue)
						})

						Convey("When sending another connect message", func() {
							err := gateway.PublishConnect(&types.ConnectMessage{
								GatewayID: "dev",
							})
							time.Sleep(10 * time.Millisecond)
							Convey("There should be no error", func() {
								So(err, ShouldBeNil)
							})
							Convey("The gateway should still be connected", func() {
								So(b.gateways.Contains("dev"), ShouldBeTrue)
							})
						})

						Convey("When subscribing to uplink messages on the TTN side", func() {
							msg, _ := ttn.SubscribeUplink("dev")
							time.Sleep(10 * time.Millisecond)

							Convey("When sending an uplink message on the Gateway side", func() {
								err := gateway.PublishUplink(&types.UplinkMessage{
									GatewayID: "dev",
									Message:   &pb_router.UplinkMessage{},
								})
								Convey("There should be no error", func() {
									So(err, ShouldBeNil)
								})

								Convey("Then it should arrive on the TTN side", func() {
									select {
									case <-time.After(time.Second):
										So("Timeout Exceeded", ShouldBeFalse)
									case _, ok := <-msg:
										So(ok, ShouldBeTrue)
									}
								})
							})
						})

						Convey("When subscribing to downlink messages on the Gateway side", func() {
							msg, _ := gateway.SubscribeDownlink("dev")
							time.Sleep(10 * time.Millisecond)

							Convey("When sending a downlink message on the TTN side", func() {
								err := ttn.PublishDownlink(&types.DownlinkMessage{
									GatewayID: "dev",
									Message:   &pb_router.DownlinkMessage{},
								})
								Convey("There should be no error", func() {
									So(err, ShouldBeNil)
								})

								Convey("Then it should arrive on the Gateway side", func() {
									select {
									case <-time.After(time.Second):
										So("Timeout Exceeded", ShouldBeFalse)
									case _, ok := <-msg:
										So(ok, ShouldBeTrue)
									}
								})
							})
						})

						Convey("When subscribing to status messages on the TTN side", func() {
							msg, _ := ttn.SubscribeStatus("dev")
							time.Sleep(10 * time.Millisecond)

							Convey("When sending an status message on the Gateway side", func() {
								err := gateway.PublishStatus(&types.StatusMessage{
									GatewayID: "dev",
									Message:   &pb_gateway.Status{},
								})
								Convey("There should be no error", func() {
									So(err, ShouldBeNil)
								})

								Convey("Then it should arrive on the TTN side", func() {
									select {
									case <-time.After(time.Second):
										So("Timeout Exceeded", ShouldBeFalse)
									case _, ok := <-msg:
										So(ok, ShouldBeTrue)
									}
								})
							})
						})

						Convey("When sending a disconnect message", func() {
							err := gateway.PublishDisconnect(&types.DisconnectMessage{
								GatewayID: "dev",
							})
							time.Sleep(10 * time.Millisecond)
							Convey("There should be no error", func() {
								So(err, ShouldBeNil)
							})
							Convey("The gateway should be disconnected", func() {
								So(b.gateways.Contains("dev"), ShouldBeFalse)
							})

							Convey("When sending another disconnect message", func() {
								err := gateway.PublishDisconnect(&types.DisconnectMessage{
									GatewayID: "dev",
								})
								time.Sleep(10 * time.Millisecond)
								Convey("There should be no error", func() {
									So(err, ShouldBeNil)
								})
								Convey("The gateway should still be disconnected", func() {
									So(b.gateways.Contains("dev"), ShouldBeFalse)
								})
							})
						})

						Convey("When stopping the Exchange", func() {
							b.Stop()
						})

					})
				})
			})

			Convey("When starting the Exchange", func() {
				b.Start(1, 10*time.Millisecond)

				Convey("When stopping the Exchange", func() {
					b.Stop()
				})
			})

		})

	})
}
