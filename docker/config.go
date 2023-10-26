package docker

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-units"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

// Config - Docker Component configuration
type Config struct {
	// Name - the name of the container, and the ID of the blueprint component
	// Name cannot be empty
	Name string `json:"name" yaml:"name"`

	// Env - environment variables for the container
	Env map[string]string `json:"env,omitempty" yaml:"env,omitempty"`

	// Ports - list of ports to expose
	// we don't map internal ports to a different external ports since it won't be consistent
	// for open network situations used in local development powered by docker network mode "host"
	Ports []Port `json:"ports,omitempty" yaml:"ports,omitempty"`

	// Waiters - list of waiters. A waiter is a function responsible for waiting for healthy status
	// of the container before finishing the container start process
	Waiters []Waiter `json:"waiters,omitempty" yaml:"waiters,omitempty"`

	// ImagePullOptions - options for pulling the container image
	ImagePullOptions *ImagePullOptions `json:"image_pull_options,omitempty" yaml:"image_pull_options,omitempty"`

	// Hostname - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L71
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	// Domainname - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L72
	Domainname string `json:"domainname,omitempty" yaml:"domainname,omitempty"`

	// User - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L73
	User string `json:"user,omitempty" yaml:"user,omitempty"`

	// AttachStdin - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L74
	AttachStdin bool `json:"attach_stdin,omitempty" yaml:"attach_stdin,omitempty"`

	// AttachStdout - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L75
	AttachStdout bool `json:"attach_stdout,omitempty" yaml:"attach_stdout,omitempty"`

	// AttachStderr - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L76
	AttachStderr bool `json:"attach_stderr,omitempty" yaml:"attach_stderr,omitempty"`

	// Tty - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L78
	Tty bool `json:"tty,omitempty" yaml:"tty,omitempty"`

	// OpenStdin - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L79
	OpenStdin bool `json:"open_stdin,omitempty" yaml:"open_stdin,omitempty"`

	// StdinOnce - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L80
	StdinOnce bool `json:"stdin_once,omitempty" yaml:"stdin_once,omitempty"`

	// Cmd - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L82
	Cmd StrSlice `json:"cmd,omitempty" yaml:"cmd,omitempty"`

	// Healthcheck - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L83
	Healthcheck *Healthcheck `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty"`

	// ArgsEscaped - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L84
	ArgsEscaped bool `json:"args_escaped,omitempty" yaml:"args_escaped,omitempty"`

	// Image - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L85
	// Image cannot be empty
	Image string `json:"image,omitempty" yaml:"image,omitempty"`

	// Volumes - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L86
	Volumes map[string]struct{} `json:"volumes,omitempty" yaml:"volumes,omitempty"`

	// WorkingDir - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L87
	WorkingDir string `json:"working_dir,omitempty" yaml:"working_dir,omitempty"`

	// Entrypoint - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L88
	Entrypoint StrSlice `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`

	// NetworkDisabled - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L89
	NetworkDisabled bool `json:"network_disabled,omitempty" yaml:"network_disabled,omitempty"`

	// MacAddress - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L90
	MacAddress string `json:"mac_address,omitempty" yaml:"mac_address,omitempty"`

	// OnBuild - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L91
	OnBuild []string `json:"on_build,omitempty" yaml:"on_build,omitempty"`

	// Labels - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L92
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// StopSignal - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L93
	StopSignal string `json:"stop_signal,omitempty" yaml:"stop_signal,omitempty"`

	// StopTimeout - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L94
	StopTimeout *int `json:"stop_timeout,omitempty" yaml:"stop_timeout,omitempty"`

	// Shell - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L95
	Shell StrSlice `json:"shell,omitempty" yaml:"shell,omitempty"`

	// Binds - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L381
	Binds []string `json:"binds,omitempty" yaml:"binds,omitempty"`

	// ContainerIDFile - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L382
	ContainerIDFile string `json:"container_id_file,omitempty" yaml:"container_id_file,omitempty"`

	// LogConfig - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L383
	LogConfig *LogConfig `json:"log_config,omitempty" yaml:"log_config,omitempty"`

	// RestartPolicy - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L386
	RestartPolicy *RestartPolicy `json:"restart_policy,omitempty" yaml:"restart_policy,omitempty"`

	// VolumeDriver - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L388
	VolumeDriver string `json:"volume_driver,omitempty" yaml:"volume_driver,omitempty"`

	// VolumesFrom - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L389
	VolumesFrom []string `json:"volumes_from,omitempty" yaml:"volumes_from,omitempty"`

	// ConsoleSize - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L390
	ConsoleSize []uint `json:"console_size,omitempty" yaml:"console_size,omitempty"`

	// Annotations - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L391
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`

	// CapAdd - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L394
	CapAdd StrSlice `json:"cap_add,omitempty" yaml:"cap_add,omitempty"`

	// CapDrop - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L395
	CapDrop StrSlice `json:"cap_drop,omitempty" yaml:"cap_drop,omitempty"`

	// CgroupnsMode - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L396
	CgroupnsMode container.CgroupnsMode `json:"cgroupns_mode,omitempty" yaml:"cgroupns_mode,omitempty"`

	// DNS - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L397
	DNS []string `json:"dns,omitempty" yaml:"dns,omitempty"`

	// DNSOptions - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L398
	DNSOptions []string `json:"dns_options,omitempty" yaml:"dns_options,omitempty"`

	// DNSSearch - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L399
	DNSSearch []string `json:"dns_search,omitempty" yaml:"dns_search,omitempty"`

	// ExtraHosts - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L400
	ExtraHosts []string `json:"extra_hosts,omitempty" yaml:"extra_hosts,omitempty"`

	// GroupAdd - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L401
	GroupAdd []string `json:"group_add,omitempty" yaml:"group_add,omitempty"`

	// IpcMode - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L402
	IpcMode container.IpcMode `json:"ipc_mode,omitempty" yaml:"ipc_mode,omitempty"`

	// Cgroup - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L403
	Cgroup container.CgroupSpec `json:"cgroup,omitempty" yaml:"cgroup,omitempty"`

	// Links - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L404
	Links []string `json:"links,omitempty" yaml:"links,omitempty"`

	// OomScoreAdj - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L405
	OomScoreAdj int `json:"oom_score_adj,omitempty" yaml:"oom_score_adj,omitempty"`

	// PidMode - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L406
	PidMode container.PidMode `json:"pid_mode,omitempty" yaml:"pid_mode,omitempty"`

	// Privileged - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L407
	Privileged bool `json:"privileged,omitempty" yaml:"privileged,omitempty"`

	// PublishAllPorts - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L408
	PublishAllPorts bool `json:"publish_all_ports,omitempty" yaml:"publish_all_ports,omitempty"`

	// ReadonlyRootfs - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L409
	ReadonlyRootfs bool `json:"readonly_rootfs,omitempty" yaml:"readonly_rootfs,omitempty"`

	// SecurityOpt - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L410
	SecurityOpt []string `json:"security_opt,omitempty" yaml:"security_opt,omitempty"`

	// StorageOpt - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L411
	StorageOpt map[string]string `json:"storage_opt,omitempty" yaml:"storage_opt,omitempty"`

	// Tmpfs - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L412
	Tmpfs map[string]string `json:"tmpfs,omitempty" yaml:"tmpfs,omitempty"`

	// UTSMode - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L413
	UTSMode container.UTSMode `json:"uts_mode,omitempty" yaml:"uts_mode,omitempty"`

	// UsernsMode - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L414
	UsernsMode container.UsernsMode `json:"userns_mode,omitempty" yaml:"userns_mode,omitempty"`

	// ShmSize - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L415
	ShmSize int64 `json:"shm_size,omitempty" yaml:"shm_size,omitempty"`

	// Sysctls - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L416
	Sysctls map[string]string `json:"sysctls,omitempty" yaml:"sysctls,omitempty"`

	// Runtime - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L417
	Runtime string `json:"runtime,omitempty" yaml:"runtime,omitempty"`

	// Isolation - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L420
	Isolation container.Isolation `json:"isolation,omitempty" yaml:"isolation,omitempty"`

	// Resources - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L423
	Resources *Resources `json:"resources,omitempty" yaml:"resources,omitempty"`

	// Mounts - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L426
	Mounts []Mount `json:"mounts,omitempty" yaml:"mounts,omitempty"`

	// MaskedPaths - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L429
	MaskedPaths []string `json:"masked_paths,omitempty" yaml:"masked_paths,omitempty"`

	// ReadonlyPaths - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L432
	ReadonlyPaths []string `json:"readonly_paths,omitempty" yaml:"readonly_paths,omitempty"`

	// Init - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L435
	Init *bool `json:"init,omitempty" yaml:"init,omitempty"`

	// PlatformConfig describes the platform which the image in the manifest runs on
	PlatformConfig *PlatformConfig `json:"platform_config,omitempty" yaml:"platform_config,omitempty"`
}

type Port struct {
	// Port to expose
	Port string `json:"port,omitempty" yaml:"port,omitempty"`

	// Protocol to use - "tcp"/"udp". can be left empty, the default value in this case will be "tcp"
	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}

type WaiterType string

const (
	// WaiterTypeString - waits for a log line to contain a string value
	WaiterTypeString WaiterType = "string"

	// WaiterTypeRegex - waits for a log line to match a regex
	WaiterTypeRegex WaiterType = "regex"

	// WaiterTypeDuration - waits for a certain amount of time
	WaiterTypeDuration WaiterType = "duration"
)

type Waiter struct {
	// Type - WaiterType type to use
	Type WaiterType `json:"type,omitempty" yaml:"type,omitempty"`

	// String - only for Type == "string"
	// the string value to wait for
	String string `json:"string,omitempty" yaml:"string,omitempty"`

	// Regex - only for Type == "regex"
	// the regex value to wait for
	// compiled to a go regexp.Regexp using re2 syntax
	Regex string `json:"regex,omitempty" yaml:"regex,omitempty"`

	// Duration - only for Type == "duration"
	// the duration to wait
	// parsed as a go duration using time.ParseDuration
	Duration string `json:"duration,omitempty" yaml:"duration,omitempty"`
}

type ImagePullOptions struct {
	// Disabled allow disabling image pull/remove
	// this is useful when the image already exists on the machine and we want it to remain after cleanup
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	// All - used for https://github.com/moby/moby/blob/v24.0.6/api/types/client.go#L279
	All bool `json:"all,omitempty" yaml:"all,omitempty"`

	// RegistryAuth - used for https://github.com/moby/moby/blob/v24.0.6/api/types/client.go#L280
	// only available when RegistryAuthFunc is not set
	RegistryAuth string `json:"registry_auth,omitempty" yaml:"registry_auth,omitempty"`

	// PrivilegeFunc - used for https://github.com/moby/moby/blob/v24.0.6/api/types/client.go#L281
	// available only via code, not available in config files
	PrivilegeFunc types.RequestPrivilegeFunc `json:"-" yaml:"-"`

	// Platform - used for https://github.com/moby/moby/blob/v24.0.6/api/types/client.go#L282
	Platform string `json:"platform,omitempty" yaml:"platform,omitempty"`

	// a lazy load function for the RegistryAuth
	// available only via code, not available in config files
	// used when loading the auth file take a long time, and you want to avoid loading it when it's not needed
	RegistryAuthFunc func() (string, error) `json:"-" yaml:"-"`
}

type Healthcheck struct {
	// Test - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L44
	Test []string `json:"test,omitempty" yaml:"test,omitempty"`

	// Interval - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L47
	Interval time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`

	// Timeout - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L48
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// StartPeriod - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L49
	StartPeriod time.Duration `json:"start_period,omitempty" yaml:"start_period,omitempty"`

	// Retries - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/config.go#L53
	Retries int `json:"retries,omitempty" yaml:"retries,omitempty"`
}

