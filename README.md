# Cosmic-Radiance

Just another Riot Games API rate limiter with a cosmic glance ✨. 

## Features

- Automatic endpoint and platform detection
- Automatic rate limit discovery
- Customizable Timeout and good Retry-After handling
- GZIP handling to reduce traffic
- Prometheus metrics to create dashboards about the rate limit status and queue sizes
- up to 99% close to uptime rate limits[^1]

## Getting started

To get started, clone the project from GitHub.  
After that, install the dependencies with:

```sh
go mod tidy
```

Next, you need to set a few environment variables. For that, you can copy the .env.example and adjust the settings to your needs.  
Finally, run the project with

```sh
go run cmd/cosmic-radiance/main.go
```

## Error Codes

All error codes are returned as by the Riot Games API. There were a few additionall error codes added. 

|  Code   | Where to be found | What does this mean                                              |
| :-----: | ----------------- | ---------------------------------------------------------------- |
| **408** | Metrics           | The request timed out. Check the `Retry-After` header.           |
| **499** | Metrics           | The requesting client dropped the request.                       |
| **500** | Metrics and Proxy | The request to the Riot Games API failed before it was executed. |

## Contributing

This rate limiter is by no means feature complete, however it should be able to run in production without any issues. If you have any ideas or improvements, feel free to open an issue or a pull request. 

## License

This project is licensed under the Apache 2.0 License. 


---
[^1]: Tested with 100 available request per second (rps) in which the average was 99rps.