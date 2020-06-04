#!/bin/bash

awk '{ split($0,parts,"="); print parts[2]}' $(pwd)/deploy/version/version.properties