type LogConfig struct {
	// Type - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L321
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// Config - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L322
	Config map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
}

type RestartPolicy struct {
	// Name - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L274
	Name string

	// MaximumRetryCount - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L275
	MaximumRetryCount int
}

type Resources struct {
	// CPUShares - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L328
	CPUShares int64 `json:"cpu_shares,omitempty" yaml:"cpu_shares,omitempty"`

	// Memory - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L329
	Memory int64 `json:"memory,omitempty" yaml:"memory,omitempty"`

	// NanoCPUs - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L330
	NanoCPUs int64 `json:"nano_cp_us,omitempty" yaml:"nano_cp_us,omitempty"`

	// CgroupParent - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L333
	CgroupParent string `json:"cgroup_parent,omitempty" yaml:"cgroup_parent,omitempty"`

	// BlkioWeight - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L334
	BlkioWeight uint16 `json:"blkio_weight,omitempty" yaml:"blkio_weight,omitempty"`

	// BlkioWeightDevice - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L335
	BlkioWeightDevice []WeightDevice `json:"blkio_weight_device,omitempty" yaml:"blkio_weight_device,omitempty"`

	// BlkioDeviceReadBps - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L336
	BlkioDeviceReadBps []ThrottleDevice `json:"blkio_device_read_bps,omitempty" yaml:"blkio_device_read_bps,omitempty"`

	// BlkioDeviceWriteBps - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L337
	BlkioDeviceWriteBps []ThrottleDevice `json:"blkio_device_write_bps,omitempty" yaml:"blkio_device_write_bps,omitempty"`

	// BlkioDeviceReadIOps - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L338
	BlkioDeviceReadIOps []ThrottleDevice `json:"blkio_device_read_i_ops,omitempty" yaml:"blkio_device_read_i_ops,omitempty"`

	// BlkioDeviceWriteIOps - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L339
	BlkioDeviceWriteIOps []ThrottleDevice `json:"blkio_device_write_i_ops,omitempty" yaml:"blkio_device_write_i_ops,omitempty"`

	// CPUPeriod - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L340
	CPUPeriod int64 `json:"cpu_period,omitempty" yaml:"cpu_period,omitempty"`

	// CPUQuota - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L341
	CPUQuota int64 `json:"cpu_quota,omitempty" yaml:"cpu_quota,omitempty"`

	// CPURealtimePeriod - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L342
	CPURealtimePeriod int64 `json:"cpu_realtime_period,omitempty" yaml:"cpu_realtime_period,omitempty"`

	// CPURealtimeRuntime - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L343
	CPURealtimeRuntime int64 `json:"cpu_realtime_runtime,omitempty" yaml:"cpu_realtime_runtime,omitempty"`

	// CpusetCpus - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L344
	CpusetCpus string `json:"cpuset_cpus,omitempty" yaml:"cpuset_cpus,omitempty"`

	// CpusetMems - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L345
	CpusetMems string `json:"cpuset_mems,omitempty" yaml:"cpuset_mems,omitempty"`

	// Devices - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L346
	Devices []DeviceMapping `json:"devices,omitempty" yaml:"devices,omitempty"`

	// DeviceCgroupRules - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L347
	DeviceCgroupRules []string `json:"device_cgroup_rules,omitempty" yaml:"device_cgroup_rules,omitempty"`

	// DeviceRequests - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L348
	DeviceRequests []DeviceRequest `json:"device_requests,omitempty" yaml:"device_requests,omitempty"`

	// KernelMemoryTCP - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L353
	KernelMemoryTCP int64 `json:"kernel_memory_tcp,omitempty" yaml:"kernel_memory_tcp,omitempty"`

	// MemoryReservation - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L354
	MemoryReservation int64 `json:"memory_reservation,omitempty" yaml:"memory_reservation,omitempty"`

	// MemorySwap - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L355
	MemorySwap int64 `json:"memory_swap,omitempty" yaml:"memory_swap,omitempty"`

	// MemorySwappiness - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L356
	MemorySwappiness *int64 `json:"memory_swappiness,omitempty" yaml:"memory_swappiness,omitempty"`

	// OomKillDisable - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L357
	OomKillDisable *bool `json:"oom_kill_disable,omitempty" yaml:"oom_kill_disable,omitempty"`

	// PidsLimit - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L358
	PidsLimit *int64 `json:"pids_limit,omitempty" yaml:"pids_limit,omitempty"`

	// Ulimits - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L359
	Ulimits []Ulimit `json:"ulimits,omitempty" yaml:"ulimits,omitempty"`

	// CPUCount - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L362
	CPUCount int64 `json:"cpu_count,omitempty" yaml:"cpu_count,omitempty"`

	// CPUPercent - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L363
	CPUPercent int64 `json:"cpu_percent,omitempty" yaml:"cpu_percent,omitempty"`

	// IOMaximumIOps - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L364
	IOMaximumIOps uint64 `json:"io_maximum_i_ops,omitempty" yaml:"io_maximum_i_ops,omitempty"`

	// IOMaximumBandwidth - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L365
	IOMaximumBandwidth uint64 `json:"io_maximum_bandwidth,omitempty" yaml:"io_maximum_bandwidth,omitempty"`
}

