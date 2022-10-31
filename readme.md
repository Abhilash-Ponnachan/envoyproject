
### Commands
```bash
# retag envoy image
$ docker tag envoyproxy/envoy:v1.23-latest envoyproxy/envoy

$ docker image ls
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
envoyproxy/envoy    latest              71f9c29a67ae        3 weeks ago         132MB
envoyproxy/envoy    v1.23-latest        71f9c29a67ae        3 weeks ago         132MB
...

# run envoy docker image
$ docker run --rm -d --name=envoy -p 10000:10000 envoyproxy/envoy
bfc7c9ab4cd378f0c32a5f3a521f9cf0742067708a1be002ccc632a89f9da9d8
# default port for envoy is 10000, publish to host on same

```

Check browser http://localhost:10000

>>> proxied to "Envoy home page"

### Commands
```bash
# exec into 'evnoy' container 
$ docker exec -it envoy sh 
# check running processes
\# ps -aux
USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
envoy          1  1.5  0.5 2363152 43748 ?       Ssl  12:35   0:01 envoy -c /etc/envoy/envoy.yaml
# Note the command `envoy -c <pathto config yaml>`
# cehck the config file
\# ls /etc/envoy
envoy.yaml

# create host dir 'proxy'
$ mkdir proxy

# copy envoy-container /etc/envoy/ content to proxy/config
$ docker cp envoy:/etc/envoy/ proxy/config

$ tree
.
├── proxy
│   └── config
│       └── envoy.yaml
└── readme.md

# make a copy of the envoy.yaml file
$ cp proxy/config/envoy.yaml proxy/config/envoy_org.yaml

# view contents of proxy/config/envoy.yaml
...

# edit occurances of www.envoyproxy.io with -> www.google.com
$ vim proxy/config/envoy.yaml

# edit, find and replace 

# stop running envoy container
$ docker stop envoy

# run envoy contianer again same as before, but this time mount config from host
$ docker run --rm -d --name=envoy -p 10000:10000 -v $(pwd)/proxy/config:/etc/envoy envoyproxy/envoy
873b22138de171de6989bfbc008830b18868220f11e9e72e0cc57fbc1efce065
```

Now again check browser http://localhost:1000

>>> proxied to "Google Search home page"

### Fresh Configuration
Make a configuration from scratch. Target different backends (upstream) servers.

Simple backend app in Go.
Shows a simple web page with the Name of the app, and the Host it is running on. By default it runs on PORT 8080, but that can be customised via Env var.
The AppName, background and foreground colours can be customised via Env vars.

We containerise it and run it as Docker containers.

Build docker image
```bash
$ docker build -t demoapp .

$ docker image ls
```

Run the app and expose it on Port 1111
```bash
$ docker run --rm --name=demoapp1 -p 1111:8080 demoapp
```
>>> show screen image

Run with diff app name and colours
```bash
$ docker run --rm --name=demoapp2 -p 1112:8080 -e "APPNAME=App 2" -e "BGCOLOR=green" demoapp
```
>>> show screen image

### Basic Components of Envoy Configuration

  Listeners -> Filters -> Routes -> Clusters -> Endpoints

>>> Diagram

  Description (summary)

### Simple Static Reverse Proxy Config 

A simple reverse proxy configuration, with just one backend app (cluster)

#### Empty Listener

First, a listener that does nothing! Edit our `prox/config/envoy.yaml` as shown below.
```yaml
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 8080
    filter_chains: [{}]
```
Now Run `Envoy proxy` again with Docker (expose `Port 8080` as that is the listening port we specified in the config above) using the <a name="#envoy-docker-run-mount">following command</a>.

```bash
$ docker run --rm -d --name=envoy -p 8080:8080 -v $(pwd)/proxy/config:/etc/envoy envoyproxy/envoy
```

There is no filter specified, so the reverse-proxy will do nothing, just send an "empty response" back.

```bash
$ curl -v http://localhost:8080

* TCP_NODELAY set
* Connected to localhost (127.0.0.1) port 8080 (#0)
...
* Empty reply from server
```

#### Static Response

As a next step we can add a filter to our listener, specifically an `HTTP` filter and make that respond with a _direct/static response text_. To achieve that we have to expand our `envoy.yaml` config file to add the `filters` and `routes` as shown below:

```yaml
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 8080
    filter_chains: 
      - filters:
        - name: envoy.filters.network.http_connection_manager
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
            stat_prefix: http_direct_response
            http_filters:
            - name: envoy.filters.http.router
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
            route_config:
              virtual_hosts:
              - name: direct_static_response
                domains: ["*"]
                routes:
                - match:
                    prefix: "/"
                  direct_response:
                    status: 200
                    body:
                      inline_string: "Hello from Envoy Proxy!"
```

Whilst it does look quite verbose, at a high level we are adding an `envoy.filters.network.http_connection_manager` in the `filter chain` and to that attaching an `envoy.filters.http.router` filter and a `route config` that describe how to route/handle an `HTTP` request based on the origin (domain/path). In this case we are specifying a _direct inline response string_ instead of an upstream `cluster`. _Later we shall change that to pint to our demo application as the cluster_. The [References](#references) section gives a number of links that I found useful to help us wrap our head around the configuration.

If we restart our `Envoy Docker` container, mounting the directory with the modified `envoy.yaml` (stop the running `Envoy container` then re-run the [command we used previously](#envoy-docker-run-mount)). And try to access that via a browser or `cURL` we should see our _hard coded_ string response from the _proxy_.

```bash
$ curl http://localhost:8080
Hello from Envoy Proxy!
```



<a id="#references">

### References:

</a> 

https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto

https://tetrate.io/blog/get-started-with-envoy-in-5-minutes/

https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples




