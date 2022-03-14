#!/usr/bin/env bash

cfssl genkey -initca root-csr.json | cfssljson -bare ca