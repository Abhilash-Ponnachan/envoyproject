
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

### chechk browser http://localhost:10000
>>> proxied to "Envoy home page"

### Commands
```bash
# exec into 'evnoy' container and check /etc/envoy
$ docker exec -it envoy sh 
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

### now again check browser http://localhost:1000
>>> proxied to "Google Search home page"



