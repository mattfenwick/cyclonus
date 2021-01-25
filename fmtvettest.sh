#!/bin/bash

set -xv

make fmt

make vet

make test
