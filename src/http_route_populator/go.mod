module github.com/cloudfoundry/routing-perf-release/http_route_populator

go 1.16

replace gopkg.in/fsnotify.v1 v1.4.7 => github.com/fsnotify/fsnotify v1.4.7

require (
	github.com/fsnotify/fsnotify v1.4.7
	github.com/hpcloud/tail v1.0.0
	github.com/nats-io/nats v1.8.1
	github.com/nats-io/nats.go v1.8.1
	github.com/nats-io/nkeys v0.1.0
	github.com/nats-io/nuid v1.0.1
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7
	golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7
	golang.org/x/text v0.3.2
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	gopkg.in/yaml.v2 v2.2.2
)
