package monitor

import (

)

//func (ctx Context) writeMeasurement(measurementName string, tags map[string]string, fields map[string]interface{}) {
//}

//func (ctx Context) writePoint(monitorClient client.Client,
//	// Create a new point batch
//	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
//		Database:  "tcp-proxy-pool",
//		Precision: "ns",
//	})
//	if err != nil {
//		ctx.Logger.Error(logErrorCreatingMonitorBatch, err)
//	}
//
//	pt, err := client.NewPoint(measurementName, tags, fields, time.Now())
//	if err != nil {
//		log.Fatal(err)
//	}
//	bp.AddPoint(pt)
//
//	if err := monitorClient.Write(bp); err != nil {
//		ctx.Logger.Error(err)
//	}
//}