type WeightDevice struct {
	// Path - used for https://github.com/moby/moby/blob/v24.0.6/api/types/blkiodev/blkio.go#L7
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Weight - used for https://github.com/moby/moby/blob/v24.0.6/api/types/blkiodev/blkio.go#L8
	Weight uint16 `json:"weight,omitempty" yaml:"weight,omitempty"`
}

type ThrottleDevice struct {
	// Path - used for https://github.com/moby/moby/blob/v24.0.6/api/types/blkiodev/blkio.go#L17
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Rate - used for https://github.com/moby/moby/blob/v24.0.6/api/types/blkiodev/blkio.go#L18
	Rate uint64 `json:"rate,omitempty" yaml:"rate,omitempty"`
}

type DeviceMapping struct {
	// PathOnHost - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L267
	PathOnHost string `json:"path_on_host,omitempty" yaml:"path_on_host,omitempty"`

	// PathInContainer - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L268
	PathInContainer string `json:"path_in_container,omitempty" yaml:"path_in_container,omitempty"`

	// CgroupPermissions - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L269
	CgroupPermissions string `json:"cgroup_permissions,omitempty" yaml:"cgroup_permissions,omitempty"`
}

type DeviceRequest struct {
	// Driver - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L258
	Driver string `json:"driver,omitempty" yaml:"driver,omitempty"`

	// Count - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L259
	Count int `json:"count,omitempty" yaml:"count,omitempty"`

	// DeviceIDs - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L260
	DeviceIDs []string `json:"device_ids,omitempty" yaml:"device_ids,omitempty"`

	// Capabilities - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L261
	Capabilities [][]string `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`

	// Options - used for https://github.com/moby/moby/blob/v24.0.6/api/types/container/hostconfig.go#L262
	Options map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
}

