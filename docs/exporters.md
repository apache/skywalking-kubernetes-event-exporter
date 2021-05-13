# Exporters

## SkyWalking

[SkyWalking Exporter](../pkg/exporter/skywalking.go) exports the events into Apache SkyWalking OAP server.

The configurations of SkyWalking Exporter can be found [here](../assets/default-config.yaml).

## Console

[Console Exporter](../pkg/exporter/console.go) exports the events into console logs, this exporter is typically used for
debugging.

The configurations of Console Exporter can be found [here](../assets/default-config.yaml).
