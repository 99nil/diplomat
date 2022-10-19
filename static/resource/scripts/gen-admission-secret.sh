#!/bin/bash

set -e

NAMESPACE=${NAMESPACE:-kubeedge}
SERVICE=${SERVICE:-"diplomat-admission"}
SECRET=${SECRET:-"diplomat-admission-secret"}
CERTDIR=${CERTDIR:-"/etc/kubeedge/admission-certs"}
ENABLE_CREATE_SECRET=${ENABLE_CREATE_SECRET:-true}
CN=${CN:-"${SERVICE}.${NAMESPACE}.svc"}
IP=${IP:-"127.0.0.1"}
if [[ ${IP} != "127.0.0.1" ]]; then
  echo "生成IP证书：${IP}"
  CN=${IP}
fi

if [[ ! -x "$(command -v openssl)" ]]; then
    echo "openssl not found"
    exit 1
fi

function createCerts() {
  echo "creating certs in dir ${CERTDIR} "

  cat <<EOF > ${CERTDIR}/csr.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICE}
DNS.2 = ${SERVICE}.${NAMESPACE}
DNS.3 = ${SERVICE}.${NAMESPACE}.svc
IP = ${IP}
EOF

  openssl genrsa -out ${CERTDIR}/ca.key 2048
  openssl req -x509 -days 3650 -new -nodes -key ${CERTDIR}/ca.key -subj "/CN=${CN}" -out ${CERTDIR}/ca.crt

  openssl genrsa -out ${CERTDIR}/server.key 2048
  openssl req -new -days 3650 -key ${CERTDIR}/server.key -subj "/CN=${CN}" -out ${CERTDIR}/server.csr -config ${CERTDIR}/csr.conf

  openssl x509 -req -days 3650 -in  ${CERTDIR}/server.csr -CA  ${CERTDIR}/ca.crt -CAkey  ${CERTDIR}/ca.key \
  -CAcreateserial -out  ${CERTDIR}/server.crt \
  -extensions v3_req -extfile  ${CERTDIR}/csr.conf
}

function createObjects() {
  # `ENABLE_CREATE_SECRET` should always be set to `true` unless it has been already created.
  if [[ "${ENABLE_CREATE_SECRET}" = true ]]; then
      kubectl get ns ${NAMESPACE} || kubectl create ns ${NAMESPACE}

      # create the secret with CA cert and server cert/key
      kubectl create secret generic ${SECRET} \
          --from-file=server.key=${CERTDIR}/server.key \
          --from-file=server.crt=${CERTDIR}/server.crt \
          --from-file=ca.crt=${CERTDIR}/ca.crt \
          -n ${NAMESPACE}
  fi
}

mkdir -p ${CERTDIR}
createCerts
createObjects
