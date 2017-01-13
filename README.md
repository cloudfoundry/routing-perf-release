# Routing Performance Release

## About
A BOSH release containing benchmarking tools and utilities to setup the
performance testing for the [gorouter](https://github.com/cloudfoundry/gorouter)
and [TCP router](https://github.com/cloudfoundry-incubator/cf-tcp-router).

This release will deploy:
- a static backend app that returns 1kB of static data by default (configurable using `gostatic.response_size` property)
- a VM for running the load test
- processes that populate the routing table for both routers with test data

## Get the code

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

## Running the load tests

Deploy performance release on your environment. Run the below command

```
bosh run errand throughputramp
```

Errand will upload CPU stats and performance results to an S3 bucket specified in the manifest.

### Running Jupyter Notebook for displaying graphs

1. If you have not already done so, download this repo to your local machine.
1. Install [Docker](https://docs.docker.com/) locally.
1. Verify the installation by running `docker -v`.
1. Create a directory and download CPU stats and performance test files from the
   S3 bucket specified in the above section.
1. Rename CPU stats file to `cpuStats.csv` and performance test file to
   `perfResults.csv`. Currently the notebook is configured to look for files
   with these names. Keep track of what the file names were before to provide
   a reference point for multiple investigations.
1. Run the below command to start the Docker container. Replace
   `PATH_TO_ROUTING_PERF_RELEASE` with the actual path to this repo on your
   local machine. The `-v LOCAL_DIR:CONTAINER_DIR` command will mount a local
   directory on your machine to a volume located at `CONTAINER_DIR` inside
   this Docker container.

   ```
   docker run -it -p 8888:8888 -v PATH_TO_ROUTING_PERF_RELEASE/src/jupyter_notebook:/home/jovyan/work jupyter/scipy-notebook
   ```

1. The `docker` command will present a token URL that you should copy/paste
   into a browser to start the notebook session.
   For example:

   ```
   ...
   [I 19:17:28.023 NotebookApp] The Jupyter Notebook is running at: http://[all ip addresses on your system]:8888/?token=1ce634bd85e4101a74d4114880642381e4f7244af7843093
   [I 19:17:28.023 NotebookApp] Use Control-C to stop this server and shut down all kernels (twice to skip confirmation).
   [C 19:17:28.023 NotebookApp]

       Copy/paste this URL into your browser when you connect for the first time,
       to login with a token:
           http://localhost:8888/?token=1ce634bd85e4101a74d4114880642381e4f7244af7843093

   ```

1. Click on the file named `Performance_Data.ipynb`.
1. Click on the title menu `Cell` and click on `Run All` to regenerate the
   notebook outputs.
1. To compare the current data set with another follow these instructions
   1. Add CPU stats and performance results to folder $routing-perf-release-path/src/jupyter_notebook/
   1. Rename CPU stats file to `old_cpuStats.csv` and performance test file to `old_perfResults.csv`.
   1. Go to the Notebook server page and update variable `compareDatasets` to `True` and rerun all the cells


### Troubleshooting Jupyter Notebook

1. If you have problem connecting to `localhost:8888`, you could add the
   `-network=host` flag to the above Docker command. You will have to use the
   IP of the Docker host instead of `localhost` to connect to the Notebook
   server. To obtain the IP address of the Docker host, run the command
   `docker-machine ls`.
