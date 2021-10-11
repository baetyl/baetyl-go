package task

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

const (
	TaskSuccess = "success"
	TaskFail    = "fail"

	RetryGap    = 10 * time.Millisecond
)

type TaskMessage struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"task"`
	Args     []interface{}          `json:"args"`
	Kwargs   map[string]interface{} `json:"kwargs"`
	Retries  int                    `json:"retries"`
	Expires  *time.Time             `json:"expires"`
}

type ResultMessage struct {
	ID        string        `json:"id"`
	Status    string        `json:"status"`
	Traceback string        `json:"traceback"`
	Result    interface{}   `json:"result"`
}

type BrokerMessage struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type TaskResult struct {
	ID      string
	backend TaskBackend
	result  *ResultMessage
}

// Encode returns base64 json encoded string
func (tm *TaskMessage) Encode() (string, error) {
	jsonData, err := json.Marshal(tm)
	if err != nil {
		return "", err
	}
	encodedData := base64.StdEncoding.EncodeToString(jsonData)
	return encodedData, err
}

// Decode return taskMessage
func (bm *BrokerMessage) Decode() (*TaskMessage, error) {
	body, err := base64.StdEncoding.DecodeString(bm.Value)
	if err != nil {
		return nil, err
	}
	msg := &TaskMessage{}
	err = json.Unmarshal(body, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// Get result synchronize
func (tr *TaskResult) Get(timeout time.Duration) (*ResultMessage, error) {
	ticker := time.NewTicker(RetryGap)
	timeoutChan := time.After(timeout)
	defer ticker.Stop()
	for {
		select {
		case <- timeoutChan:
			return nil, fmt.Errorf("timeout result for %s", tr.ID)
		case <- ticker.C:
			result, err := tr.AsyncGet()
			if err == ErrResultNotFound {
				continue
			}
			return result, nil
		}
	}
}

// AsyncGet result
func (tr *TaskResult) AsyncGet() (*ResultMessage, error) {
	if tr.result != nil {
		return tr.result, nil
	}
	return tr.backend.GetResult(tr.ID)
}
