package kube

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		mode           string
		kubeconfigPath string
		name           string
		wantErr        bool
	}{
		{
			name:           "test new client",
			mode:           "kubeconfig",
			kubeconfigPath: "/Users/guohao/.kube/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("kubernetes.mode", tt.mode)
			viper.Set("kubernetes.kubeconfig.path", tt.kubeconfigPath)
			got, err := NewClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)
		})
	}
}