type Ulimit struct {
	// Name - used for https://github.com/docker/go-units/blob/master/ulimit.go#L11
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Hard - used for https://github.com/docker/go-units/blob/master/ulimit.go#L12
	Hard int64 `json:"hard,omitempty" yaml:"hard,omitempty"`

	// Soft - used for https://github.com/docker/go-units/blob/master/ulimit.go#L13
	Soft int64 `json:"soft,omitempty" yaml:"soft,omitempty"`
}

type Mount struct {
	// Type - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L26
	Type mount.Type `json:"type,omitempty" yaml:"type,omitempty"`

	// Source - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L30
	Source string `json:"source,omitempty" yaml:"source,omitempty"`

	// Target - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L31
	Target string `json:"target,omitempty" yaml:"target,omitempty"`

	// ReadOnly - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L32
	ReadOnly bool `json:"read_only,omitempty" yaml:"read_only,omitempty"`

	// Consistency - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L33
	Consistency mount.Consistency `json:"consistency,omitempty" yaml:"consistency,omitempty"`

	// BindOptions - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L35
	BindOptions *BindOptions `json:"bind_options,omitempty" yaml:"bind_options,omitempty"`

	// VolumeOptions - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L36
	VolumeOptions *VolumeOptions `json:"volume_options,omitempty" yaml:"volume_options,omitempty"`

	// TmpfsOptions - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L37
	TmpfsOptions *TmpfsOptions `json:"tmpfs_options,omitempty" yaml:"tmpfs_options,omitempty"`
}

