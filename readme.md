
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

Simple backend app in `Go`.
Shows a simple web page with the Name of the app, and the Host it is running on. By default it runs on PORT 8080, but that can be customised via Env var.
The app name, background and foreground colours can be customised via Env vars.

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

### Reverse Proxy Config 

`Envoy` configuration is very powerful in terms of all the options it gives us, however this also makes it quite complex, and difficult to find our way around it. One good way to understand it is to start with a bare-bones _listener_ that does nothing and then start layering functionality bit-by-bit and see what changes we need to do to achieve that. By the end of our journey, we will be more at home with the _configuration_ and more confident in playing around with it.

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
Now Run `Envoy proxy` again with Docker (expose `Port 8080` as that is the listening port we specified in the config above) using the <a id="#envoy-docker-run-mount">following command</a>.

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

#### Proxy to Backend App

The next logical step would be to _route_ the traffic to some _upstream_(backend application). In our case we shall use our `demoapp` application that we wrote. We shall run our `Envoy` proxy and the `demoapp`as `Docker` containers and route traffic the _proxy_ to the _app_. For this to work in `Docker` the _containers_ need a way to discover each other using their '_container name_' (else we will have to keep modifying the `envoy.yaml` config with the dynamically assigned `IP` of the _backend app container_). The simplest way to do this is as follows:

- Create a user-defined `Docker` _network_
- When launching the _containers_, specify a `--name` for the _container_ and attach them to the user-defined _network_ (via the `--net` option)
- Now, container on that _network_ can address each other using the `--anme` specified

```bash
# create docker network
$ docker network create nw_demo_apps

# launch demoapp as app1 into that network
$ docker run --rm --name=app1 -p 1111:8080 --net=nw_demo_apps demoapp
# Note: we have published to localhost:1111, this is just for testing, not needed for the proxy
```

Next we modify our `envoy.yaml` to add a _cluster_ (which points to the above `app1`) specify that _cluster_ as the _upstream_ for the _route config_ in our _listener_.

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
                  route:
                    cluster: app-one
  clusters:
  - name: app-one
    connect_timeout: 3s
    type: strict_dns
    load_assignment:
      cluster_name: app-one
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: app1
                port_value: 8080
```

Note: that in the _cluster_ section the `type: strict_dns` is required if we want it to resolve the DNS name `app1` to its `IP` (I learned that the hard way).

Now we launch the _proxy_ again using the same command we used previously but this time add it to the `nw_dwmo_apps` `Docker` _network_.

```bash
$ docker run --rm -d --name=envoy -p 8080:8080 -v $(pwd)/proxy/config:/etc/envoy --net=nw_demo_apps envoyproxy/envoy
```

Test it out by going to `http://localhost:8080` in your web browser, you should see a web page (served through the _proxy_) showing the name of the application as `Demo App - 1` and the _host_ it is running on (a `Docker` _container id_ in this case). If we access the app container directly using `http://localhost:1111` we should see the exact sage page.

> > Insert image here

##### Docker Compose

Since we will be repeatedly creating & destroying containers to try things out, it makes sense to use `docker-compose` to define our _setup_(_infrastructure_) rather typing out these long _shell commands_ all the time. I have a `setup` directory, with a `docker-compose.yaml` where we shall define our `Docker` setup. The `docker-compose` manifest for our simple setup above, with a single _app_ container and a _proxy_ container attached to a _user-defined network_ looks as shown. It is pretty self explanatory if you are familiar with `docker-compose`.

```yaml
version: "3"
services:
  webapp:
    image: demoapp
    ports:
    - "1111:8080"
    container_name: app1
    networks:
    - envoy_demo_nw
  proxy:
    image: envoyproxy/envoy
    ports:
    - "8080:8080"
    container_name: envoy
    networks:
    - envoy_demo_nw
    volumes: 
    - ../proxy/config:/etc/envoy
networks: 
  envoy_demo_nw:
    ipam:
      driver: default
```

