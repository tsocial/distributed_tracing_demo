#!/usr/bin/env bash
go run github.com/tsocial/distributed_tracing_demo localhost:3000 &
go run github.com/tsocial/distributed_tracing_demo localhost:4000
