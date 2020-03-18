package v1

import "time"

// ReportSpec report spec
type ReportSpec struct {
	Time        time.Time   `json:"time,omitempty"`
	Node        NodeInfo    `json:"node,omitempty"`
	NodeStats   NodeStatus  `json:"nodestats,omitempty"`
	AppVersions AppVersions `json:"apps,omitempty"`
	AppStats    []AppStatus `json:"appstats,omitempty"`
}

// NodeInfo node info
type NodeInfo struct {
}

// NodeStatus node status
type NodeStatus struct {
}

// AppVersions app versions
type AppVersions map[string]string

// AppStatus app status
type AppStatus struct {
}

// DesireSpec desire spec
type DesireSpec struct {
}