type BindOptions struct {
	// Propagation - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L85
	Propagation mount.Propagation `json:"propagation,omitempty" yaml:"propagation,omitempty"`

	// NonRecursive - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L86
	NonRecursive bool `json:"non_recursive,omitempty" yaml:"non_recursive,omitempty"`

	// CreateMountpoint - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L87
	CreateMountpoint bool `json:"create_mountpoint,omitempty" yaml:"create_mountpoint,omitempty"`
}

type VolumeOptions struct {
	// NoCopy - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L92
	NoCopy bool `json:"no_copy,omitempty" yaml:"no_copy,omitempty"`

	// Labels - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L93
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// DriverConfig - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L94
	DriverConfig *Driver `json:"driver_config,omitempty" yaml:"driver_config,omitempty"`
}

type Driver struct {
	// Name - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L99
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Options - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L100
	Options map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
}

type TmpfsOptions struct {
	// SizeBytes - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L113
	SizeBytes int64 `json:"size_bytes,omitempty" yaml:"size_bytes,omitempty"`

	// Mode - used for https://github.com/moby/moby/blob/v24.0.6/api/types/mount/mount.go#L115
	Mode os.FileMode `json:"mode,omitempty" yaml:"mode,omitempty"`
}

type PlatformConfig struct {
	// Architecture - used for https://github.com/opencontainers/image-spec/blob/v1.1.0-rc4/specs-go/v1/descriptor.go#L56
	Architecture string `json:"architecture,omitempty" yaml:"architecture,omitempty"`

	// OS - used for https://github.com/opencontainers/image-spec/blob/v1.1.0-rc4/specs-go/v1/descriptor.go#L56
	OS string `json:"os,omitempty" yaml:"os,omitempty"`

	// OSVersion - used for https://github.com/opencontainers/image-spec/blob/v1.1.0-rc4/specs-go/v1/descriptor.go#L56
	OSVersion string `json:"os_version,omitempty" yaml:"os_version,omitempty"`

	// OSFeatures - used for https://github.com/opencontainers/image-spec/blob/v1.1.0-rc4/specs-go/v1/descriptor.go#L56
	OSFeatures []string `json:"os_features,omitempty" yaml:"os_features,omitempty"`

	// Variant - used for https://github.com/opencontainers/image-spec/blob/v1.1.0-rc4/specs-go/v1/descriptor.go#L56
	Variant string `json:"variant,omitempty" yaml:"variant,omitempty"`
}

