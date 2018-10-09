#!/bin/bash

# Download tools
go get -v github.com/golang/mock/gomock
go get -v github.com/Masterminds/glide

# Install mockgen
go install github.com/golang/mock/mockgen
