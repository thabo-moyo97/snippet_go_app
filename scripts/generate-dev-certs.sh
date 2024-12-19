#!/bin/bash
set -e

# Create tls directory if it doesn't exist
mkdir -p tls

# Generate CA private key
openssl genrsa -out tls/ca.key 2048

# Generate CA certificate
openssl req -x509 -new -nodes -key tls/ca.key -sha256 -days 365 -out tls/ca.pem \
    -subj "/C=GB/ST=London/L=London/O=Development/CN=Local Development CA"

# Generate server private key
openssl genrsa -out tls/key.pem 2048

# Create certificate signing request configuration
cat > tls/csr.conf << EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[dn]
C = GB
ST = London
L = London
O = Development
CN = localhost

[req_ext]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
EOF

# Generate certificate signing request
openssl req -new -key tls/key.pem -out tls/csr.pem -config tls/csr.conf

# Generate certificate
openssl x509 -req -in tls/csr.pem -CA tls/ca.pem -CAkey tls/ca.key \
    -CAcreateserial -out tls/cert.pem -days 365 \
    -sha256 -extensions req_ext -extfile tls/csr.conf

# Clean up temporary files
rm tls/csr.conf tls/csr.pem

echo "Development certificates generated successfully" 