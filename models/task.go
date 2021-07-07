package models

type Arg struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type Task struct {
	Id          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	RetryTimes  int       `json:"retrytimes,omitempty"`
	Namespace   string    `json:"namespace,omitempty"`
	JobName     string    `json:"jobName,omitempty"`
	Args        []Arg     `json:"args,omitempty"`
	Async       bool      `json:"async,omitempty"`
}
