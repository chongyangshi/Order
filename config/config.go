package config

import (
	"fmt"
	"io/ioutil"

	"github.com/icydoge/orderrrr/proto"

	yaml "gopkg.in/yaml.v2"
)

// CurrentVersion represents the current version number of orderrrr
const CurrentVersion = 0.1

// IsVersionCompatible checks whether a given config version is compatible
func IsVersionCompatible(version float64) bool {
	return version == CurrentVersion
}

// Parse config from a YAML ConfigMap (which orderrrr can use to manage orderrrr!)
// specifying which Secrets and ConfigMaps affect which resource controllers,
// and what cooldown they all have for rolling restarts. We won't attempt to
// parse pod controllers (Deployments, DaemonSets, ReplicaSets, StatefulSets,
// CronJobs/Jobs) as determining mounting and reference relationships can
// be challenging in the Kubernetes API schema, and globally restarting all
// resources at once when any referenced or mounted resource they share
// changes may not necessarily be the desired behaviour, especially in a
// large cluster.

// Config is a global state storing the runtime config, which is read only.
// However, orderrrr can monitor and restart itself if its config changes in
// the cluster.
var Config proto.OrderrrrConfig

// LoadConfig loads the runtime configurations, it should be called before
// controllers are started.
func LoadConfig(configPath string) error {
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Error reading config from %s: %v", configPath, err)
	}

	Config = OrderrrrConfig{}
	err = yaml.Unmarshal(configBytes, &Config)
	if err == nil {
		return fmt.Errorf("Got parsing config from %s: %v", configPath, err)
	}

	return nil
}
