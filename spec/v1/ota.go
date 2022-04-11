package v1

type OtaInfo struct {
	ApkInfo     *ApkInfo        `json:"apkInfo,omitempty"`
	DeviceGroup *OtaDeviceGroup `json:"deviceGroup,omitempty"`
	Task        *OtaTask        `json:"task,omitempty"`
}

type ApkInfo struct {
	Key         string `json:"key,omitempty"`
	Url         string `json:"url,omitempty"`
	Md5         string `json:"md5,omitempty"`
	Sha1        string `json:"sha1,omitempty"`
	PackName    string `json:"packName,omitempty"`
	Size        int64  `json:"size,omitempty"`
	VersionCode int    `json:"versionCode,omitempty"`
	Version     string `json:"version"`
}

type OtaDeviceGroup struct {
	AddTestDeviceGroupName string `json:"addTestDeviceGroupName,omitempty"`
	AddTestDeviceGroupId   int    `json:"addTestDeviceGroupId,omitempty"`
	AddTestDeviceGroupKey  string `json:"addTestDeviceGroupKey,omitempty"`
	AddDeviceGroupName     string `json:"addDeviceGroupName,omitempty"`
	AddDeviceGroupId       int    `json:"addDeviceGroupId,omitempty"`
	AddDeviceGroupKey      string `json:"addDeviceGroupKey,omitempty"`
	DelTestDeviceGroupName string `json:"delTestDeviceGroupName,omitempty"`
	DelTestDeviceGroupId   int    `json:"delTestDeviceGroupId,omitempty"`
	DelTestDeviceGroupKey  string `json:"delTestDeviceGroupKey,omitempty"`
	DelDeviceGroupName     string `json:"delDeviceGroupName,omitempty"`
	DelDeviceGroupId       int    `json:"delDeviceGroupId,omitempty"`
	DelDeviceGroupKey      string `json:"delDeviceGroupKey,omitempty"`
}

type OtaTask struct {
	AddTaskName string `json:"addTaskName,omitempty"`
	AddTaskId   int    `json:"addTaskId,omitempty"`
	DelTaskName string `json:"delTaskName,omitempty"`
	DelTaskId   int    `json:"delTaskId,omitempty"`
}