type StrSlice strslice.StrSlice

func (s *StrSlice) UnmarshalYAML(value *yaml.Node) error {
	var p []string
	err := value.Decode(&p)
	if err != nil {
		var str string
		err = value.Decode(&str)
		if err != nil {
			return err
		}

		p = append(p, str)
	}

	*s = p
	return nil
}

type runConfig struct {
	hostname         string
	containerConfig  *container.Config
	hostConfig       *container.HostConfig
	networkingConfig *network.NetworkingConfig
	platformConfig   *ocispec.Platform
	waiters          []waiterFunc
}

func (c Config) initialize() (*runConfig, error) {
	if c.Name == "" {
		return nil, ErrInvalidConfig{Property: "name", Msg: "cannot be empty"}
	}

	if c.Image == "" {
		return nil, ErrInvalidConfig{Property: "image", Msg: "cannot be empty"}
	}

	if l := len(c.ConsoleSize); l != 0 && l != 2 {
		return nil, ErrInvalidConfig{Property: "console_size", Msg: "must have exactly two elements"}
	}

	waiters := make([]waiterFunc, len(c.Waiters))
	for i, waiter := range c.Waiters {
		f, err := validateWaiter(waiter)
		if err != nil {
			return nil, err
		}

		waiters[i] = f
	}

	result := &runConfig{
		containerConfig: c.containerConfig(),
		hostConfig:      c.hostConfig(),
		platformConfig:  c.PlatformConfig.build(),
		waiters:         waiters,
	}

	return result, nil
}

func (c Config) imagePullOptions() (types.ImagePullOptions, error) {
	result := types.ImagePullOptions{}

	if c.ImagePullOptions != nil {
		var auth string
		if c.ImagePullOptions.RegistryAuthFunc != nil {
			var err error
			auth, err = c.ImagePullOptions.RegistryAuthFunc()
			if err != nil {
				return types.ImagePullOptions{}, err
			}
		} else {
			auth = c.ImagePullOptions.RegistryAuth
		}

		result.All = c.ImagePullOptions.All
		result.RegistryAuth = auth
		result.PrivilegeFunc = c.ImagePullOptions.PrivilegeFunc
		result.Platform = c.ImagePullOptions.Platform
	}

	return result, nil
}

func (c Config) containerConfig() *container.Config {
	env := make([]string, 0, len(c.Env))
	for key, value := range c.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return &container.Config{
		Hostname:        c.Hostname,
		Domainname:      c.Domainname,
		User:            c.User,
		AttachStdin:     c.AttachStdin,
		AttachStdout:    c.AttachStdout,
		AttachStderr:    c.AttachStderr,
		ExposedPorts:    nil,
		Tty:             c.Tty,
		OpenStdin:       c.OpenStdin,
		StdinOnce:       c.StdinOnce,
		Env:             env,
		Cmd:             strslice.StrSlice(c.Cmd),
		Healthcheck:     c.Healthcheck.build(),
		ArgsEscaped:     c.ArgsEscaped,
		Image:           c.Image,
		Volumes:         c.Volumes,
		WorkingDir:      c.WorkingDir,
		Entrypoint:      strslice.StrSlice(c.Entrypoint),
		NetworkDisabled: c.NetworkDisabled,
		MacAddress:      c.MacAddress,
		OnBuild:         c.OnBuild,
		Labels:          c.Labels,
		StopSignal:      c.StopSignal,
		StopTimeout:     c.StopTimeout,
		Shell:           strslice.StrSlice(c.Shell),
	}
}

