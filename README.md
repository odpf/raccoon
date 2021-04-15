# Raccoon
![build workflow](https://github.com/odpf/raccoon/actions/workflows/build.yaml/badge.svg)
![package workflow](https://github.com/odpf/raccoon/actions/workflows/package.yaml/badge.svg)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=apache)](LICENSE)
[![Version](https://img.shields.io/github/v/release/odpf/raccoon?logo=semantic-release)](Version)

Raccoon is high throughput, low-latency service that provides an API to ingest clickstream data from mobile apps, sites and publish it to Kafka. Raccoon uses the Websocket protocol for peer-to-peer communication and protobuf as the serialization format. It provides an event type agnostic API that accepts a batch (array) of events in protobuf format. Refer [here](https://github.com/odpf/proton/tree/main/odpf/raccoon) for proto definition format that Raccoon accepts.

<p align="center"><img src="./docs/assets/overview.png" /></p>

## Key Features

* **Event Agnostic:** Raccoon API is event agnostic. This allows you to push any event with any schema.
* **Metrics:** Built in monitoring includes latency and active connections.

To know more, follow the detailed [documentation](docs) 

## Usage

Explore the following resources to get started with Raccoon:

* [Guides](docs/guides) provides guidance on deployment and client sample.
* [Concepts](docs/concepts) describes all important Raccoon concepts.
* [Reference](docs/reference) contains details about configurations, metrics and other aspects of Raccoon.
* [Contribute](docs/contribute/contribution.md) contains resources for anyone who wants to contribute to Raccoon.

## Run with Docker
You need to have docker installed in your system. You have two options to run Raccoon on Docker. First is to use docker compose, second is to use image from docker hub. 

To use image from the docker hub, download Raccoon [docker image](https://hub.docker.com/r/odpf/raccoon/). Have kafka running on your local and run the following.
```
# Download docker image from docker hub
$ docker pull odpf/raccoon

# Run the following docker command with minimal config.
$ docker run -p 8080:8080 -e SERVER_WEBSOCKET_PORT=8080 -e SERVER_WEBSOCKET_CONN_UNIQ_ID_HEADER=x-user-id -e PUBLISHER_KAFKA_CLIENT_BOOTSTRAP_SERVERS=host.docker.internal:9093 -e EVENT_DISTRIBUTION_PUBLISHER_PATTERN=clickstream-%s-log odpf/raccoon
```

You can also use `docker-compose` on this repo. The `docker-compose` provides raccoon along with Kafka setup. Make sure to adjust the `application.yml` config to point to that kafka. To run it you can run the following.
```
# Run raccoon along with kafka setup
make docker-run
# Stop the docker compose
make docker-stop
```

## Running locally
Prerequisite:
- You need to have [GO](https://golang.org/) 1.14 or above installed
- You need `protoc` [installed](https://github.com/protocolbuffers/protobuf#protocol-compiler-installation)

```sh
# Clone the repo
$ git clone https://github.com/odpf/raccoon.git  

# Build the executable
$ make

# Configure env variables
$ vim application.yaml

# Run Raccoon
$ ./out/raccoon
```
**Note:** Read the detail of each configurations [here](/docs/reference/configuration.md).

## Running tests 
```sh
# Running unit tests
$ make test

# Running integration tests
$ cp application.yml.integration-test application.yml
$ make docker-run
$ INTEGTEST_BOOTSTRAP_SERVER=localhost:9094 INTEGTEST_HOST=ws://localhost:8080 INTEGTEST_TOPIC_FORMAT="clickstream-%s-log" go test ./integration -v
```

## Contribute

Development of Raccoon happens in the open on GitHub, and we are grateful to the community for contributing bugfixes and improvements. Read below to learn how you can take part in improving Raccoon.

Read our [contributing guide](docs/contribute/contribution.md) to learn about our development process, how to propose bugfixes and improvements, and how to build and test your changes to Raccoon.

To help you get your feet wet and get you familiar with our contribution process, we have a list of [good first issues](https://github.com/odpf/raccoon/labels/good%20first%20issue) that contain bugs which have a relatively limited scope. This is a great place to get started.

## Credits

This project exists thanks to all the [contributors](https://github.com/odpf/raccoon/graphs/contributors).

## License
Firehose is [Apache 2.0](LICENSE) licensed.