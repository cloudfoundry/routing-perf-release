# Routing Performance Release

## About
A BOSH release containing benchmarking tools and utilities to setup the
performance testing for the [gorouter](https://github.com/cloudfoundry/gorouter)
and [TCP router](https://github.com/cloudfoundry-incubator/cf-tcp-router).

This release will deploy a static backend app that returns 1kB of static data
by default. This can be configured via the `gostatic.response_size` job
property for gostatic.

### Get the code

1. Fetch release repo

  ```
  mkdir -p ~/workspace
  cd ~/workspace
  git clone https://github.com/cloudfoundry-incubator/routing-perf-release.git
  cd routing-perf-release/
  ```

1. Automate `$GOPATH` and `$PATH` setup

  This BOSH release doubles as a `$GOPATH`. It will automatically be set up for you if you have [direnv](http://direnv.net) installed.

  ```
  direnv allow
  ```

  If you do not wish to use direnv, you can simply `source` the `.envrc` file in
  the root of the release repo.  You may manually need to update your `$GOPATH`
  and `$PATH` variables as you switch in and out of the directory.

## Deploying to BOSH on AWS

### Prerequisites

1. Install and start [BOSH on AWS](http://bosh.io/docs/init-aws.html).
1. Upload the latest AWS Trusty Go-Agent stemcell to BOSH. You can download it first if you prefer.

	```
	bosh upload stemcell https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent
	```

1. Make sure you have [NATS](https://github.com/cloudfoundry/nats-release/)
   and [gorouter](https://github.com/cloudfoundry/gorouter/) deployed.
   The `http_route_populator` job will need gorouter listening on NATS to
   populate the routing table with routes pointing to the static backend.
1. Make sure you have
   [Routing API and TCP Router](https://github.com/cloudfoundry-incubator/routing-release)
   deployed. The `tcp_route_populator` will require the Routing API in order
   to have TCP routes pointing to the static backend.

   > **NOTE**: The `routing_api.auth_disabled` property should be set to `true`
   > since the tcp_route_populator does not support grabbing a UAA token for
   > authentication.

### Upload Release, Create a Deployment Manifest, and Deploy
1. Clone this repo; see [Get the code](#get-the-code).

1. Create and upload the release
  ```sh
  cd ~/workspace/routing-perf-release/
  bosh create release
  bosh -n upload release
  ```

1. Fill out the [cloud-config](manifests/cloud-config-aws.yml) file and the
   [deployment manifest](manifests/perf.yml) with the proper values.

1. Update the cloud-config on your director. Beware that support for using v2
   manifests on the same director as v1 manifest deployments is supported
   after BOSH v257.
  ```sh
  bosh update cloud-config manifests/cloud-config-aws.yml
  ```

1. Deploy the release

  ```sh
  bosh -n -d manifests/perf.yml deploy
  ```

## Deploying to BOSH on BOSH Lite

This assumes you have running CF and Routing Release deployments running on
BOSH Lite and that BOSH Lite is updated to a recent version that can support
v2 manifest and v1 manifest deployments.

You can follow the above instructions with this
[cloud-config](manifests/cloud-config-bosh-lite.yml) and
[deployment manifest](manifests/perf-bosh-lite.yml)
Verify that the parameters in the deployment manifest are correct and simply
BOSH deploy.

## Deploying to BOSH on other IaaS's

If you are deploying this release on any other IaaS's, you can update the
[cloud-config](manifests/cloud-config-aws.yml) with the correct
`cloud_properties`. For more information, refer to the
[BOSH documentation](http://bosh.io/docs).
