# NATS Messenger
NATS Messenger is a lightweight Go-based messaging client built on NATS. It provides tools for sending messages, streaming files, and testing distributed setups with multiple NATS datacenters via Docker Compose.

## 🚀 Docker Commands
### Build Docker image (standard)
```
make docker-build
```
### Builds a Docker image using the standard deploy/dockerfile:
```
mrtdeh/nats-client
```
### Build Docker image (local / portable binary)
```
make docker-build-local
```
- Uses a portable, statically compiled binary for Docker.
- Recommended for local development or testing in lightweight containers.
### Start multi-datacenter cluster (DC1 + DC2 + DC3)
```
make docker-up
```
- Spins up all three datacenters (DC1, DC2, DC3) using their respective docker-compose files.

- Forces container recreation and removes orphans.

- Builds the local image before starting DC1.

### Stop and remove all datacenters
```
make docker-down
```
- Stops all NATS datacenters.

- Removes all containers created by the Compose files.

## ⚙️ How it Works
### Attach to Clients
After starting the multi-datacenter cluster, You can attach to any of chat clients in all DCs. in below we attach to app1 container from the DC1 zone (or whichever container you want) to use the NATS Messenger CLI:
```
docker attach dc1-app1
```
Once inside the container, the CLI will be running. The commands available are:

### Command Description
**goto** : Change the current messaging path.
**send** : Publish a chat message to the current path.
**sendfile** : Stream a file to the current path.
**nodes** : List all available node's path in the all datacenters.
**dcs** : List all datacenters in the cluster.
**exit / quit** : Exit the CLI.

### Usage
**List of node's path with health(ok/nok) :**
```
root> nodes
```
outputed paths:
```
nodes/dc1/app3 : ok
nodes/dc1/app1 : ok
nodes/dc1/app2 : ok
nodes/dc2/app3 : ok
nodes/dc2/app2 : ok
nodes/dc2/app1 : ok
nodes/dc3/app3 : ok
nodes/dc3/app1 : ok
nodes/dc3/app2 : ok
```

**Change chat space to specific node path :**
```
root> goto nodes/dc3/app1
```
**Send text and file to current path :**
```
nodes/dc3/app1> send "Hello from DC1!"
nodes/dc3/app1> sendfile /tmp/test.txt
send new file name=/tmp/test.txt size=21531 id=877dc70a-315e-4d22-8188-de72b23f5748
send file finish : 21531/21531
```
**You can sure by see logs of right client :**
```
docker logs -f dc3-app1
```
## 🖥️ Test NATS Servers
You can use the nats CLI to connect to your NATS servers:
```bash
# Connect as system user for system commands like below:
nats -s sys:sys@localhost:4221 server list
# This command lists all NATS servers in the cluster using system credentials.

# Connect as application user for general commands like below:
nats -s localhost:4221 stream list
# This command lists all streams available for the application user.
```
## 🏗 Purpose
This setup allows you to:

- Run a multi-datacenter NATS environment locally for testing

- Validate messaging, file transfer, and Key-Value store functionality

- Debug distributed systems with a lightweight CLI client

## 📌 Notes
- Docker Compose v2+ is required

- Images are tagged as: mrtdeh/nats-client

- DC-specific Compose files are located in deploy/