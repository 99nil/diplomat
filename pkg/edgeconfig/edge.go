// Package edgeconfig
// Created by zc on 2021/12/1.
package edgeconfig

import (
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/99nil/diplomat/global/constants"
	"github.com/99nil/diplomat/pkg/util"

	coreV1 "k8s.io/api/core/v1"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDefaultEdgeCoreConfig returns a full EdgeCoreConfig object
func NewDefaultEdgeCoreConfig(rootDir, resourceDir string) *EdgeCore {
	hostnameOverride, err := os.Hostname()
	if err != nil {
		hostnameOverride = "default-edge-node"
	}

	msgs := apimachineryvalidation.NameIsDNSSubdomain(hostnameOverride, false)
	if len(msgs) > 0 {
		hostnameOverride = "default-edge-node"
	}

	localIP, _ := util.GetLocalIP(hostnameOverride)
	cfg := &EdgeCore{
		TypeMeta: metaV1.TypeMeta{
			Kind:       Kind,
			APIVersion: path.Join(GroupName, APIVersion),
		},
		DataBase: &DataBase{
			DriverName: constants.DataBaseDriverName,
			AliasName:  constants.DataBaseAliasName,
			DataSource: filepath.Join(resourceDir, constants.DataBaseDataSource),
		},
		Modules: &EdgeModules{
			Edged: &Edged{
				Enable:                      true,
				Labels:                      map[string]string{},
				Annotations:                 map[string]string{},
				NodeStatusUpdateFrequency:   10,
				RuntimeType:                 "docker",
				DockerAddress:               "unix:///var/run/docker.sock",
				RemoteRuntimeEndpoint:       "unix:///var/run/dockershim.sock",
				RemoteImageEndpoint:         "unix:///var/run/dockershim.sock",
				NodeIP:                      localIP,
				ClusterDNS:                  "",
				ClusterDomain:               "",
				ConcurrentConsumers:         5,
				EdgedMemoryCapacity:         7852396000,
				PodSandboxImage:             "registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.2",
				ImagePullProgressDeadline:   60,
				RuntimeRequestTimeout:       2,
				HostnameOverride:            hostnameOverride,
				RegisterNodeNamespace:       "default",
				CustomInterfaceName:         "",
				RegisterNode:                true,
				DevicePluginEnabled:         true,
				GPUPluginEnabled:            true,
				ImageGCHighThreshold:        80,
				ImageGCLowThreshold:         40,
				MaximumDeadContainersPerPod: 1,
				CGroupDriver:                constants.CGroupDriverCGroupFS,
				CgroupsPerQOS:               true,
				CgroupRoot:                  "",
				CNIConfDir:                  "/etc/cni/net.d",
				CNIBinDir:                   "/opt/cni/bin",
				CNICacheDir:                 "/var/lib/cni/cache",
				NetworkPluginMTU:            1500,
				VolumeStatsAggPeriod:        time.Second * 60,
				EnableMetrics:               true,
				NetworkPluginName:           "cni",
			},
			EdgeHub: &EdgeHub{
				Enable:             true,
				Heartbeat:          15,
				MessageQPS:         30,
				MessageBurst:       60,
				ProjectID:          "e632aba927ea4ac2b575ec1603d56f10",
				TLSCAFile:          filepath.Join(rootDir, "/ca/rootCA.crt"),
				TLSCertFile:        filepath.Join(rootDir, "/certs/server.crt"),
				TLSPrivateKeyFile:  filepath.Join(rootDir, "/certs/server.key"),
				RotateCertificates: true,
				Quic: &EdgeHubQUIC{
					Enable:           false,
					HandshakeTimeout: 30,
					ReadDeadline:     15,
					Server:           net.JoinHostPort(localIP, "10001"),
					WriteDeadline:    15,
				},
				WebSocket: &EdgeHubWebSocket{
					Enable:           true,
					HandshakeTimeout: 30,
					ReadDeadline:     15,
					Server:           net.JoinHostPort(localIP, "10000"),
					WriteDeadline:    15,
				},
				HTTPServer: (&url.URL{
					Scheme: "https",
					Host:   net.JoinHostPort(localIP, "10002"),
				}).String(),
				Token: "",
			},
			EventBus: &EventBus{
				Enable:               true,
				MqttQOS:              0,
				MqttRetain:           false,
				MqttSessionQueueSize: 100,
				MqttServerExternal:   "tcp://127.0.0.1:1883",
				MqttServerInternal:   "tcp://127.0.0.1:1884",
				MqttSubClientID:      "",
				MqttPubClientID:      "",
				MqttUsername:         "",
				MqttPassword:         "",
				MqttMode:             constants.MqttModeExternal,
				TLS: &EventBusTLS{
					Enable:                false,
					TLSMqttCAFile:         filepath.Join(rootDir, "/ca/rootCA.crt"),
					TLSMqttCertFile:       filepath.Join(rootDir, "/certs/server.crt"),
					TLSMqttPrivateKeyFile: filepath.Join(rootDir, "/certs/server.key"),
				},
			},
			MetaManager: &MetaManager{
				Enable:             true,
				ContextSendGroup:   "hub",
				ContextSendModule:  "websocket",
				RemoteQueryTimeout: 60,
				Debug:              false,
				MetaServer: &MetaServer{
					Enable:            true,
					Debug:             false,
					Server:            "127.0.0.1:10550",
					Scheme:            "https",
					TLSCAFile:         filepath.Join(rootDir, "/ca/rootCA.crt"),
					TLSCertFile:       filepath.Join(rootDir, "/certs/server.crt"),
					TLSPrivateKeyFile: filepath.Join(rootDir, "/certs/server.key"),
				},
			},
			ServiceBus: &ServiceBus{
				Enable:  false,
				Server:  "127.0.0.1",
				Port:    9060,
				Timeout: 60,
			},
			DeviceTwin: &DeviceTwin{
				Enable: true,
			},
			DBTest: &DBTest{
				Enable: false,
			},
			EdgeStream: &EdgeStream{
				Enable:                  true,
				TLSTunnelCAFile:         filepath.Join(rootDir, "/ca/rootCA.crt"),
				TLSTunnelCertFile:       filepath.Join(rootDir, "/certs/server.crt"),
				TLSTunnelPrivateKeyFile: filepath.Join(rootDir, "/certs/server.key"),
				HandshakeTimeout:        30,
				ReadDeadline:            15,
				TunnelServer:            net.JoinHostPort("127.0.0.1", "10004"),
				WriteDeadline:           15,
			},
		},
	}
	return cfg
}

// EdgeCore indicates the EdgeCore config which read from EdgeCore config file
type EdgeCore struct {
	metaV1.TypeMeta
	// DataBase indicates database info
	// +Required
	DataBase *DataBase `json:"database,omitempty"`
	// Modules indicates EdgeCore modules config
	// +Required
	Modules *EdgeModules `json:"modules,omitempty"`
	// FeatureGates is a map of feature names to bools that enable or disable alpha/experimental features.
	FeatureGates map[string]bool `json:"featureGates,omitempty"`
}

// DataBase indicates the database info
type DataBase struct {
	// DriverName indicates database driver name
	// default "sqlite3"
	DriverName string `json:"driverName,omitempty"`
	// AliasName indicates alias name
	// default "default"
	AliasName string `json:"aliasName,omitempty"`
	// DataSource indicates the data source path
	// default "/var/lib/kubeedge/edgecore.db"
	DataSource string `json:"dataSource,omitempty"`
}

// EdgeModules indicates the modules which edgeCore will be used
type EdgeModules struct {
	// Edged indicates edged module config
	// +Required
	Edged *Edged `json:"edged,omitempty"`
	// EdgeHub indicates edgeHub module config
	// +Required
	EdgeHub *EdgeHub `json:"edgeHub,omitempty"`
	// EventBus indicates eventBus config for edgeCore
	// +Required
	EventBus *EventBus `json:"eventBus,omitempty"`
	// MetaManager indicates meta module config
	// +Required
	MetaManager *MetaManager `json:"metaManager,omitempty"`
	// ServiceBus indicates serviceBus module config
	ServiceBus *ServiceBus `json:"serviceBus,omitempty"`
	// DeviceTwin indicates deviceTwin module config
	DeviceTwin *DeviceTwin `json:"deviceTwin,omitempty"`
	// DBTest indicates dbTest module config
	DBTest *DBTest `json:"dbTest,omitempty"`
	// EdgeStream indicates edgestream module config
	// +Required
	EdgeStream *EdgeStream `json:"edgeStream,omitempty"`
}

// Edged indicates the config fo edged module
// edged is lighted-kubelet
type Edged struct {
	// Enable indicates whether the module is enabled
	// if set to false (for debugging etc.), skip checking other edged configs.
	// default true
	Enable bool `json:"enable"`
	// Labels indicates current node labels
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations indicates current node annotations
	Annotations map[string]string `json:"annotations,omitempty"`
	// Taints indicates current node taints
	Taints []coreV1.Taint `json:"taints,omitempty"`
	// NodeStatusUpdateFrequency indicates node status update frequency (second)
	// default 10
	NodeStatusUpdateFrequency int32 `json:"nodeStatusUpdateFrequency,omitempty"`
	// RuntimeType indicates cri runtime ,support: docker, remote
	// default "docker"
	RuntimeType string `json:"runtimeType,omitempty"`
	// DockerAddress indicates docker server address
	// default "unix:///var/run/docker.sock"
	DockerAddress string `json:"dockerAddress,omitempty"`
	// RemoteRuntimeEndpoint indicates remote runtime endpoint
	// default "unix:///var/run/dockershim.sock"
	RemoteRuntimeEndpoint string `json:"remoteRuntimeEndpoint,omitempty"`
	// RemoteImageEndpoint indicates remote image endpoint
	// default "unix:///var/run/dockershim.sock"
	RemoteImageEndpoint string `json:"remoteImageEndpoint,omitempty"`
	// NodeIP indicates current node ip.
	// Setting the value overwrites the automatically detected IP address
	// default get local host ip
	NodeIP string `json:"nodeIP"`
	// ClusterDNS indicates cluster dns
	// Note: Can not use "omitempty" option,  It will affect the output of the default configuration file
	// +Required
	ClusterDNS string `json:"clusterDNS"`
	// ClusterDomain indicates cluster domain
	// Note: Can not use "omitempty" option,  It will affect the output of the default configuration file
	ClusterDomain string `json:"clusterDomain"`
	// EdgedMemoryCapacity indicates memory capacity (byte)
	// default 7852396000
	EdgedMemoryCapacity int64 `json:"edgedMemoryCapacity,omitempty"`
	// PodSandboxImage is the image whose network/ipc namespaces containers in each pod will use.
	// +Required
	// default kubeedge/pause:3.1
	PodSandboxImage string `json:"podSandboxImage,omitempty"`
	// ImagePullProgressDeadline indicates image pull progress dead line (second)
	// default 60
	ImagePullProgressDeadline int32 `json:"imagePullProgressDeadline,omitempty"`
	// RuntimeRequestTimeout indicates runtime request timeout (second)
	// default 2
	RuntimeRequestTimeout int32 `json:"runtimeRequestTimeout,omitempty"`
	// HostnameOverride indicates hostname
	// default os.Hostname()
	HostnameOverride string `json:"hostnameOverride,omitempty"`
	// RegisterNode enables automatic registration
	// default true
	RegisterNode bool `json:"registerNode,omitempty"`
	//RegisterNodeNamespace indicates register node namespace
	// default "default"
	RegisterNodeNamespace string `json:"registerNodeNamespace,omitempty"`
	// CustomInterfaceName indicates the name of the network interface used for obtaining the IP address.
	// Setting this will override the setting 'NodeIP' if provided.
	// If this is not defined the IP address is obtained by the hostname.
	// default ""
	CustomInterfaceName string `json:"customInterfaceName,omitempty"`
	// ConcurrentConsumers indicates concurrent consumers for pod add or remove operation
	// default 5
	ConcurrentConsumers int `json:"concurrentConsumers,omitempty"`
	// DevicePluginEnabled indicates enable device plugin
	// default false
	// Note: Can not use "omitempty" option, it will affect the output of the default configuration file
	DevicePluginEnabled bool `json:"devicePluginEnabled"`
	// GPUPluginEnabled indicates enable gpu plugin
	// default false,
	// Note: Can not use "omitempty" option, it will affect the output of the default configuration file
	GPUPluginEnabled bool `json:"gpuPluginEnabled"`
	// ImageGCHighThreshold indicates image gc high threshold (percent)
	// default 80
	ImageGCHighThreshold int32 `json:"imageGCHighThreshold,omitempty"`
	// ImageGCLowThreshold indicates image gc low threshold (percent)
	// default 40
	ImageGCLowThreshold int32 `json:"imageGCLowThreshold,omitempty"`
	// MaximumDeadContainersPerPod indicates max num dead containers per pod
	// default 1
	MaximumDeadContainersPerPod int32 `json:"maximumDeadContainersPerPod,omitempty"`
	// CGroupDriver indicates container cgroup driver, support: cgroupfs, systemd
	// default "cgroupfs"
	// +Required
	CGroupDriver string `json:"cgroupDriver,omitempty"`
	// NetworkPluginName indicates the name of the network plugin to be invoked,
	// if an empty string is specified, use noop plugin
	// default ""
	NetworkPluginName string `json:"networkPluginName,omitempty"`
	// CNIConfDir indicates the full path of the directory in which to search for CNI config files
	// default "/etc/cni/net.d"
	CNIConfDir string `json:"cniConfDir,omitempty"`
	// CNIBinDir indicates a comma-separated list of full paths of directories
	// in which to search for CNI plugin binaries
	// default "/opt/cni/bin"
	CNIBinDir string `json:"cniBinDir,omitempty"`
	// CNICacheDir indicates the full path of the directory in which CNI should store cache files
	// default "/var/lib/cni/cache"
	CNICacheDir string `json:"cniCacheDirs,omitempty"`
	// NetworkPluginMTU indicates the MTU to be passed to the network plugin
	// default 1500
	NetworkPluginMTU int32 `json:"networkPluginMTU,omitempty"`
	// CgroupsPerQOS enables QoS based Cgroup hierarchy: top level cgroups for QoS Classes
	// And all Burstable and BestEffort pods are brought up under their
	// specific top level QoS cgroup.
	// Default: true
	CgroupsPerQOS bool `json:"cgroupsPerQOS"`
	// CgroupRoot is the root cgroup to use for pods.
	// If CgroupsPerQOS is enabled, this is the root of the QoS cgroup hierarchy.
	// Default: ""
	CgroupRoot string `json:"cgroupRoot"`
	// EdgeCoreCgroups is the absolute name of cgroups to isolate the edgecore in
	// Dynamic Kubelet Config (beta): This field should not be updated without a full node
	// reboot. It is safest to keep this value the same as the local config.
	// Default: ""
	EdgeCoreCgroups string `json:"edgeCoreCgroups,omitempty"`
	// systemCgroups is absolute name of cgroups in which to place
	// all non-kernel processes that are not already in a container. Empty
	// for no container. Rolling back the flag requires a reboot.
	// Dynamic Kubelet Config (beta): This field should not be updated without a full node
	// reboot. It is safest to keep this value the same as the local config.
	// Default: ""
	SystemCgroups string `json:"systemCgroups,omitempty"`
	// How frequently to calculate and cache volume disk usage for all pods
	// Dynamic Kubelet Config (beta): If dynamically updating this field, consider that
	// shortening the period may carry a performance impact.
	// Default: "1m"
	VolumeStatsAggPeriod time.Duration `json:"volumeStatsAggPeriod,omitempty"`
	// EnableMetrics indicates whether enable the metrics
	// default true
	EnableMetrics bool `json:"enableMetrics,omitempty"`
	// Debug enable debug mode, it is helpful to debug Cloud Hub and Edge Hub
	Debug bool `json:"debug,omitempty"`
	// DebugEdgedAddr In normal mode, Edge Core cannot run on local.
	// Therefore, the Edge invocation address is exposed for making a remote invocation of the edged service
	DebugEdgedAddr string `json:"debugEdgedAddr,omitempty"`
	// PodStatusSyncInterval indicates pod status sync
	// default 60
	PodStatusSyncInterval int32 `json:"podStatusSyncInterval,omitempty"`
}

// EdgeHub indicates the EdgeHub module config
type EdgeHub struct {
	// Enable indicates whether the module is enabled
	// if set to false (for debugging etc.), skip checking other EdgeHub configs.
	// default true
	Enable bool `json:"enable"`
	// Heartbeat indicates heart beat (second)
	// default 15
	Heartbeat int32 `json:"heartbeat,omitempty"`
	// MessageQPS is the QPS to allow while send message to cloudHub.
	// DefaultQPS: 30
	MessageQPS int32 `json:"messageQPS,omitempty"`
	// MessageBurst is the burst to allow while send message to cloudHub.
	// DefaultBurst: 60
	MessageBurst int32 `json:"messageBurst,omitempty"`
	// ProjectID indicates project id
	// default e632aba927ea4ac2b575ec1603d56f10
	ProjectID string `json:"projectID,omitempty"`
	// TLSCAFile set ca file path
	// default "/etc/kubeedge/ca/rootCA.crt"
	TLSCAFile string `json:"tlsCaFile,omitempty"`
	// TLSCertFile indicates the file containing x509 Certificate for HTTPS
	// default "/etc/kubeedge/certs/server.crt"
	TLSCertFile string `json:"tlsCertFile,omitempty"`
	// TLSPrivateKeyFile indicates the file containing x509 private key matching tlsCertFile
	// default "/etc/kubeedge/certs/server.key"
	TLSPrivateKeyFile string `json:"tlsPrivateKeyFile,omitempty"`
	// Quic indicates quic config for EdgeHub module
	// Optional if websocket is configured
	Quic *EdgeHubQUIC `json:"quic,omitempty"`
	// WebSocket indicates websocket config for EdgeHub module
	// Optional if quic is configured
	WebSocket *EdgeHubWebSocket `json:"websocket,omitempty"`
	// Token indicates the priority of joining the cluster for the edge
	Token string `json:"token"`
	// HTTPServer indicates the server for edge to apply for the certificate.
	HTTPServer string `json:"httpServer,omitempty"`
	// RotateCertificates indicates whether edge certificate can be rotated
	// default true
	RotateCertificates bool `json:"rotateCertificates,omitempty"`
	// Debug enable debug mode, it is helpful to debug Cloud Hub and Edge Hub
	Debug bool `json:"debug,omitempty"`
	// DebugTargetEdgedServer the DebugEdgedAddr that is set in edged
	DebugTargetEdgedServer string `json:"debugTargetEdgedServer,omitempty"`
	// DebugEdgeHubAddr Edge Hub expose addr for metaClient to use
	DebugEdgeHubAddr string `json:"debugEdgeHubAddr,omitempty"`
}

// EdgeHubQUIC indicates the quic client config
type EdgeHubQUIC struct {
	// Enable indicates whether enable this protocol
	// default false
	Enable bool `json:"enable"`
	// HandshakeTimeout indicates hand shake timeout (second)
	// default 30
	HandshakeTimeout int32 `json:"handshakeTimeout,omitempty"`
	// ReadDeadline indicates read dead line (second)
	// default 15
	ReadDeadline int32 `json:"readDeadline,omitempty"`
	// Server indicates quic server address (ip:port)
	// +Required
	Server string `json:"server,omitempty"`
	// WriteDeadline indicates write deadline (second)
	// default 15
	WriteDeadline int32 `json:"writeDeadline,omitempty"`
}

// EdgeHubWebSocket indicates the websocket client config
type EdgeHubWebSocket struct {
	// Enable indicates whether enable this protocol
	// default true
	Enable bool `json:"enable"`
	// HandshakeTimeout indicates handshake timeout (second)
	// default  30
	HandshakeTimeout int32 `json:"handshakeTimeout,omitempty"`
	// ReadDeadline indicates read dead line (second)
	// default 15
	ReadDeadline int32 `json:"readDeadline,omitempty"`
	// Server indicates websocket server address (ip:port)
	// +Required
	Server string `json:"server,omitempty"`
	// WriteDeadline indicates write dead line (second)
	// default 15
	WriteDeadline int32 `json:"writeDeadline,omitempty"`
}

// EventBus indicates the event bus module config
type EventBus struct {
	// Enable indicates whether the module is enabled
	// skip checking other EventBus configs.
	// default true
	Enable bool `json:"enable"`
	// MqttQOS indicates mqtt qos
	// 0: QOSAtMostOnce, 1: QOSAtLeastOnce, 2: QOSExactlyOnce
	// default 0
	// Note: Can not use "omitempty" option,  It will affect the output of the default configuration file
	MqttQOS uint8 `json:"mqttQOS"`
	// MqttRetain indicates whether server will store the message and can be delivered to future subscribers,
	// if this flag set true, sever will store the message and can be delivered to future subscribers
	// default false
	// Note: Can not use "omitempty" option,  It will affect the output of the default configuration file
	MqttRetain bool `json:"mqttRetain"`
	// MqttSessionQueueSize indicates the size of how many sessions will be handled.
	// default 100
	MqttSessionQueueSize int32 `json:"mqttSessionQueueSize,omitempty"`
	// MqttServerInternal indicates internal mqtt broker url
	// default "tcp://127.0.0.1:1884"
	MqttServerInternal string `json:"mqttServerInternal,omitempty"`
	// MqttServerExternal indicates external mqtt broker url
	// default "tcp://127.0.0.1:1883"
	MqttServerExternal string `json:"mqttServerExternal,omitempty"`
	// MqttSubClientID indicates mqtt subscribe ClientID
	// default ""
	MqttSubClientID string `json:"mqttSubClientID"`
	// MqttPubClientID indicates mqtt publish ClientID
	// default ""
	MqttPubClientID string `json:"mqttPubClientID"`
	// MqttUsername indicates mqtt username
	// default ""
	MqttUsername string `json:"mqttUsername"`
	// MqttPassword indicates mqtt password
	// default ""
	MqttPassword string `json:"mqttPassword"`
	// MqttMode indicates which broker type will be choose
	// 0: internal mqtt broker enable only.
	// 1: internal and external mqtt broker enable.
	// 2: external mqtt broker enable only
	// +Required
	// default: 2
	MqttMode constants.MqttMode `json:"mqttMode"`
	// Tls indicates tls config for EventBus module
	TLS *EventBusTLS `json:"eventBusTLS,omitempty"`
}

// EventBusTLS indicates the EventBus tls config with MQTT broker
type EventBusTLS struct {
	// Enable indicates whether enable tls connection
	// default false
	Enable bool `json:"enable"`
	// TLSMqttCAFile sets ca file path
	// default "/etc/kubeedge/ca/rootCA.crt"
	TLSMqttCAFile string `json:"tlsMqttCAFile,omitempty"`
	// TLSMqttCertFile indicates the file containing x509 Certificate for HTTPS
	// default "/etc/kubeedge/certs/server.crt"
	TLSMqttCertFile string `json:"tlsMqttCertFile,omitempty"`
	// TLSMqttPrivateKeyFile indicates the file containing x509 private key matching tlsMqttCertFile
	// default "/etc/kubeedge/certs/server.key"
	TLSMqttPrivateKeyFile string `json:"tlsMqttPrivateKeyFile,omitempty"`
}

// MetaManager indicates the MetaManager module config
type MetaManager struct {
	// Enable indicates whether the module is enabled
	// default true
	Enable bool `json:"enable"`
	// ContextSendGroup indicates send group
	ContextSendGroup string `json:"contextSendGroup,omitempty"`
	// ContextSendModule indicates send module
	ContextSendModule string `json:"contextSendModule,omitempty"`
	// RemoteQueryTimeout indicates remote query timeout (second)
	// default 60
	RemoteQueryTimeout int32 `json:"remoteQueryTimeout,omitempty"`
	// The config of MetaServer
	MetaServer *MetaServer `json:"metaServer,omitempty"`
	// Debug enable debug mode, it is helpful to debug Cloud Hub and Edge Hub
	Debug bool `json:"debug,omitempty"`
	// DebugTargetEdgeHubAddr the debugEdgeHubAddr set in Edge Hub
	DebugTargetEdgeHubAddr string `json:"debugTargetEdgeHubAddr,omitempty"`
}

type MetaServer struct {
	// Enable indicates whether the module is enabled
	// default true
	Enable bool   `json:"enable"`
	Debug  bool   `json:"debug,omitempty"`
	Server string `json:"server,omitempty"`
	Scheme string `json:"scheme,omitempty"`
	// TLSCAFile set ca file path
	// default "/etc/kubeedge/ca/rootCA.crt"
	TLSCAFile string `json:"tlsCAFile,omitempty"`
	// TLSCertFile indicates the file containing x509 Certificate for HTTPS
	// default "/etc/kubeedge/certs/server.crt"
	TLSCertFile string `json:"tlsCertFile,omitempty"`
	// TLSPrivateKeyFile indicates the file containing x509 private key matching tlsCertFile
	// default "/etc/kubeedge/certs/server.key"
	TLSPrivateKeyFile string `json:"tlsPrivateKeyFile,omitempty"`
}

// ServiceBus indicates the ServiceBus module config
type ServiceBus struct {
	// Enable indicates whether ServiceBus is enabled,
	// if set to false (for debugging etc.), skip checking other ServiceBus configs.
	// default false
	Enable bool `json:"enable"`
	// Address indicates address for http server
	Server string `json:"server"`
	// Port indicates port for http server
	Port int `json:"port"`
	// Timeout indicates timeout for servicebus receive message
	Timeout int `json:"timeout"`
}

// DeviceTwin indicates the DeviceTwin module config
type DeviceTwin struct {
	// Enable indicates whether the module is enabled
	// if set to false (for debugging etc.), skip checking other DeviceTwin configs.
	// default true
	Enable bool `json:"enable"`
}

// DBTest indicates the DBTest module config
type DBTest struct {
	// Enable indicates whether DBTest is enabled,
	// if set to false (for debugging etc.), skip checking other DBTest configs.
	// default false
	Enable bool `json:"enable"`
}

// EdgeStream indicates the stream controller
type EdgeStream struct {
	// Enable indicates whether the module is enabled
	// default true
	Enable bool `json:"enable"`
	// TLSTunnelCAFile indicates ca file path
	// default /etc/kubeedge/ca/rootCA.crt
	TLSTunnelCAFile string `json:"tlsTunnelCAFile,omitempty"`

	// TLSTunnelCertFile indicates the file containing x509 Certificate for HTTPS
	// default /etc/kubeedge/certs/server.crt
	TLSTunnelCertFile string `json:"tlsTunnelCertFile,omitempty"`
	// TLSTunnelPrivateKeyFile indicates the file containing x509 private key matching tlsCertFile
	// default /etc/kubeedge/certs/server.key
	TLSTunnelPrivateKeyFile string `json:"tlsTunnelPrivateKeyFile,omitempty"`

	// HandshakeTimeout indicates handshake timeout (second)
	// default  30
	HandshakeTimeout int32 `json:"handshakeTimeout,omitempty"`
	// ReadDeadline indicates read deadline (second)
	// default 15
	ReadDeadline int32 `json:"readDeadline,omitempty"`
	// TunnelServer indicates websocket server address (ip:port)
	// +Required
	TunnelServer string `json:"server,omitempty"`
	// WriteDeadline indicates write deadline (second)
	// default 15
	WriteDeadline int32 `json:"writeDeadline,omitempty"`
}