We can execute `$ docker-compose up` (optionally use `-d` flag if we want it to launch it in the background) to bring up all our containers and associated resources and configuration. And, `$ docker-compose down` to bring it down. Note, that we have to run that command from the directory that has the `docker-compose.yaml` file.

From hereon we shall keep modify work with this `docker-compose.yaml` whenever we need to add or change containers to our setup.

#### Load-balancing

Now we shall add two instances of the an app (let's name them `app1-1` and `app1-2`) and modify our `envoy.yaml` config to simulate how multiple requests gets load-balanced between the instances. To achieve this, first we modify our `docker-compose.yaml` to add an additional `service` with the same _image_, so that we have two _containers_ running for the same _app_. Note that we changed the `ports` directive to `expose` , since we do not need to _bind_ them to _host_ anymore, we just need them exposed from the _container_ so that they can be accessed from the `Docker network` they are on.

```yaml
version: "3"
services:
  webapp-1:
    image: demoapp
    expose:
    - "8080"
    container_name: app1-1
    networks:
    - envoy_demo_nw
  webapp-2:
    image: demoapp
    expose:
    - "8080"
    container_name: app1-2
    networks:
    - envoy_demo_nw
  proxy:
    image: envoyproxy/envoy
    ports:
    - "8080:8080"
    container_name: envoy
    networks:
    - envoy_demo_nw
    volumes: 
    - ../proxy/config:/etc/envoy
networks: 
  envoy_demo_nw:
    ipam:
      driver: default
```

Modify the `envoy.yaml`configuration to add an `endpoint` under the `lb_endpoints` section in the `cluster`.

```yaml
static_resources:
  listeners:
 	# not shown here for brevity...
  clusters:
  - name: app-one
    connect_timeout: 3s
    type: strict_dns
    load_assignment:
      cluster_name: app-one
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: app1-1
                port_value: 8080
        - endpoint:
            address:
              socket_address:
                address: app1-2
                port_value: 8080

```

With that simple change `Envoy` will load-balance multiple requests(that match the `filter`) between those two`endpoints` in the cluster. Now if we navigate to `http://localhost:8080` in a browser and keep refreshing the page, we should see our familiar web-page (for `Demo App-1`), but the `Host` name should keep alternating between two `Docker` _container Ids_.

> >  Insert image 1 & 2 LB

Of course this is a very simple demonstration of that capability using the default `round-robin` algorithm (we can control it using the `lb_policy` directive if needed). `Envoy` can do much more complex load-balancing. The [References](#references) section has a link for details on the various types of load-balancing it can do.

#### Routing to Multiple Backends 

So far we have all requests send to the same _backend_ (or _cluster_). Now we shall setup two different apps (`app1` and `app2`) as _backends_ an see how we can _route_ traffic to each. `Envoy` provides so many different _match_ options such as `path`, `prefix`, `headers`, `query_params` etc. to decide how to route traffic ([References](#references) for detailed documentation). In our example we shall keep it simple and just _route_ based on the `prefix` in the `URL`.

To simulate two different apps, we can launch our `demoapp` with some `env` _variables_ to change the _displayed name_, and its _background colour_. We provide these in our modified `docker-compose.yaml`.

```yaml
version: "3"
services:
  webapp-1:
    image: demoapp
    expose:
    - "8080"
    container_name: app1
    networks:
    - envoy_demo_nw
  webapp-2:
    image: demoapp
    expose:
    - "8080"
    container_name: app2
    environment: 
    - APPNAME=Demo App-2
    - BGCOLOR=#defc79
    networks:
    - envoy_demo_nw
  proxy:
    image: envoyproxy/envoy
    ports:
    - "8080:8080"
    container_name: envoy
    networks:
    - envoy_demo_nw
    volumes: 
    - ../proxy/config:/etc/envoy
networks: 
  envoy_demo_nw:
    ipam:
      driver: default
```

It is almost same as before, we simply changed the `container_name` to `app1` and `app2` and add a couple of `env` variables to `app2` to make it look different (_lime_ background colour) and display name as `Demo App-2`.

Then we modify the `envoy.yaml` config to add `app2` as a different _cluster_, then specify a couple of `match` `prefix` directives to _route_ the traffic to the respective cluster.

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
              - name: route_app1_app2
                domains: ["*"]
                routes:
                - match:
                    prefix: "/app1"
                  route:
                    cluster: app-one
                - match:
                    prefix: "/app2"
                  route:
                    cluster: app-two
  clusters:
  - name: app-one
    connect_timeout: 3s
    type: strict_dns
    load_assignment:
      cluster_name: app-one
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: app1
                port_value: 8080
  - name: app-two
    connect_timeout: 3s
    type: strict_dns
    load_assignment:
      cluster_name: app-two
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: app2
                port_value: 8080
```

This should be all that is needed to route traffic to _cluster_ (`app-one` or `app-two`) based on the path _prefix_. And sort of works, if we launch our `docker-compose` now and try to browse `http://localhost:8080/app1` or `http://localhost:8080/app2` we will see our web pages, but without some of the _style_ settings and _favicon_. 

> > > screen shots

Similarly is we try to access the `http://localhost:8080/app1/api/info` endpoint we will end up the `index.html` page !!!

The issue is that when `Envoy` does a `prefix` `match` for a _route_ it does not seem to _forward_ _"nested"_ paths along (such as `app1/assets/` or `app1/api`). Therefore these paths have to be explicitly specified pointing to the right _cluster_. This can seem repetitive. In our example to make this work we would have to add extra entries in the `routes` section for each _path_. That section of the `envoy.yaml` will now look like.

```  yaml
...
			route_config:
              virtual_hosts:
              - name: route_app1_app2
                domains: ["*"]
                routes:
                - match:
                    prefix: "/app1/assets"
                  route:
                    cluster: app-one
                    prefix_rewrite: "/assets"
                - match:
                    prefix: "/app1/api"
                  route:
                    cluster: app-one
                    prefix_rewrite: "/api"
                - match:
                    prefix: "/app1"
                  route:
                    cluster: app-one
                    prefix_rewrite: "/"
                - match:
                    prefix: "/app2/assets"
                  route:
                    cluster: app-two
                    prefix_rewrite: "/assets"
                - match:
                    prefix: "/app2/api"
                  route:
                    cluster: app-two
                    prefix_rewrite: "/api"
                - match:
                    prefix: "/app2"
                  route:
                    cluster: app-two
                    prefix_rewrite: "/"
...
```

Also note that we use the `prefix_rewrite` directive to remove the _"prefix"_ part from the request sent to the _upstream_ app. The _prefix_ part is only useful within the context of the _downstream_ and the _proxy_ to determine the _route_. The backend app (`demoapp` in our case) will not know how to handle a request with the _prefix_ path (`app1` or `app2` etc.).

With this modification to our `envoy.yaml` configuration we should be able to successfully route to `app1` or `app2` using a prefix and be able to see the full working application.

> > > insert screen shot

##### Regex Match & Rewrite

To me that seems like a lot of repetitive entries, and also seems to need knowledge of the upstream application paths to make it work. Unfortunately I could not find any _"wildcard"_ matching mechanism as far as I know with `Envoy`. 

The next best thing however seems to be to use its _Regex_ matching and rewriting capability. So we shall now modify the `envoy.yaml` config using `safe_regex` _match_ and `regex_rewrite`. The modified `route` section in our `envoy.yaml` with the _Regex_ will look like:

```yaml

```

Now if we test out our _URLs_ `http://localhost:8080/app1` or `app2` in the browser we should see the same result we saw previously. And yes, even though we can avoid the repetition, and reduce the lines of configuration, I think this can get quite complicated and error prone. Error's in _Regexes_ can be infamously hard to test and debug. As a case in point here is a link (https://blog.cloudflare.com/cloudflare-outage/) to a `CloudFlare` outage caused by a badly behaving _Regex_. So be extra cautious when taking this approach. 



#### Dynamic Configuration



<a id="#references">

### References:

</a> 

https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto

https://tetrate.io/blog/get-started-with-envoy-in-5-minutes/

https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples

https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancers#arch-overview-load-balancing-types

https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-routematch




