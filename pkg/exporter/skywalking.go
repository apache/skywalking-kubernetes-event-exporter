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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	sw "skywalking.apache.org/repo/goapi/collect/event/v3"

	"github.com/apache/skywalking-kubernetes-event-exporter/pkg/k8s"

	"google.golang.org/grpc"
	k8score "k8s.io/api/core/v1"

	"github.com/apache/skywalking-kubernetes-event-exporter/configs"
	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
	"github.com/apache/skywalking-kubernetes-event-exporter/pkg/event"
)

// SkyWalking Exporter exports the events into Apache SkyWalking OAP server.
type SkyWalking struct {
	config     SkyWalkingConfig
	client     sw.EventServiceClient
	connection *grpc.ClientConn
	stopper    chan struct{}
}

type SkyWalkingConfig struct {
	Address  string         `mapstructure:"address"`
	Template *EventTemplate `mapstructure:"template"`
}

func init() {
	s := &SkyWalking{
		stopper: make(chan struct{}),
	}
	RegisterExporter(s.Name(), s)
}

func (exporter *SkyWalking) Init() error {
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
	exporter.connection = conn
	exporter.client = sw.NewEventServiceClient(conn)

	return nil
}

func (exporter *SkyWalking) Name() string {
	return "skywalking"
}

// TODO error handling
func (exporter *SkyWalking) Export(events chan *k8score.Event) {
	logger.Log.Debugf("exporting events into %+v", exporter.Name())

	stream, err := exporter.client.Collect(context.Background())
	for err != nil {
		select {
		case <-exporter.stopper:
			drain(events)
			return
		default:
			logger.Log.Errorf("failed to connect to SkyWalking server. %+v", err)
			time.Sleep(3 * time.Second)
			stream, err = exporter.client.Collect(context.Background())
		}
	}

	defer func() {
		if err := stream.CloseSend(); err != nil {
			logger.Log.Warnf("failed to close stream. %+v", err)
		}
	}()

	func() {
		for {
			select {
			case <-exporter.stopper:
				drain(events)
				return
			case kEvent := <-events:
				if kEvent == event.Stopper {
					return
				}
				logger.Log.Debugf("exporting event to %v: %v", exporter.Name(), kEvent)

				t := sw.Type_Normal
				if kEvent.Type == "Warning" {
					t = sw.Type_Error
				}
				swEvent := &sw.Event{
					Uuid:      string(kEvent.UID),
					Source:    &sw.Source{},
					Name:      kEvent.Reason,
					Type:      t,
					Message:   kEvent.Message,
					StartTime: kEvent.FirstTimestamp.UnixNano() / 1000000,
					EndTime:   kEvent.LastTimestamp.Unix() / 1000000,
				}
				if exporter.config.Template != nil {
					exporter.config.Template.render(swEvent, kEvent)
					logger.Log.Debugf("rendered event is: %+v", swEvent)
				}
				if err := stream.Send(swEvent); err != nil {
					logger.Log.Errorf("failed to send event to %+v. %+v", exporter.Name(), err)
				}
			}
		}
	}()
}

func (tmplt *EventTemplate) render(swEvent *sw.Event, kEvent *k8score.Event) {
	templateCtx := k8s.Registry.GetContext(kEvent)

	logger.Log.Debugf("template context %+v", templateCtx)

	render := func(t *template.Template, destination *string) error {
		if t == nil {
			return nil
		}

		var buf bytes.Buffer

		if err := t.Execute(&buf, templateCtx); err != nil {
			return err
		}

		if buf.Len() > 0 {
			*destination = buf.String()
		}

		return nil
	}

	if err := render(tmplt.messageTemplate, &swEvent.Message); err != nil {
		logger.Log.Debugf("failed to render the template, using the default event content. %+v", err)
	}

	if err := render(tmplt.sourceTemplate.serviceTemplate, &swEvent.Source.Service); err != nil {
		logger.Log.Debugf("failed to render service template, using the default event content. %+v", err)
	}
	if err := render(tmplt.sourceTemplate.serviceInstanceTemplate, &swEvent.Source.ServiceInstance); err != nil {
		logger.Log.Debugf("failed to render service instance template, using the default event content. %+v", err)
	}
	if err := render(tmplt.sourceTemplate.endpointTemplate, &swEvent.Source.Endpoint); err != nil {
		logger.Log.Debugf("failed to render endpoin template, using the default event content. %+v", err)
	}
}

func (exporter *SkyWalking) Stop() {
	exporter.stopper <- struct{}{}
	close(exporter.stopper)

	if err := exporter.connection.Close(); err != nil {
		logger.Log.Errorf("failed to close connection. %+v", err)
	}
}
