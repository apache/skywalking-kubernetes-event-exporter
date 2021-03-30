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

package configs

import (
	evnt "github.com/apache/skywalking-kubernetes-event-exporter/pkg/event"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
)

type FilterConfig struct {
	Reason   string `yaml:"reason"`
	Message  string `yaml:"message"`
	MinCount int32  `yaml:"min-count"`
	Type     string `yaml:"type"`
	Action   string `yaml:"action"`

	Kind      string `yaml:"kind"`
	Namespace string `yaml:"namespace"`
	Name      string `yaml:"name"`

	Exporter string `yaml:"exporter"`
}

func (filter *FilterConfig) Filter(event *v1.Event) bool {
	if event == evnt.Stopper {
		return false
	}
	return false
}

type ExporterConfig map[string]interface{}

type Config struct {
	Filters   []*FilterConfig           `mapstructure:"filters"`
	Exporters map[string]ExporterConfig `mapstructure:"exporters"`
}

var GlobalConfig Config

func ParseConfig(content []byte) error {
	return yaml.Unmarshal(content, &GlobalConfig)
}
