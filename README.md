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
# 1. Get started
pipes init --servers <ip1,ip2,ip3>

# 2. Build micro-services
pipes build service_1 service_3 service_3
# you can add binary or executable
# >> Docker images will be build based on each executable

# 3. Run the worflow of micro-services using the classic '|'
pipes run "service_1 <some_arg> | service_2 | service_3"
# >> Containers are spawned accross your cluster
# >> pipes return the result of the workflow.

# 3.bis. In daemon mode an API is automatically generated
pipes run -d "service_1 | service_2 | service_3"

# 4. You can query the API through the CLI
pipes query "some_data"

# 4.bis. Or through the API of the worflow
curl -d query="some_data" http://<addr>/
>> Job ID: <id>
curl http://<addr>/jobs/<id>/

# 4. List your worklows
pipes ps -a
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
