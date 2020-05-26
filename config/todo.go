package config

// Parse config from a YAML ConfigMap (which orderrrr can use to manage orderrrr!)
// specifying which Secrets and ConfigMaps affect which resource controllers,
// and what cooldown they all have for rolling restarts. We won't attempt to
// parse pod controllers (Deployments, DaemonSets, ReplicaSets, StatefulSets,
// CronJobs/Jobs) as determining mounting and reference relationships can
// be challenging in the Kubernetes API schema, and globally restarting all
// resources at once when any referenced or mounted resource they share
// changes may not necessarily be the desired behaviour, especially in a
// large cluster.
