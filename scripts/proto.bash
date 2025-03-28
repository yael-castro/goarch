#!/bin/bash

pwd
protoc --proto_path=./pkg/pb --go_out=. purchase.proto header.proto