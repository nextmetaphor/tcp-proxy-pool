package controller

import (
	"github.com/influxdata/influxdb/client/v2"
	"time"
	"log"
	"strings"
)

const (
	logErrorCreatingMonitorBatch = "Error creating monitoring batch"
)

func (ctx *Context) CreateMonitor() *client.Client {
	if (ctx == nil) || (strings.TrimSpace(ctx.Settings.Monitor.Address) == "") {
		return nil
	}

	monitorClient, err := client.NewUDPClient(client.UDPConfig{
		Addr: ctx.Settings.Monitor.Address,
	})
	if err != nil {
		ctx.Logger.Error(logErrorCreatingMonitorConnection, err)
	}

	ctx.InfluxDBClient = &monitorClient

	return &monitorClient
}

func (ctx *Context) WritePoint(measurementName string, tags map[string]string, fields map[string]interface{}) {
	if (ctx == nil) || (strings.TrimSpace(ctx.Settings.Monitor.Address) == "") {
		return
	}

	// TODO - new batch every time???
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  ctx.Settings.Monitor.Database,
		Precision: "ns",
	})
	if err != nil {
		ctx.Logger.Error(logErrorCreatingMonitorBatch, err)
	}

	pt, err := client.NewPoint(measurementName, tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	if err := (*ctx.InfluxDBClient).Write(bp); err != nil {
		ctx.Logger.Error(err)
	}
}
