module github.com/apache/skywalking-kubernetes-event-exporter

go 1.16

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	google.golang.org/grpc v1.36.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.5
	k8s.io/client-go v0.20.5
	skywalking.apache.org/repo/goapi v0.0.0-20210401043526-44170b5d980b
)