func (c *Healthcheck) build() *container.HealthConfig {
	if c == nil {
		return nil
	}

	return &container.HealthConfig{
		Test:        c.Test,
		Interval:    c.Interval,
		Timeout:     c.Timeout,
		StartPeriod: c.StartPeriod,
		Retries:     c.Retries,
	}
}

func (c Config) hostConfig() *container.HostConfig {
	var consoleSize [2]uint
	if len(c.ConsoleSize) > 0 {
		consoleSize = [2]uint{c.ConsoleSize[0], c.ConsoleSize[1]}
	}
	return &container.HostConfig{
		Binds:           c.Binds,
		ContainerIDFile: c.ContainerIDFile,
		LogConfig:       c.LogConfig.build(),
		RestartPolicy:   c.RestartPolicy.build(),
		AutoRemove:      true,
		VolumeDriver:    c.VolumeDriver,
		VolumesFrom:     c.VolumesFrom,
		ConsoleSize:     consoleSize,
		Annotations:     c.Annotations,
		CapAdd:          strslice.StrSlice(c.CapAdd),
		CapDrop:         strslice.StrSlice(c.CapDrop),
		CgroupnsMode:    c.CgroupnsMode,
		DNS:             c.DNS,
		DNSOptions:      c.DNSOptions,
		DNSSearch:       c.DNSSearch,
		ExtraHosts:      c.ExtraHosts,
		GroupAdd:        c.GroupAdd,
		IpcMode:         c.IpcMode,
		Cgroup:          c.Cgroup,
		Links:           c.Links,
		OomScoreAdj:     c.OomScoreAdj,
		PidMode:         c.PidMode,
		Privileged:      c.Privileged,
		PublishAllPorts: c.PublishAllPorts,
		ReadonlyRootfs:  c.ReadonlyRootfs,
		SecurityOpt:     c.SecurityOpt,
		StorageOpt:      c.StorageOpt,
		Tmpfs:           c.Tmpfs,
		UTSMode:         c.UTSMode,
		UsernsMode:      c.UsernsMode,
		ShmSize:         c.ShmSize,
		Sysctls:         c.Sysctls,
		Runtime:         c.Runtime,
		Isolation:       c.Isolation,
		Resources:       c.Resources.build(),
		Mounts:          mapSlice(c.Mounts, Mount.build),
		MaskedPaths:     c.MaskedPaths,
		ReadonlyPaths:   c.ReadonlyPaths,
		Init:            c.Init,
	}
}

func (c *LogConfig) build() container.LogConfig {
	if c == nil {
		return container.LogConfig{}
	}

	return container.LogConfig{
		Type:   c.Type,
		Config: c.Config,
	}
}

func (c *RestartPolicy) build() container.RestartPolicy {
	if c == nil {
		return container.RestartPolicy{}
	}

	return container.RestartPolicy{
		Name:              c.Name,
		MaximumRetryCount: c.MaximumRetryCount,
	}
}

