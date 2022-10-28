module github.com/99nil/diplomat

go 1.18

replace github.com/99nil/dsync => ./staging/src/github.com/99nil/dsync

require (
	github.com/99nil/dsync v0.0.0-00010101000000-000000000000
	github.com/99nil/gopkg v0.0.0-20220607055250-e19b23d7661a
	github.com/AlecAivazis/survey/v2 v2.3.6
	github.com/go-chi/chi v1.5.4
	github.com/kubeedge/kubeedge v1.11.2
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/r3labs/sse/v2 v2.8.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.12.0
	github.com/zc2638/aide v0.0.1
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.22.6
	k8s.io/apimachinery v0.22.6
	k8s.io/client-go v0.22.6
	k8s.io/klog/v2 v2.60.1
	k8s.io/kubernetes v1.22.6
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/containerd v1.5.10 // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/badger/v3 v3.2103.2 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.12+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/flatbuffers v1.12.1 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.12.3 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.0.3 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/net v0.0.0-20220617184016-355a448f1bc9 // indirect
	golang.org/x/oauth2 v0.0.0-20220608161450-d0670ef3b1eb // indirect
	golang.org/x/sys v0.0.0-20220615213510-4f61da869c0c // indirect
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd // indirect
	google.golang.org/grpc v1.46.2 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiserver v0.22.6 // indirect
	k8s.io/cluster-bootstrap v0.22.6 // indirect
	k8s.io/component-base v0.22.6 // indirect
	k8s.io/component-helpers v0.22.6 // indirect
	k8s.io/cri-api v0.22.6 // indirect
	k8s.io/kube-openapi v0.0.0-20220328201542-3ee0da9b0b42 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

replace (
	github.com/kubeedge/beehive v0.0.0 => github.com/kubeedge/beehive v1.7.0
	github.com/kubeedge/viaduct v0.0.0 => github.com/kubeedge/viaduct v1.7.0
	k8s.io/api v0.0.0 => k8s.io/api v0.22.6
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.22.6
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.22.6
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.22.6
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.22.6
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.22.6 // indirect
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.22.6 // indirect
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.22.6
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.15.8-beta.1
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.22.6
	k8s.io/component-helpers v0.0.0 => k8s.io/component-helpers v0.22.6
	k8s.io/controller-manager v0.0.0 => k8s.io/controller-manager v0.22.6
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.22.6
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.22.6
	k8s.io/gengo v0.0.0 => k8s.io/gengo v0.22.6 // indirect
	k8s.io/heapster => k8s.io/heapster v1.2.0-beta.1 // indirect
	k8s.io/klog => k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.22.6
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.22.6
	k8s.io/kube-openapi v0.0.0 => k8s.io/kube-openapi v0.22.6 // indirect
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.22.6
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.22.6
	k8s.io/kubectl => k8s.io/kubectl v0.22.6
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.22.6
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.22.6
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.22.6
	k8s.io/mount-utils v0.0.0 => k8s.io/mount-utils v0.22.6
	k8s.io/node-api v0.0.0 => k8s.io/node-api v0.22.6 // indirect
	k8s.io/pod-security-admission v0.0.0 => k8s.io/pod-security-admission v0.22.6
	k8s.io/repo-infra v0.0.0 => k8s.io/repo-infra v0.22.6 // indirect
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.22.6
	k8s.io/utils v0.0.0 => k8s.io/utils v0.22.6
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.0 => sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.27
)
