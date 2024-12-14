package types

import (
	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
)

type SecretStoreCreateReq struct {
	SecretStoreYaml string `json:"secret_store_yaml"`
}

type SecretStoreCreateResponse struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SecretStoreGetOptions struct {
	ID string
}

type DescribeSecretStoreResponse struct {
	Item    SecretStoreDetail `json:"item"`
	Success bool              `json:"success"`
	Message string            `json:"message"`
}

type SecretStoreHealth struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SecretStoreDetail struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Provider    string            `json:"provider"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	Environment []string          `json:"environment"`
	Path        string            `json:"path,omitempty"`
	LastSynced  string            `json:"lastSynced"`
	CreatedAt   string            `json:"createdAt"`
	LastUpdated string            `json:"lastUpdated"`
	Health      SecretStoreHealth `json:"health"`
}

type ListSecretStoreResponse struct {
	Success bool                `json:"success"`
	Total   int                 `json:"total"`
	Items   []SecretStoreDetail `json:"items"`
	Message string              `json:"message"`
	Error   string              `json:"error,omitempty"`
}

type SecretStoreUpdateRequest struct {
	Name    string                        `json:"name,omitempty"`
	Path    string                        `json:"path,omitempty"`
	Auth    *esv1beta1.VaultAuth          `json:"auth,omitempty"`
	Server  string                        `json:"server,omitempty"`
	Version esv1beta1.VaultKVStoreVersion `json:"version,omitempty"`
}

type SecretStoreUpdateResponse struct {
	Item    SecretStoreDetail `json:"item"`
	Success bool              `json:"success"`
	Message string            `json:"message"`
}

type DeleteSecretStoreResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
