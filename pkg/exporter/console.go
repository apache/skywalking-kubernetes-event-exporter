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
	"encoding/json"
	"fmt"
	sw "skywalking.apache.org/repo/goapi/collect/event/v3"

	k8score "k8s.io/api/core/v1"

	"github.com/apache/skywalking-kubernetes-event-exporter/configs"
	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
	"github.com/apache/skywalking-kubernetes-event-exporter/pkg/event"
)

// Console Exporter exports the events into console logs, this exporter is typically
// used for debugging.
type Console struct {
	config  ConsoleConfig
	stopper chan struct{}
}

type ConsoleConfig struct {
	Template *EventTemplate `mapstructure:"template"`
}

func init() {
	s := &Console{
		stopper: make(chan struct{}),
	}
	RegisterExporter(s.Name(), s)
}

func (exporter *Console) Init() error {
	config := ConsoleConfig{}

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

	exporter.config = config

	return nil
}

func (exporter *Console) Name() string {
	return "console"
}

func (exporter *Console) Export(events chan *k8score.Event) {
	logger.Log.Debugf("exporting events into %+v", exporter.Name())

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
				if bytes, err := json.Marshal(swEvent); err != nil {
					logger.Log.Errorf("failed to send event to %+v, %+v", exporter.Name(), err)
				} else {
					logger.Log.Infoln(string(bytes))
				}
			}
		}
	}()
}

func (exporter *Console) Stop() {
	exporter.stopper <- struct{}{}
	close(exporter.stopper)
}
