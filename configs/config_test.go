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
	"testing"

	"k8s.io/api/core/v1"
)

func TestFilterConfig_Filter(t *testing.T) {
	type fields struct {
		Reason    string
		Message   string
		MinCount  int32
		Type      string
		Action    string
		Kind      string
		Namespace string
		Name      string
		Exporters []string
	}
	type args struct {
		event *v1.Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "filter reason exactly",
			fields: fields{Reason: "Killed"},
			args:   args{event: &v1.Event{Reason: "Killed"}},
			want:   false,
		},
		{
			name:   "filter reason by regexp",
			fields: fields{Reason: "Killed|Killing"},
			args:   args{event: &v1.Event{Reason: "Killing"}},
			want:   false,
		},
		{
			name:   "filter reason by regexp",
			fields: fields{Reason: "Killed|Killing"},
			args:   args{event: &v1.Event{Reason: "Started"}},
			want:   true,
		},

		{
			name:   "filter message by regexp",
			fields: fields{Message: "Killing|Killed .*"},
			args:   args{event: &v1.Event{Message: "Killing reviews"}},
			want:   false,
		},
		{
			name:   "filter message by regexp",
			fields: fields{Message: "Killing|Killed .*"},
			args:   args{event: &v1.Event{Message: "Killed reviews"}},
			want:   false,
		},
		{
			name:   "filter message by regexp",
			fields: fields{Message: "Killing|Killed .*"},
			args:   args{event: &v1.Event{Message: "Started reviews"}},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &FilterConfig{
				Reason:    tt.fields.Reason,
				Message:   tt.fields.Message,
				MinCount:  tt.fields.MinCount,
				Type:      tt.fields.Type,
				Action:    tt.fields.Action,
				Kind:      tt.fields.Kind,
				Namespace: tt.fields.Namespace,
				Name:      tt.fields.Name,
				Exporters: tt.fields.Exporters,
			}
			filter.Init()
			if got := filter.Filter(tt.args.event); got != tt.want {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