func (c *Resources) build() container.Resources {
	if c == nil {
		return container.Resources{}
	}

	return container.Resources{
		CPUShares:            c.CPUShares,
		Memory:               c.Memory,
		NanoCPUs:             c.NanoCPUs,
		CgroupParent:         c.CgroupParent,
		BlkioWeight:          c.BlkioWeight,
		BlkioWeightDevice:    mapSlice(c.BlkioWeightDevice, WeightDevice.build),
		BlkioDeviceReadBps:   mapSlice(c.BlkioDeviceReadBps, ThrottleDevice.build),
		BlkioDeviceWriteBps:  mapSlice(c.BlkioDeviceWriteBps, ThrottleDevice.build),
		BlkioDeviceReadIOps:  mapSlice(c.BlkioDeviceReadIOps, ThrottleDevice.build),
		BlkioDeviceWriteIOps: mapSlice(c.BlkioDeviceWriteIOps, ThrottleDevice.build),
		CPUPeriod:            c.CPUPeriod,
		CPUQuota:             c.CPUQuota,
		CPURealtimePeriod:    c.CPURealtimePeriod,
		CPURealtimeRuntime:   c.CPURealtimeRuntime,
		CpusetCpus:           c.CpusetCpus,
		CpusetMems:           c.CpusetMems,
		Devices:              mapSlice(c.Devices, DeviceMapping.build),
		DeviceCgroupRules:    c.DeviceCgroupRules,
		DeviceRequests:       mapSlice(c.DeviceRequests, DeviceRequest.build),
		KernelMemoryTCP:      c.KernelMemoryTCP,
		MemoryReservation:    c.MemoryReservation,
		MemorySwap:           c.MemorySwap,
		MemorySwappiness:     c.MemorySwappiness,
		OomKillDisable:       c.OomKillDisable,
		PidsLimit:            c.PidsLimit,
		Ulimits:              mapSlice(c.Ulimits, Ulimit.build),
		CPUCount:             c.CPUCount,
		CPUPercent:           c.CPUPercent,
		IOMaximumIOps:        c.IOMaximumIOps,
		IOMaximumBandwidth:   c.IOMaximumBandwidth,
	}
}

func (c WeightDevice) build() *blkiodev.WeightDevice {
	return &blkiodev.WeightDevice{
		Path:   c.Path,
		Weight: c.Weight,
	}
}

func (c ThrottleDevice) build() *blkiodev.ThrottleDevice {
	return &blkiodev.ThrottleDevice{
		Path: c.Path,
		Rate: c.Rate,
	}
}

func (c DeviceMapping) build() container.DeviceMapping {
	return container.DeviceMapping{
		PathOnHost:        c.PathOnHost,
		PathInContainer:   c.PathInContainer,
		CgroupPermissions: c.CgroupPermissions,
	}
}

func (c DeviceRequest) build() container.DeviceRequest {
	return container.DeviceRequest{
		Driver:       c.Driver,
		Count:        c.Count,
		DeviceIDs:    c.DeviceIDs,
		Capabilities: c.Capabilities,
		Options:      c.Options,
	}
}

func (c Ulimit) build() *units.Ulimit {
	return &units.Ulimit{
		Name: c.Name,
		Hard: c.Hard,
		Soft: c.Soft,
	}
}

func (c Mount) build() mount.Mount {
	return mount.Mount{
		Type:          c.Type,
		Source:        c.Source,
		Target:        c.Target,
		ReadOnly:      c.ReadOnly,
		Consistency:   c.Consistency,
		BindOptions:   c.BindOptions.build(),
		VolumeOptions: c.VolumeOptions.build(),
		TmpfsOptions:  c.TmpfsOptions.build(),
	}
}

func (c *BindOptions) build() *mount.BindOptions {
	if c == nil {
		return nil
	}

	return &mount.BindOptions{
		Propagation:      c.Propagation,
		NonRecursive:     c.NonRecursive,
		CreateMountpoint: c.CreateMountpoint,
	}
}

func (c *VolumeOptions) build() *mount.VolumeOptions {
	if c == nil {
		return nil
	}

	return &mount.VolumeOptions{
		NoCopy:       c.NoCopy,
		Labels:       c.Labels,
		DriverConfig: c.DriverConfig.build(),
	}
}

func (c *Driver) build() *mount.Driver {
	if c == nil {
		return nil
	}

	return &mount.Driver{
		Name:    c.Name,
		Options: c.Options,
	}
}

func (c *TmpfsOptions) build() *mount.TmpfsOptions {
	if c == nil {
		return nil
	}

	return &mount.TmpfsOptions{
		SizeBytes: c.SizeBytes,
		Mode:      c.Mode,
	}
}

func (c *PlatformConfig) build() *ocispec.Platform {
	if c == nil {
		return nil
	}

	return &ocispec.Platform{
		Architecture: c.Architecture,
		OS:           c.OS,
		OSVersion:    c.OSVersion,
		OSFeatures:   c.OSFeatures,
		Variant:      c.Variant,
	}
}

func mapSlice[T1, T2 any](slice []T1, mapper func(T1) T2) []T2 {
	result := make([]T2, len(slice))
	for i, m := range slice {
		result[i] = mapper(m)
	}

	return result
}

type ErrInvalidConfig struct {
	Property string
	Msg      string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid docker config - property %s: %s", e.Property, e.Msg)
}
