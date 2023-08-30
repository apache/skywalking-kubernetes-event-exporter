module github.com/apache/skywalking-kubernetes-event-exporter

go 1.16

require (
	github.com/hashicorp/golang-lru v0.5.4
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	google.golang.org/grpc v1.40.0
	gopkg.in/yaml.v3 v3.0.0
	k8s.io/api v0.20.5
	k8s.io/client-go v0.20.5
	skywalking.apache.org/repo/goapi v0.0.0-20220412071816-33e4ea2a99b4
)
