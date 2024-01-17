#!/bin/bash

################################################################################
# Parameters
################################################################################

ca_name='Tonysoft Root CA'
cs_name='server.tonysoft.com'
key_length=4096
exp_days=365

################################################################################
# Certificate Authority (CA)
################################################################################

# Generate the private key for the CA:
openssl genrsa -out test-ca.key $key_length

# Generate the CSR for the CA:
openssl req -new -key test-ca.key -out test-ca.csr -sha256 -subj "/CN=$ca_name"

# Create the config file for the CA:
cat > test-ca.cnf <<EOL
[root_ca]
basicConstraints = critical,CA:TRUE,pathlen:1
keyUsage = critical, nonRepudiation, cRLSign, keyCertSign
subjectKeyIdentifier=hash
EOL

# Generate the certificate for the CA:
openssl x509 -req -sha256 -days $exp_days -extensions root_ca -in test-ca.csr -signkey test-ca.key -extfile test-ca.cnf -out test-ca.crt

################################################################################
# Compute Server (CS)
################################################################################

# Generate the private key for the CS:
openssl genrsa -out test-cs.key $key_length

# Generate the CSR for the CS:
openssl req -new -key test-cs.key -out test-cs.csr -sha256 -subj "/CN=$cs_name"

# Create the config file for the CS:
cat > test-cs.cnf <<EOL
[server]
authorityKeyIdentifier=keyid,issuer
basicConstraints = critical,CA:FALSE
extendedKeyUsage=serverAuth
keyUsage = keyEncipherment, dataEncipherment, digitalSignature
subjectAltName = DNS:$cs_name, DNS:localhost, IP:127.0.0.1
subjectKeyIdentifier=hash
EOL

# Generate the certificate for the CS:
openssl x509 -req -sha256 -days $exp_days -extfile test-cs.cnf -extensions server -CA test-ca.crt -CAkey test-ca.key -CAcreateserial -in test-cs.csr -out test-cs.crt

################################################################################
# Cleanup
################################################################################

rm test-ca.cnf
rm test-cs.cnf
rm test-ca.csr
rm test-cs.csr
