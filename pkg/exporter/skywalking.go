/*
 * Licensed to Apache Software Foundation (ASF) under one or more contributor
 * license agreements. See the NOTICE file distributed with
 * this work for additional information regarding copyright
 * ownership. Apache Software Foundation (ASF) licenses this file to you under
 * the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	sw "skywalking.apache.org/repo/goapi/collect/event/v3"

	"google.golang.org/grpc"
	k8score "k8s.io/api/core/v1"

	"github.com/apache/skywalking-kubernetes-event-exporter/configs"
	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
)

// SkyWalking Exporter exports the events into Apache SkyWalking OAP server.
type SkyWalking struct {
	config SkyWalkingConfig
	client sw.EventServiceClient
}

type SkyWalkingConfig struct {
	Address  string         `mapstructure:"address"`
	Template *EventTemplate `mapstructure:"template"`
}

func init() {
	s := &SkyWalking{}
	RegisterExporter(s.Name(), s)
}

func (exporter *SkyWalking) Init(ctx context.Context) error {
	config := SkyWalkingConfig{}

	if c := configs.GlobalConfig.Exporters[exporter.Name()]; c == nil {
		return fmt.Errorf("configs of %+v exporter cannot be empty", exporter.Name())
	} else if marshal, err := json.Marshal(c); err != nil {
		return err
	} else if err := json.Unmarshal(marshal, &config); err != nil {
		return err
	}

	if err := config.Template.Init(); err != nil {
		return err
	}

	conn, err := grpc.Dial(config.Address, grpc.WithInsecure())
	if err != nil {
		return err
	}

	exporter.config = config
	exporter.client = sw.NewEventServiceClient(conn)

	go func() {
		<-ctx.Done()

		if err := conn.Close(); err != nil {
			logger.Log.Errorf("failed to close connection. %+v", err)
		}
	}()

	return nil
}

func (exporter *SkyWalking) Name() string {
	return "skywalking"
}

func (exporter *SkyWalking) Export(ctx context.Context, events chan *k8score.Event) {
	logger.Log.Debugf("exporting events into %+v", exporter.Name())

	stream, err := exporter.client.Collect(ctx)

	for err != nil {
		select {
		case <-ctx.Done():
			logger.Log.Debugf("stopping exporter %+v", exporter.Name())
			if err = stream.CloseSend(); err != nil {
				logger.Log.Warnf("failed to close stream. %+v", err)
			}
			return
		default:
			logger.Log.Errorf("failed to connect to SkyWalking server. %+v", err)
			time.Sleep(3 * time.Second)
			stream, err = exporter.client.Collect(ctx)
		}
	}

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debugf("stopping exporter %+v", exporter.Name())
			return
		case kEvent := <-events:
			if logger.Log.IsLevelEnabled(logrus.DebugLevel) {
				if bytes, err := json.Marshal(kEvent); err == nil {
					logger.Log.Debugf("exporting event to %v: %v", exporter.Name(), string(bytes))
				}
			}

			t := sw.Type_Normal
			if kEvent.Type == k8score.EventTypeWarning {
				t = sw.Type_Error
			}
			swEvent := &sw.Event{
				Uuid:      string(kEvent.UID),
				Source:    &sw.Source{},
				Name:      kEvent.Reason,
				Type:      t,
				Message:   kEvent.Message,
				StartTime: kEvent.FirstTimestamp.UnixNano() / 1000000,
				EndTime:   kEvent.LastTimestamp.UnixNano() / 1000000,
			}
			if exporter.config.Template != nil {
				go func() {
					renderCtx, cancel := context.WithTimeout(ctx, time.Minute)
					done := exporter.config.Template.render(renderCtx, swEvent, kEvent)
					select {
					case <-done:
						logger.Log.Debugf("done: rendered event is: %+v", swEvent)
						exporter.export(stream, swEvent)
					case <-renderCtx.Done():
						logger.Log.Debugf("canceled: rendered event is: %+v", swEvent)
						exporter.export(stream, swEvent)
					}
					cancel()
				}()
			} else {
				exporter.export(stream, swEvent)
			}
		}
	}
}

func (exporter SkyWalking) export(stream sw.EventService_CollectClient, swEvent *sw.Event) {
	if err := stream.Send(swEvent); err != nil {
		logger.Log.Errorf("failed to send event to %+v. %+v", exporter.Name(), err)
	}
}
