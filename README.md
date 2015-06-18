# PIPES

**Distribute your micro-services in the Cloud as a Unix pipe.**

*Built in Go with Docker, etcd, swarm and the Wamp protocol.*

We all love micro-services.
Mostly because they follow the Unix philisophy: one simple tool for each task.

In Unix we can easily chain those tools by the standard input/output, through the pipe.

But in a cloud environment we can't. 
We have either to use specific RPC features or build for each service a dedicated API. 
Even so we still have to manage all the orchestration and worflow process.
This makes the micro-services less agnostic, difficult to test and difficult to deploy.

What if we could just do the same in the Cloud as in our terminal ?

```sh
service_1 | service_2 | service_3
```

Enter **Pipes**, a micro-services distributed and fault-tolerant tool :

```sh
# 1. Start a project
pipes start myproject

# 2. Add some micro-services
pipes add service_1 service_3 service_3
# you can add binary, script, Dockerfile, buildpack ...

# 3. Define the worflow of the micro-services using the classic '|'
pipes link "service_1 | service_2 | service_3"

# 4. Deploy and run the whole worflow across your servers
pipes run

# 5. Query your project through the CLI
pipes query "some_data"
# or through the API automatically generated for you project.
curl -d "some_data" http://my_project/api/
```

## Features

**Easy to develop, test and deploy**

*Pipes* is agnostic, you can <u>use and mix</u> micro-services in whatever language you use: Python, Ruby, Node, Java, Go, C, Perl, Bash, etc.

*Pipes* is not intrusive, you don't have to add specific functions inside your code.

In all cases your micro-services will behave in the exact same way:

- in development, using the standard Unix `|`
- in test, using *Pipes* locally or on tests servers
- in production, using *Pipes* on your servers.

**Distributed**

The micro-services are deployed and distributed across your servers.

You can start with only 1 server (or even locally) and add or remove more servers afterwards, *Pipes* will manage the rebalancing of the micro-services.

*Pipes* handle the communication between each micro-service, no matter the server in which they are deployed.

**Isolated**

Each micro-service is built inside a Docker container.

Docker containers allow micro-services to be isolated from each other and from the server, and ensure the reproductability of your worflow.

**Scalable and loadbalanced**

You can scale up or down the number of replicas for each micro-service, or for the whole project, *Pipes* will spawn or kill the replicas accordingly.

*Pipes* loadbalance the workload across all the replicas of one micro-service.

You can also auto-scale: *Pipes* will monitor each micro-service and scale them up or down if needed.

**Fault-tolerant**

If one micro-service crashes, *Pipes* will restart it automatically.

If one server is down, *Pipes* will rebalance its micro-services to other healthy servers in the cluster.

## Getting started

*Pipes* is easy to use and has no dependancies.
It's available for Linux, MacOSX and Windows.

**1. Download**

[Link of the binaries].

**2.1. Local usage**

You can test *Pipes* locally.

You just need Docker in your computer (or boot2docker on MacOSX and Windows).

```sh
pipes init
# It will look for $DOCKER_HOST environnment variable (boot2docker)
# or use the local standard host ie. unix:///var/run/docker.sock
```

**2.2. Production usage**

Docker must be running in each of your server, using the standard tcp port:

```sh
docker -d -H tcp://0.0.0.0:2375

# Or for TLS support
# docker -d -H tcp://0.0.0.0:2376
```

Then in your computer (or whatever machine which can acces your servers):

```sh
pipes init <ip1>:<port1> <ip2>:<port2> <ip3>:<port3>

# For example
pipes init 12.67.23.46:2375 12.89.04.48:2375 12.89.12.67:2375
```

## Architecture

*Pipes* is written in Go and built on innovative technologies :

- Docker for the container engine
- Wamp protocol for the communication between micro-services
- etcd for the shared configuration
- Docker Swarm for the orchestration

**Docker**

Who needs to present [Docker](https://www.docker.com/) anymore ?

*Pipes* use Docker to build each micro-service inside an isolated container.

All internal components of *Pipes* are also installed in Docker containers, and monitored by the *Pipes* orchestration.

*Pipes* run <u>everything</u> on Docker.

**Wamp protocol**

[Wamp](http://wamp.ws/) is a standard protocol implementing both RPC and PubSub in top of Websockets.

*Pipes* use Wampace, a router and client implemtation of the Wamp protocol in Go.

For each micro-service a wamp client is in charge:

- in one side of the communication with the router
- in the other side of calling the underlying micro-service

The Wamp router dispatches the communication between micro-services, through the clients.

The API automatically generated for each project is also a wamp client which can:

1. Initiate the worflow when receiving a request
2. Retrieve the result to respond

**etcd**

[etcd](https://coreos.com/etcd/) is a fault-tolerant key/value store.

*Pipes* use etcd to store and share the configuration between all the internal components.

Several etcd instances are installed across the servers.
etcd instances use a multi-master model with automatic master election (Raft consensus algorithm) and replication of data.

**Docker swarm**

[Docker swarm](https://docs.docker.com/swarm/) is an orchestration tool to manage Docker containers.

*Pipes* use swarm to: 

- watch the health of each server
- schedule the deployment of containers across the servers
- manage the lifecycle of containers (scaling, rebalancing)

A swarm agent is installed in each server to monitor its health.

Several swarm manager instances are installed across the servers. 
As etcd, Swarm manager instances use a multi-master model with automatic master election (Raft consensus algorithm).
