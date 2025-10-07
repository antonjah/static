[![ci](https://github.com/antonjah/static/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/antonjah/static/actions/workflows/main.yml)
![Docker Pulls](https://img.shields.io/docker/pulls/antonjah/static?style=flat-square&logo=appveyor)
![Docker Image Size (tag)](https://img.shields.io/docker/image-size/antonjah/static/latest?style=flat-square&logo=appveyor)


# static

A simple HTTP server that serves static responses

## configuration

| parameter         | default   | description                                                               |
|:------------------|:----------|:--------------------------------------------------------------------------|
| HOSTNAME          | 127.0.0.1 | Bind hostname                                                             |
| PORT              | 8080      | Bind port                                                                 |
| LOG_LEVEL         | info      | Zerolog log level                                                         |
| LOG_PRETTY        | false     | Turn on to prettify logs (standard is JSON)                               |
| ENDPOINTS_PATH    | $PWD      | Path to the endpoints configuration file (defaults to current directory)  |
| TLS_ENABLED       | false     | Enable TLS                                                                |
| TLS_CERTIFICATE   |           | Path to TLS certificate (required if TLS_ENABLED is true)                 |
| TLS_KEY           |           | Path to TLS key (required if TLS_ENABLED is true)                         |
| TLS_CA            |           | Path to TLS CA                                                            |
| TLS_VERIFY_CLIENT | false     | Enable client certificate verification (requires TLS_CA)                  |

## endpoints configuration

Configuring static endpoints is done with a yaml file.  
An example can be found [here](example/endpoints.yaml)

There is also a JSON schema provided [here](example/schema.json) that can be used to validate the configuration.

A very basic example with one path and one method:

```yaml
endpoints:
  - path: /tea/pot
    methods:
      - method: "GET"
        status-code: 418
        body: "I'm a teapot"
        headers:
          content-type: "text/plain"
```

Example request:

```log
$ curl -i localhost:8080/tea/pot

HTTP/1.1 418 I'm a teapot
Content-Type: text/plain
Date: Wed, 30 Nov 2022 08:59:12 GMT
Content-Length: 12

I'm a teapot
```

## building, testing, and linting

Building and testing is done using the provided [Makefile](Makefile):

```shell
$ make help
                                                                                                                                              
  build                          Build mockserver
  buildimage                     Build docker image
  help                           Show help
  lint                           Run linting
  tagimage                       Tag the docker image
  test                           Run linting and unittest
  unittest                       Run unittest
```
