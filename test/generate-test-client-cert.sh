#!/usr/bin/env bash

cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=client client.json | cfssljson -bare client-1
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=client -hostname=spiffe://example.com/users/bad-user-789 client.json | cfssljson -bare client-2
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=client -hostname=spiffe://example.com/users/good-user-456 client.json | cfssljson -bare client-3