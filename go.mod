module github.com/apache/skywalking-kubernetes-event-exporter

go 1.16

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	google.golang.org/grpc v1.36.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.5
	k8s.io/client-go v0.20.5
	skywalking/network v0.0.0-00010101000000-000000000000
)

replace skywalking/network => ./api/skywalking/network
