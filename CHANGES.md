# Change Log

## 1.1.0

- Bump gRPC to 1.83.2, logrus to 1.9.4 and yaml.v3 to 3.0.1 to fix security alerts. Building now requires Go 1.25.
- Bump golang.org/x/net to 0.62.0 and golang.org/x/crypto to 0.58.0, and build with Go 1.26 to fix security alerts in the dependencies and the Go standard library. Building now requires Go 1.26.

## 1.0

- Add Apache SkyWalking exporter to export events into SkyWalking OAP.
- Add console exporter for debugging purpose.
