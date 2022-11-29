# static

A simple HTTP server that serves static responses

## configuration

| parameter      | default   | description                                 |
|:---------------|:----------|:--------------------------------------------|
| HOSTNAME       | 127.0.0.1 | Bind hostname                               |
| PORT           | 8080      | Bind port                                   |
| LOG_LEVEL      | info      | Zerolog log level                           |
| LOG_PRETTY     | false     | Turn on to prettify logs (standard is JSON) |
| ENDPOINTS_PATH |           | Path to the endpoints configuration file    |

## entrypoint configuration

Configuring static endpoints is done with a yaml file.  
An example can be found [here](example/endpoints.yaml)