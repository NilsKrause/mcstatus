package api

import (
	"bytes"
	"encoding/json"
	mcclient "git.0cd.xyz/michael/mcstatus/client"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net"
	"net/http"
)

func initStatus (tournament *echo.Group) {
	tournament.GET("/status", getStatus)
	tournament.GET("/status/", getStatus)
}


func getStatus(c echo.Context) error {
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
	}

	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	type Request struct{Address string; Port int; Version uint64}
	request := &Request{}
	if err := c.Bind(request); err != nil {
		return err
	}

	client, err := mcclient.New(request.Address, request.Port, request.Version)

	if err != nil {
		return err
	}

	defer func(Conn net.Conn) {
		_ = Conn.Close()
	}(client.Conn)

	status, err := client.GetStatus()
	if err != nil {
		return err
	}

	b, _ := json.Marshal(&status)

	return c.String(http.StatusOK, string(b))
}