package utils

import (
	"github.com/jinzhu/copier"
	kl "k8s.io/apimachinery/pkg/labels"
)

// IsLabelMatch match resource according to labels
func IsLabelMatch(labelSelector string, labels map[string]string) (bool, error) {
	selector, err := kl.Parse(labelSelector)

	if err != nil {
		return false, err
	}

	labelSet := kl.Set{}
	copier.Copy(&labelSet, &labels)

	return selector.Matches(labelSet), nil
}