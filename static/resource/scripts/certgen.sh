#!/usr/bin/env bash

set -o errexit

NAMESPACE=${NAMESPACE:-kubeedge}
readonly caPath=${CA_PATH:-/etc/diplomat/ca}
readonly caSubject=${CA_SUBJECT:-/C=CN/ST=Zhejiang/L=Hangzhou/O=KubeEdge/CN=kubeedge.io/CN=127.0.0.1}
readonly certPath=${CERT_PATH:-/etc/diplomat/certs}
readonly subject=${SUBJECT:-/C=CN/ST=Zhejiang/L=Hangzhou/O=KubeEdge/CN=kubeedge.io/CN=127.0.0.1}

genCA() {
  #2
    local IPs=(${@:1})
    echo $IPs
    local subj=${subject}
    if [ -n "$IPs" ]; then
        for ip in ${IPs[*]}; do
            subj="${subj}/CN=${ip}"
        done
    fi
    echo ${subj}
    openssl genrsa -des3 -out ${caPath}/rootCA.key -passout pass:kubeedge.io 4096
    openssl req -x509 -new -nodes -key ${caPath}/rootCA.key -sha256 -days 3650 \
    -subj ${subj} -passin pass:kubeedge.io -out ${caPath}/rootCA.crt
}

ensureCA() {
  #1
    local serverIPs=$1
    echo $serverIPs
    if [ ! -e ${caPath}/rootCA.key ] || [ ! -e ${caPath}/rootCA.crt ]; then
        genCA $serverIPs
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

genCsr() {
  #3
    local name=$1 IPs=(${@:2})
    local subj=${subject}
    if [ -n "$IPs" ]; then
        for ip in ${IPs[*]}; do
            subj="${subj}/CN=${ip}"
        done
    fi
    echo ${subj}
    openssl genrsa -out ${certPath}/${name}.key 2048
    openssl req -new -key ${certPath}/${name}.key -subj ${subj} -out ${certPath}/${name}.csr
}

genCert() {
  #4
    local name=$1 IPs=(${@:2})
    echo "IPS: " $IPs
    if  [ -z "$IPs" ] ;then
        openssl x509 -req -in ${certPath}/${name}.csr -CA ${caPath}/rootCA.crt -CAkey ${caPath}/rootCA.key \
        -CAcreateserial -passin pass:kubeedge.io -out ${certPath}/${name}.crt -days 365 -sha256
    else
        index=1
        SUBJECTALTNAME="subjectAltName = IP.1:127.0.0.1"
        for ip in ${IPs[*]}; do
            SUBJECTALTNAME="${SUBJECTALTNAME},"
            index=$(($index+1))
            SUBJECTALTNAME="${SUBJECTALTNAME}IP.${index}:${ip}"
        done
        echo $SUBJECTALTNAME > /tmp/server-extfile.cnf
        openssl x509 -req -in ${certPath}/${name}.csr -CA ${caPath}/rootCA.crt -CAkey ${caPath}/rootCA.key \
        -CAcreateserial -passin pass:kubeedge.io -out ${certPath}/${name}.crt -days 365 -sha256 -extfile /tmp/server-extfile.cnf
    fi
}

genCertAndKey() {
    while getopts ':i:h' opt; do
      case $opt in
          i) IPS=($OPTARG)
             ;;
          h) usage;;
          ?) usage;;
      esac
    done
    local name=$1 serverIPs=${IPS}
#    if [[ $serverIPs == *"Usage:"* ]];then
#        echo $serverIPs
#        exit 1
#    fi
    ensureFolder
    ensureCA $serverIPs
    genCsr $name $serverIPs
    genCert $name $serverIPs
}

stream() {
    ensureFolder
    local k8sCertPath=""
    local createK8sSecret=false
    readonly CLOUDCOREIPS=${CLOUDCOREIPS}
    readonly CLOUDCORE_DOMAINS=${CLOUDCORE_DOMAINS}
    readonly streamsubject=${SUBJECT:-/C=CN/ST=Zhejiang/L=Hangzhou/O=KubeEdge}
    readonly STREAM_KEY_FILE=${certPath}/stream.key
    readonly STREAM_CSR_FILE=${certPath}/stream.csr
    readonly STREAM_CRT_FILE=${certPath}/stream.crt

    readonly K8SCA_FILE=${k8sCertPath:-/etc/kubernetes/pki}/ca.crt
    readonly K8SCA_KEY_FILE=${k8sCertPath:-/etc/kubernetes/pki}/ca.key

    while getopts ":c:p:h" opt
    do
      case $opt in
        c)
        if [ "${OPTARG}" == "true" ]; then
          createK8sSecret=true
        fi
        ;;
        p)
        if [ -n "${OPTARG}" ]; then
          k8sCertPath=${OPTARG}
        fi
        ;;
        h)
        echo "Set -c true will create stream k8s secret after generated certificates.
