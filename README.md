# SkyWalking Kubernetes Event Exporter

SkyWalking Kubernetes Event Exporter is able to watch, filter, and send Kubernetes events into SkyWalking backend,
afterwards, SkyWalking associates the events with the system metrics and thus gives you an overview about how the
metrics are effected by the events.

## Configurations

Configurations are in YAML format, or config map if running inside Kubernetes,
otherwise, [the default configuration file](assets/default-config.yaml) will be used if there is neither `-c` option
specified in the command line interface nor config map is created in Kubernetes.

All available configuration items and their documentations can be found
in [the default configuration file](assets/default-config.yaml).

## Deployments

Go to [the /deployments](deployments) directory, modify according to your needs,
and `kubectl apply -f skywalking-kubernetes-event-exporter.yaml`.

You can also simply run `skywalking-kubernetes-event-exporter start` in command line interface to run this exporter from
outside of Kubernetes.
