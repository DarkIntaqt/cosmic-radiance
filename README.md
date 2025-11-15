# Cosmic-Radiance

![Go Version](https://img.shields.io/badge/go-1.24.4+-blue?style=flat-square)
[![Issues](https://img.shields.io/github/issues/DarkIntaqt/cosmic-radiance?style=flat-square)](https://github.com/DarkIntaqt/cosmic-radiance/issues)
[![License](https://img.shields.io/github/license/darkintaqt/cosmic-radiance?style=flat-square)](https://github.com/DarkIntaqt/cosmic-radiance/blob/main/LICENSE)

Just another Riot Games API rate limiter with a cosmic glance ✨. 

## Features

- Automatic endpoint and platform detection
- Automatic rate limit discovery
- Customizable Timeout and good Retry-After handling
- GZIP handling to reduce traffic
- Prioritize requests with a `X-Priority: high` header
- Prometheus metrics to create dashboards about the rate limit status and queue sizes
- up to 99% close to uptime rate limits[^1]

###

## Getting started 

Click on the option that suits you best. For a normal production use-case, Docker is recommended.

<details>
<summary>Docker</summary>

### Docker

To get started with Docker, pull the image from the registry first:

```
docker pull ghcr.io/darkintaqt/cosmic-radiance:latest
```

Next, you need to set a few environment variables. For that, you can copy the .env.example and adjust the settings to your needs.  

#### Running using `docker run`

Finally, run the project through the CLI:

```
docker run --env-file ./env ghcr.io/darkintaqt/cosmic-radiance:latest
```

#### Running using `docker-compose`

Then, add cosmic-radiance to your docker-compose.yml file
```yml
services:
  cosmic-radiance:
    container_name: cosmic-radiance
    image: ghcr.io/darkintaqt/cosmic-radiance:latest
    ports:
      - "${PORT:-8001}:8001"
    environment:
      - API_KEY=${API_KEY:-}
      - MODE=${MODE:-}
      - TIMEOUT=${TIMEOUT:-}
      - PRIORITY_QUEUE_SIZE=${PRIORITY_QUEUE_SIZE:-}
      - PROMETHEUS=${PROMETHEUS:-}
# you can add more env variables using this schema
```

Finally, start the project using
```
docker compose up cosmic-radiance
```

Then, you can start requesting `http://localhost:PORT/<platform>/<method>` or `http://<platform>.api.riotgames.com/<method> (with proxy-pass)`, based on your `MODE` (see configuration). 

Keep in mind, that other docker container might need to be in the same docker network in order to use cosmic-radiance.

---

</details>

<details>
<summary>Go (CLI)</summary>

### Go (CLI)

To get started with Go, clone the project from GitHub.

```
git clone https://github.com/DarkIntaqt/cosmic-radiance.git
```

Next, you need to set a few environment variables. For that, you can copy the .env.example and adjust the settings to your needs.  
After that, you need to install the dependencies using:

```
go mod tidy
```

Finally, you can start the project with 

```
go run cmd/cosmic-radiance/main.go
```

Then, you can start requesting `http://localhost:PORT/<platform>/<method>` or `http://<platform>.api.riotgames.com/<method> (with proxy-pass)`, based on your `MODE` (see configuration). 

---

</details>

<details>
<summary>Go (package)</summary>

### Go (package)

To get started with using cosmic-radiance as a go package, install the package into your current workspace

```
go get github.com/DarkIntaqt/cosmic-radiance/ratelimiter
```

Next, you need to set a few environment variables. For that, you can copy the .env.example and adjust the settings to your needs.  

```go
package main

import (
   "github.com/DarkIntaqt/cosmic-radiance/ratelimiter"
)

func main() {
   port := 8080
   limiter := ratelimiter.Init(port)

   limiter.Start()

   // other logic
   limiter.Stop()
}
```

Then, you can start requesting `http://localhost:PORT/<platform>/<method>` or `http://<platform>.api.riotgames.com/<method> (with proxy-pass)`, based on your `MODE` (see configuration). 

---

</details>


> [!WARNING]
> If you want to deploy cosmic-radiance to production, please use a [tagged version](https://github.com/DarkIntaqt/cosmic-radiance/releases), since development may take place in the main branch.

## Configuration

There are several .env variables which can fine tune cosmic-radiance. Some are required

| Name                   | Function                                                                                                                                                                                                                                                                             |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| API_KEY **required**   | Your Riot Games API key. Cosmic-radiance needs your key to fire requests to the Riot Games API.                                                                                                                                                                                      |
| PORT **required**      | Port on which the proxy is running. Chose a port that is free. Please double-check your port and Dockerfile configuration.                                                                                                                                                           |
| MODE **required**      | Either `PATH` or `PROXY`. In path mode, you request cosmic-radiance like a normal webserver with the endpoint following the endpoint. In the proxy mode, you can use proxy-pass to redirect <platform>.api.riotgames.com requests directly to cosmic-radiance. You need to use http. |
| TIMEOUT                | The wait time after which incoming requests are getting rejected. Time in seconds                                                                                                                                                                                                    |
| PRIORITY_QUEUE_SIZE    | The size of the priority queue compared to the normal queue. In percent (%).                                                                                                                                                                                                         |
| PROMETHEUS             | Either `ON` or `OFF`. Disabled by default. Enable to get prometheus statistics                                                                                                                                                                                                       |
| POLLING_INTERVAL       | The time in milliseconds in which the main loop checks whether new requests can be fired and rate limits can be updated. Default is 10ms.                                                                                                                                            |
| ADDITIONAL_WINDOW_SIZE | The window size in milliseconds that gets added on top of Riot Games' windows in order to account for latency. Default is 125ms.                                                                                                                                                     |

Check the [.env.example](https://github.com/DarkIntaqt/cosmic-radiance/blob/main/.env.example) for a more detailed description. 

## Error Codes

All error codes are returned as by the Riot Games API. There were a few additional error codes added. 

|  Code   | Where to be found | What does this mean                                              |
| :-----: | ----------------- | ---------------------------------------------------------------- |
| **408** | Metrics           | The request timed out. Check the `Retry-After` header.           |
| **499** | Metrics           | The requesting client dropped the request.                       |
| **500** | Metrics and Proxy | The request to the Riot Games API failed before it was executed. |

## Contributing

This rate limiter is by no means feature complete, however it should be able to run in production without any issues. If you have any ideas or improvements, feel free to open an issue or a pull request. 

## License

This project is licensed under the Apache 2.0 License. 


[^1]: Tested with 100 available request per second (rps) in which the average was 99rps.
