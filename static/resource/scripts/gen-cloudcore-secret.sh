#!/usr/bin/env bash

set -o errexit

NAMESPACE=${NAMESPACE:-kubeedge}
SECRET=${SECRET:-"cloudcore"}
ENABLE_CREATE_SECRET=${ENABLE_CREATE_SECRET:-true}
readonly caPath=${CA_PATH:-/etc/diplomat/ca}
readonly certPath=${CERT_PATH:-/etc/diplomat/certs}
readonly subject=${SUBJECT:-/C=CN/ST=Zhejiang/L=Hangzhou/O=KubeEdge/CN=kubeedge.io}
CN=""
# TODO 支持多IP
IP=${IP:-"127.0.0.1"}
if [[ ${IP} != "127.0.0.1" ]]; then
  echo "生成IP证书：${IP}"
  CN=${IP}
fi

function createCerts() {
  echo "creating certs in dir ${CERTDIR} "
  cat <<EOF > ${certPath}/csr.conf
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
IP = 127.0.0.1
IP = ${IP}
EOF

  openssl genrsa -out ${caPath}/rootCA.key 2048
  openssl req -x509 -days 3650 -new -nodes -key ${caPath}/rootCA.key -subj "/CN=${CN}" -out ${caPath}/rootCA.crt

  openssl genrsa -out ${certPath}/edge.key 2048
  openssl req -new -days 3650 -key ${certPath}/edge.key -subj "/CN=${CN}" -out ${certPath}/edge.csr -config ${certPath}/csr.conf

  openssl x509 -req -days 3650 -in  ${certPath}/edge.csr -CA  ${caPath}/rootCA.crt -CAkey  ${caPath}/rootCA.key \
  -CAcreateserial -out  ${certPath}/edge.crt \
  -extensions v3_req -extfile  ${certPath}/csr.conf
}

function createObjects() {
  # `ENABLE_CREATE_SECRET` should always be set to `true` unless it has been already created.
  if [[ "${ENABLE_CREATE_SECRET}" = true ]]; then
      kubectl get ns ${NAMESPACE} || kubectl create ns ${NAMESPACE}

      # create the secret with CA cert and server cert/key
      kubectl create secret generic ${SECRET} \
          --from-file=edge.key=${certPath}/edge.key \
          --from-file=edge.crt=${certPath}/edge.crt \
          --from-file=rootCA.crt=${caPath}/rootCA.crt \
          --from-file=rootCA.key=${caPath}/rootCA.key \
          -n ${NAMESPACE}
  fi
}

ensureFolder() {
    if [ ! -d ${caPath} ]; then
        mkdir -p ${caPath}
    fi
    if [ ! -d ${certPath} ]; then
        mkdir -p ${certPath}
    fi
}

ensureFolder
createCerts
createObjects