It is necessary to set -p <YOUR_K8S_CERT_PATH>, which use to signed certificates."
        ;;
        ?)
        echo "unknown tag"
        exit 1;;
      esac
    done

    if [ -z "${k8sCertPath}" ]; then
      echo "You must set -p to specify k8s cert path."
      exit 1
    fi

    if [ -z "${CLOUDCOREIPS}" ] && [ -z "${CLOUDCORE_DOMAINS}" ]; then
        echo "You must set at least one of CLOUDCOREIPS or CLOUDCORE_DOMAINS Env.These environment
variables are set to specify the IP addresses or domains of all cloudcore, respectively."
        echo "If there are more than one IP or domain, you need to separate them with a space within a single env."
        exit 1
    fi

    index=1
    SUBJECTALTNAME="subjectAltName = IP.1:127.0.0.1"
    for ip in ${CLOUDCOREIPS}; do
        SUBJECTALTNAME="${SUBJECTALTNAME},"
        index=$(($index+1))
        SUBJECTALTNAME="${SUBJECTALTNAME}IP.${index}:${ip}"
    done
    index=1
    if [ ${CLOUDCORE_DOMAINS} -a "-n ${CLOUDCORE_DOMAINS}" ]; then
      for domain in ${CLOUDCORE_DOMAINS}; do
      SUBJECTALTNAME="${SUBJECTALTNAME},"
      index=$(($index+1))
      SUBJECTALTNAME="${SUBJECTALTNAME}DNS.${index}:${domain}"
      done
    fi

    cp ${K8SCA_FILE} ${caPath}/streamCA.crt
    echo $SUBJECTALTNAME > /tmp/server-extfile.cnf

    openssl genrsa -out ${STREAM_KEY_FILE}  2048
    openssl req -new -key ${STREAM_KEY_FILE} -subj ${streamsubject} -out ${STREAM_CSR_FILE}

    # verify
    openssl req -in ${STREAM_CSR_FILE} -noout -text
    openssl x509 -req -in ${STREAM_CSR_FILE} -CA ${K8SCA_FILE} -CAkey ${K8SCA_KEY_FILE} -CAcreateserial -out ${STREAM_CRT_FILE} -days 5000 -sha256 -extfile /tmp/server-extfile.cnf
    #verify
    openssl x509 -in ${STREAM_CRT_FILE} -text -noout

    if [ "${createK8sSecret}" = "true" ]; then
        kubectl get ns ${NAMESPACE} || kubectl create ns ${NAMESPACE}
        kubectl create secret generic cloudcore-stream \
            --from-file=stream.key=${certPath}/stream.key \
            --from-file=stream.crt=${certPath}/stream.crt \
            --from-file=streamCA.crt=${caPath}/streamCA.crt \
            -n ${NAMESPACE}
    fi
}

opts(){
  usage() { echo "Usage: $0 [-i] ip1,ip2,..."; exit; }
  local OPTIND
  while getopts ':i:h' opt; do
    case $opt in
        i) IFS=','
           IPS=($OPTARG)
           ;;
        h) usage;;
        ?) usage;;
    esac
  done
  echo ${IPS[*]}
}

edgesiteServer(){
    serverIPs="$(opts $*)"
    if [[ $serverIPs == *"Usage:"* ]];then
        echo $serverIPs
        exit 1
    fi
    local name=edgesite-server
    ensureFolder
    ensureCA
    genCsr $name
    genCert $name $serverIPs
    genCsr server
    genCert server $serverIPs
}

edgesiteAgent(){
    ensureFolder
    ensureCA
    local name=edgesite-agent
    genCsr $name
    genCert $name
}

# example: buildCloudcoreSecret -i 127.0.0.1,192.168.1.1
buildCloudcoreSecret() {
    ensureFolder
    while getopts ':i:h' opt; do
      case $opt in
          i) IPS=($OPTARG)
             ;;
          h) usage;;
          ?) usage;;
      esac
    done
    local name="edge"
    echo ${IPS}
    genCertAndKey ${name} -i ${IPS} > /dev/null 2>&1
    cat > ${certPath}/cloudcore-secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: cloudcore
  namespace: ${NAMESPACE}
  labels:
    app: diplomat-kubeedge
    kubeedge: cloudcore
    diplomat: diplomat
stringData:
  rootCA.crt: |
$(pr -T -o 4 ${caPath}/rootCA.crt)
  rootCA.key: |
$(pr -T -o 4 ${caPath}/rootCA.key)
  edge.crt: |
$(pr -T -o 4 ${certPath}/${name}.crt)
  edge.key: |
$(pr -T -o 4 ${certPath}/${name}.key)

EOF
    kubectl get ns ${NAMESPACE} || kubectl create ns ${NAMESPACE}
    kubectl apply -f ${certPath}/cloudcore-secret.yaml
}

$@
