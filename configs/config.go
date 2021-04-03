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
	"regexp"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"

	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"

	evnt "github.com/apache/skywalking-kubernetes-event-exporter/pkg/event"
)

type FilterConfig struct {
	Reason          string `yaml:"reason"`
	reasonRegExp    *regexp.Regexp
	Message         string `yaml:"message"`
	messageRegExp   *regexp.Regexp
	MinCount        int32  `yaml:"min-count"`
	Type            string `yaml:"type"`
	typeRegExp      *regexp.Regexp
	Action          string `yaml:"action"`
	actionRegExp    *regexp.Regexp
	Kind            string `yaml:"kind"`
	kindRegExp      *regexp.Regexp
	Namespace       string `yaml:"namespace"`
	namespaceRegExp *regexp.Regexp
	Name            string `yaml:"name"`
	nameRegExp      *regexp.Regexp

	Exporters []string `yaml:"exporters"`
}

func (filter *FilterConfig) Init() {
	logger.Log.Debugf("initializing filter config")

	filter.reasonRegExp = regexp.MustCompile(filter.Reason)
	filter.messageRegExp = regexp.MustCompile(filter.Message)
	filter.typeRegExp = regexp.MustCompile(filter.Type)
	filter.actionRegExp = regexp.MustCompile(filter.Action)
	filter.kindRegExp = regexp.MustCompile(filter.Kind)
	filter.namespaceRegExp = regexp.MustCompile(filter.Namespace)
	filter.nameRegExp = regexp.MustCompile(filter.Name)
}

// Filter the given event with this filter instance.
// Return true if the event is filtered, return false otherwise.
func (filter *FilterConfig) Filter(event *v1.Event) bool {
	if event == evnt.Stopper {
		return false
	}
	if filter.Reason != "" && !filter.reasonRegExp.MatchString(event.Reason) {
		return true
	}
	if filter.Message != "" && !filter.messageRegExp.MatchString(event.Message) {
		return true
	}
	if filter.Type != "" && !filter.typeRegExp.MatchString(event.Type) {
		return true
	}
	if filter.Action != "" && !filter.actionRegExp.MatchString(event.Action) {
		return true
	}
	if filter.Kind != "" && !filter.kindRegExp.MatchString(event.Kind) {
		return true
	}
	if filter.Namespace != "" && !filter.namespaceRegExp.MatchString(event.Namespace) {
		return true
	}
	if filter.Name != "" && !filter.nameRegExp.MatchString(event.Name) {
		return true
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
