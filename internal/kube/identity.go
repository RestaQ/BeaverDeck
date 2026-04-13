package kube

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) NamespaceUID(ctx context.Context, namespace string) (string, error) {
	obj, err := c.core.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get namespace UID: %w", err)
	}
	return string(obj.UID), nil
}

func (c *Client) ServiceAccountUID(ctx context.Context, namespace, name string) (string, error) {
	obj, err := c.core.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get service account UID: %w", err)
	}
	return string(obj.UID), nil
}
