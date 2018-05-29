package controller

import (
	"github.com/influxdata/influxdb/client/v2"
	"time"
	"log"
)

const (
	logErrorCreatingMonitorBatch = "Error creating monitoring batch"
)

func (ctx *Context) CreateMonitor() *client.Client {
	// TODO remove hardcoded address
	monitorClient, err := client.NewUDPClient(client.UDPConfig{
		Addr: "192.168.64.26:30102",
	})
	if err != nil {
		ctx.Logger.Error(logErrorCreatingMonitorConnection, err)
	}

	ctx.InfluxDBClient = &monitorClient

	return &monitorClient
}

func (ctx Context) writePoint(measurementName string, tags map[string]string, fields map[string]interface{}) {
	// TODO - new batch every time??? hardcoded database???
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "tcp-proxy-pool",
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
