#!/bin/bash

nats -s localhost:4211 server check connection
nats -s localhost:4212 server check connection
nats -s localhost:4213 server check connection
nats -s localhost:4221 server check connection
nats -s localhost:4231 server check connection