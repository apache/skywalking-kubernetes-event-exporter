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
	"text/template"

	"github.com/sirupsen/logrus"
	k8score "k8s.io/api/core/v1"
	sw "skywalking.apache.org/repo/goapi/collect/event/v3"

	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
	"github.com/apache/skywalking-kubernetes-event-exporter/pkg/k8s"
)

func (tmplt *EventTemplate) render(ctx context.Context, swEvent *sw.Event, kEvent *k8score.Event) chan bool {
	done := make(chan bool)

	go func() {
		var templateCtx k8s.TemplateContext

		select {
		case templateCtx = <-k8s.Registry.GetContext(ctx, kEvent):
		case <-ctx.Done():
			done <- true
			return
		}

		if logger.Log.IsLevelEnabled(logrus.DebugLevel) {
			if bs, err := json.Marshal(templateCtx); err == nil {
				logger.Log.Debugf("template context %+v", string(bs))
			}
		}

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

		done <- true
	}()

	return done
}
