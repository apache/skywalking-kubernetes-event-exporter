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

package pipe

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	"github.com/apache/skywalking-kubernetes-event-exporter/configs"
	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
	"github.com/apache/skywalking-kubernetes-event-exporter/pkg/event"
	exp "github.com/apache/skywalking-kubernetes-event-exporter/pkg/exporter"
	"github.com/apache/skywalking-kubernetes-event-exporter/pkg/k8s"
)

type workflow struct {
	filter   *configs.FilterConfig
	exporter exp.Exporter
	events   chan *v1.Event
}

type Pipe struct {
	Watcher   *k8s.EventWatcher
	workflows []workflow
	stopper   chan struct{}
}

func (p *Pipe) Init() error {
	p.stopper = make(chan struct{})
	p.workflows = []workflow{}

	initialized := map[string]bool{}
	for _, filter := range configs.GlobalConfig.Filters {
		if _, ok := configs.GlobalConfig.Exporters[filter.Exporter]; !ok {
			return fmt.Errorf("exporter %v is not defined", filter.Exporter)
		}
		exporter := exp.GetExporter(filter.Exporter)
		if exporter == nil {
			return fmt.Errorf("exporter %v is not defined", filter.Exporter)
		}
		if initialized[filter.Exporter] {
			continue
		}
		if err := exporter.Init(); err != nil {
			return err
		}
		initialized[filter.Exporter] = true

		events := make(chan *v1.Event)

		p.workflows = append(p.workflows, workflow{
			filter:   filter,
			exporter: exporter,
			events:   events,
		})
	}

	return nil
}

func (p *Pipe) Start() error {
	p.Watcher.Start()

	for _, wkfl := range p.workflows {
		go wkfl.exporter.Export(wkfl.events)
	}

	for stopped := false; !stopped; {
		select {
		case <-p.stopper:
			for _, w := range p.workflows {
				w.events <- event.Stopper
			}
			stopped = true
		case e := <-p.Watcher.Events:
			for _, w := range p.workflows {
				if w.filter.Filter(e) {
					continue
				}
				w.events <- e
			}
		}
	}

	return nil
}

func (p *Pipe) Stop() {
	p.Watcher.Stop()

	for _, w := range p.workflows {
		logger.Log.Debugf("stopping exporter %+v.", w.exporter.Name())
		w.exporter.Stop()
	}

	p.stopper <- struct{}{}
	close(p.stopper)
}
