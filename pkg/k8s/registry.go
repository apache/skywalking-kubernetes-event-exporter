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

package k8s

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"
)

type id struct {
	namespace string
	name      string
}

type registry struct {
	informers []cache.SharedIndexInformer
	stopCh    chan struct{}

	podIDIpMap map[id]string
	idSvcMap   map[id]*corev1.Service
	idPodMap   map[id]*corev1.Pod
	ipSvcIDMap map[string]id
}

func (r registry) OnAdd(obj interface{}) {
	switch o := obj.(type) {
	case *corev1.Pod:
		podID := id{namespace: o.Namespace, name: o.Name}
		r.podIDIpMap[podID] = o.Status.PodIP
		r.idPodMap[podID] = o
	case *corev1.Service:
		r.idSvcMap[id{namespace: o.Namespace, name: o.Name}] = o
	case *corev1.Endpoints:
		for _, subset := range o.Subsets {
			for _, address := range subset.Addresses {
				r.ipSvcIDMap[address.IP] = id{
					namespace: o.ObjectMeta.Namespace,
					name:      o.ObjectMeta.Name,
				}
			}
		}
	}
}

func (r registry) OnUpdate(oldObj, newObj interface{}) {
	r.OnDelete(oldObj)
	r.OnAdd(newObj)
}

func (r registry) OnDelete(obj interface{}) {
	switch o := obj.(type) {
	case *corev1.Pod:
		podID := id{namespace: o.Namespace, name: o.Name}
		go func() {
			time.Sleep(3 * time.Second)
			delete(r.podIDIpMap, podID)
			delete(r.idPodMap, podID)
		}()
	case *corev1.Service:
		go func() {
			time.Sleep(3 * time.Second)
			delete(r.idSvcMap, id{namespace: o.Namespace, name: o.Name})
		}()
	case *corev1.Endpoints:
		go func() {
			for _, subset := range o.Subsets {
				for _, address := range subset.Addresses {
					time.Sleep(3 * time.Second)
					delete(r.ipSvcIDMap, address.IP)
				}
			}
		}()
	}
}

func (r *registry) Start() {
	logger.Log.Debugf("starting registry")

	for _, informer := range r.informers {
		go informer.Run(r.stopCh)
	}
}

func (r *registry) Stop() {
	logger.Log.Debugf("stopping registry")

	r.stopCh <- struct{}{}
	close(r.stopCh)
}

type TemplateContext struct {
	Service *corev1.Service
	Pod     *corev1.Pod
	Event   *corev1.Event
}

func (r *registry) GetContext(e *corev1.Event) TemplateContext {
	result := TemplateContext{Event: e}

	if obj := e.InvolvedObject; obj.Kind == "Pod" {
		podID := id{
			namespace: obj.Namespace,
			name:      obj.Name,
		}
		podIP := r.podIDIpMap[podID]
		svcID := r.ipSvcIDMap[podIP]

		result.Pod = r.idPodMap[podID]
		result.Service = r.idSvcMap[svcID]
	}

	if obj := e.InvolvedObject; obj.Kind == "Service" {
		svcID := id{
			namespace: obj.Namespace,
			name:      obj.Name,
		}
		result.Service = r.idSvcMap[svcID]
	}

	if result.Pod == nil {
		result.Pod = &corev1.Pod{}
	}
	if result.Service == nil {
		result.Service = &corev1.Service{}
	}

	return result
}

var Registry = &registry{
	stopCh: make(chan struct{}),

	podIDIpMap: make(map[id]string),
	idSvcMap:   make(map[id]*corev1.Service),
	idPodMap:   make(map[id]*corev1.Pod),
	ipSvcIDMap: make(map[string]id),
}

func (r *registry) Init() error {
	logger.Log.Debugf("initializing template context registry")

	config, err := GetConfig()
	if err != nil {
		return err
	}
	client := kubernetes.NewForConfigOrDie(config)
	factory := informers.NewSharedInformerFactoryWithOptions(client, 0, informers.WithNamespace(corev1.NamespaceAll))

	r.informers = []cache.SharedIndexInformer{
		factory.Core().V1().Endpoints().Informer(),
		factory.Core().V1().Services().Informer(),
		factory.Core().V1().Pods().Informer(),
	}

	for _, informer := range Registry.informers {
		informer.AddEventHandler(Registry)
	}

	return nil
}
