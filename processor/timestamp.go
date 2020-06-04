package processor

import "time"

// parseLastRollingRestartTimeBestEffort returns current time if we fail to parse,
// which is a failing-safe behaviour to ensure we don't repeatedly restart a pod
// controller while inserting invalid timestamps. This is to be placed on pod
// controllers as order.kube-system.com/last-rolling-restart
func parseLastRollingRestartTimeBestEffort(labelValue string) time.Time {
	t, err := time.Parse(time.RFC3339, labelValue)
	if err != nil {
		return time.Now()
	}

	return t
}
