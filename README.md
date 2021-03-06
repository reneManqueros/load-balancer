# Load Balancer
#### Simple load balancer written in Go

## Features
- TCP/UDP support
- Hot swap of backends
- Low resource footprint
- Telnet management console

## Installation
Download:
```shell
git clone https://github.com/reneManqueros/load-balancer
```

Compile:
```shell
cd load-balancer && go build .
````

## Usage

Execute as load balancer:
```shell
./loadbalancer balance
```

Add a backend:
```shell
./loadbalancer add 127.0.0.1:8080
```

Add a backend via telnet - Send a "+" and the address:
```shell
telnet 127.0.0.1 33333
+127.0.0.1:8080
```

Remove a backend:
```shell
./loadbalancer add 127.0.0.1:8080
```

Remove a backend via telnet - Send a "-" and the address:
```shell
telnet 127.0.0.1 33333
-127.0.0.1:8080
```

## Parameters 

### When running as load balancer
Listen for TCP (default)
```shell
./loadbalancer balance network=tcp
```

Listen for UDP
```shell
./loadbalancer balance network=tcp
```

Change listen Port (default: 8081)
```shell
./loadbalancer balance address=:8088
```

Disable management console
```shell
./loadbalancer balance management=""
```

Config file location (default: ./backends.yml)
```shell
./loadbalancer balance config="/etc/loadbalancer/backends.yml"
```

Enable verbose mode (default: false)
```shell
./loadbalancer balance verbose=true
```

Backend timeout in milliseconds (default: 0)
```shell
./loadbalancer balance timeout=300
```

### When running as load balancer or when adding/removing backends
Change management console host/port
```shell
./loadbalancer balance management="127.0.0.1:12345"
```
 

