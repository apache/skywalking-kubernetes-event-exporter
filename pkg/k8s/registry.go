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
	"context"
	"fmt"

	"github.com/apache/skywalking-kubernetes-event-exporter/internal/pkg/logger"

	lru "github.com/hashicorp/golang-lru"
	corev1 "k8s.io/api/core/v1"

	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type id struct {
	namespace string
	name      string
}

func (i id) String() string {
	return fmt.Sprintf("%+v:%+v", i.namespace, i.name)
}

type registry struct {
	informers []cache.SharedIndexInformer

	podIDIpMap *lru.Cache // map[id]string
	idSvcMap   *lru.Cache // map[id]*corev1.Service
	idPodMap   *lru.Cache // map[id]*corev1.Pod
	ipSvcIDMap *lru.Cache // map[string]id
}

func (r registry) OnAdd(obj interface{}) {
	switch o := obj.(type) {
	case *corev1.Pod:
		podID := id{namespace: o.Namespace, name: o.Name}.String()

		r.podIDIpMap.Add(podID, o.Status.PodIP)
		r.idPodMap.Add(podID, o)
	case *corev1.Service:
		svcID := id{namespace: o.Namespace, name: o.Name}.String()

		r.idSvcMap.Add(svcID, o)
	case *corev1.Endpoints:
		for _, subset := range o.Subsets {
			for _, address := range subset.Addresses {
				svcID := id{namespace: o.ObjectMeta.Namespace, name: o.ObjectMeta.Name}.String()
				r.ipSvcIDMap.Add(address.IP, svcID)
			}
		}
	}
}

func (r registry) OnUpdate(oldObj, newObj interface{}) {
	r.OnDelete(oldObj)
	r.OnAdd(newObj)
}

func (r registry) OnDelete(_ interface{}) {
}

func (r *registry) Start(ctx context.Context) {
	logger.Log.Debugf("starting registry")

	for _, informer := range r.informers {
		go informer.Run(ctx.Done())
	}
}

type TemplateContext struct {
	Service *corev1.Service
	Pod     *corev1.Pod
	Event   *corev1.Event
}

func (r *registry) GetContext(ctx context.Context, e *corev1.Event) chan TemplateContext {
	resultCh := make(chan TemplateContext)

	result := TemplateContext{
		Event:   e,
		Pod:     &corev1.Pod{},
		Service: &corev1.Service{},
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				resultCh <- result
				return
			default:
				if obj := e.InvolvedObject; obj.Kind == "Pod" {
					podID := id{namespace: obj.Namespace, name: obj.Name}.String()

					pod, ok := r.idPodMap.Get(podID)
					if !ok {
						break
					}
					result.Pod = pod.(*corev1.Pod)

					podIP, ok := r.podIDIpMap.Get(podID)
					if !ok {
						break
					}

					svcID, ok := r.ipSvcIDMap.Get(podIP.(string))
					if !ok {
						break
					}

					svc, ok := r.idSvcMap.Get(svcID.(string))
					if !ok {
						break
					}
					result.Service = svc.(*corev1.Service)
				} else if obj.Kind == "Service" {
					svcID := id{namespace: obj.Namespace, name: obj.Name}.String()

					svc, ok := r.idSvcMap.Get(svcID)
					if !ok {
						break
					}
					result.Service = svc.(*corev1.Service)
				}

				resultCh <- result

				return
			}
			time.Sleep(3 * time.Second)
		}
	}()

	return resultCh
}

var Registry = &registry{}

func (r *registry) Init() error {
	logger.Log.Debugf("initializing template context registry")

	var err error

	if Registry.podIDIpMap, err = lru.New(1000); err != nil {
		return err
	}
	if Registry.idSvcMap, err = lru.New(1000); err != nil {
		return err
	}
	if Registry.idPodMap, err = lru.New(1000); err != nil {
		return err
	}
	if Registry.ipSvcIDMap, err = lru.New(1000); err != nil {
		return err
	}

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
