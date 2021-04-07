# SkyWalking Kubernetes Event Exporter

[![GitHub stars](https://img.shields.io/github/stars/apache/skywalking-kubernetes-event-exporter.svg?style=for-the-badge&label=Stars&logo=github)](https://github.com/apache/skywalking-kubernetes-event-exporter)
[![Twitter Follow](https://img.shields.io/twitter/follow/asfskywalking.svg?style=for-the-badge&label=Follow&logo=twitter)](https://twitter.com/AsfSkyWalking)

[![Check](https://github.com/apache/skywalking-kubernetes-event-exporter/actions/workflows/build-and-test.yaml/badge.svg)](https://github.com/apache/skywalking-kubernetes-event-exporter/actions/workflows/build-and-test.yaml)

<img src="http://skywalking.apache.org/assets/logo.svg" alt="Sky Walking logo" height="90px" align="right" />

SkyWalking Kubernetes Event Exporter is able to watch, filter, and send Kubernetes events
into [Apache SkyWalking](https://github.com/apache/skywalking) backend, afterwards, SkyWalking associates the events
with the system metrics and thus gives you an overview about how the metrics are effected by the events.

## Configurations

Configurations are in YAML format, or config map if running inside Kubernetes,
otherwise, [the default configuration file](assets/default-config.yaml) will be used if there is neither `-c` option
specified in the command line interface nor config map created in Kubernetes.

All available configuration items and their documentations can be found
in [the default configuration file](assets/default-config.yaml).

## Deployments

Go to [the /deployments/release](deployments/release) directory, modify according to your needs, and
run `kustomize build | kubectl apply -f -`.

You can also simply run `skywalking-kubernetes-event-exporter start` in command line interface to run this exporter from
outside of Kubernetes.

## Build and Test

In order to build and test the exporter before an Apache official release, you need set a Docker registry where you can
push the images, do this by `export HUB=<your-docker-hub-registry>`, and then run `make -C build/package/docker push`
to build and push the Docker images, finally, run `make -C deployments/dev deploy` to deploy the exporter.

```shell
export HUB=<your-docker-hub-registry>
make -C build/package/docker push
make -C deployments/dev deploy
```

# Download

Go to the [download page](https://skywalking.apache.org/downloads/) to download all available binaries, including macOS,
Linux, Windows.

# Contact Us

* Mailing list: **dev@skywalking.apache.org**. Send email
  to [dev-subscribe@skywalking.apache.org](mailto:dev-subscribe@skywalking.apache.org), follow the reply to subscribe
  the mail list.
* Join `skywalking` channel at [Apache Slack](http://s.apache.org/slack-invite). If the link is not working, find the
  latest one at [Apache INFRA WIKI](https://cwiki.apache.org/confluence/display/INFRA/Slack+Guest+Invites).
* Twitter, [ASFSkyWalking](https://twitter.com/ASFSkyWalking)
* QQ Group: 901167865(Recommended), 392443393
* [bilibili B站 视频](https://space.bilibili.com/390683219)

# License

[Apache 2.0 License.](LICENSE)
