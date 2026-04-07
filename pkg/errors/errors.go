package errors

import (
	"fmt"
)

// DomainError represents a domain-specific error
type DomainError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new domain error
func NewDomainError(message, code string) *DomainError {
	return &DomainError{
		Message: message,
		Code:    code,
	}
}

// NewDomainErrorWithCause creates a new domain error with a cause
func NewDomainErrorWithCause(message, code string, cause error) *DomainError {
	return &DomainError{
		Message: message,
		Code:    code,
		Err:     cause,
	}
}

// ApplicationError represents an application-level error
type ApplicationError struct {
	Message string         `json:"message"`
	Code    string         `json:"code"`
	Details map[string]any `json:"details,omitempty"`
	Err     error          `json:"-"`
}

// Error implements the error interface
func (e *ApplicationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *ApplicationError) Unwrap() error {
	return e.Err
}

// NewApplicationError creates a new application error
func NewApplicationError(message, code string) *ApplicationError {
	return &ApplicationError{
		Message: message,
		Code:    code,
	}
}

// NewApplicationErrorWithDetails creates a new application error with details
func NewApplicationErrorWithDetails(message, code string, details map[string]any) *ApplicationError {
	return &ApplicationError{
		Message: message,
		Code:    code,
		Details: details,
	}
}

// InfrastructureError represents an infrastructure-level error
type InfrastructureError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *InfrastructureError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *InfrastructureError) Unwrap() error {
	return e.Err
}

// NewInfrastructureError creates a new infrastructure error
func NewInfrastructureError(message, code string) *InfrastructureError {
	return &InfrastructureError{
		Message: message,
		Code:    code,
	}
}

// Common error codes
const (
	// Container errors
	ErrContainerIDRequired        = "CONTAINER_ID_REQUIRED"
	ErrContainerNameRequired      = "CONTAINER_NAME_REQUIRED"
	ErrContainerNamespaceRequired = "CONTAINER_NAMESPACE_REQUIRED"

	// Workload errors
	ErrWorkloadIDRequired        = "WORKLOAD_ID_REQUIRED"
	ErrWorkloadNameRequired      = "WORKLOAD_NAME_REQUIRED"
	ErrWorkloadNamespaceRequired = "WORKLOAD_NAMESPACE_REQUIRED"

	// NodePool errors
	ErrNodePoolIDRequired        = "NODEPOOL_ID_REQUIRED"
	ErrNodePoolNameRequired      = "NODEPOOL_NAME_REQUIRED"
	ErrNodePoolClusterIDRequired = "NODEPOOL_CLUSTER_ID_REQUIRED"

	// Cluster errors
	ErrClusterIDRequired         = "CLUSTER_ID_REQUIRED"
	ErrClusterNameRequired       = "CLUSTER_NAME_REQUIRED"
	ErrClusterPrometheusRequired = "CLUSTER_PROMETHEUS_REQUIRED"

	// Application errors
	ErrClusterNotFound   = "CLUSTER_NOT_FOUND"
	ErrNamespaceNotFound = "NAMESPACE_NOT_FOUND"
	ErrWorkloadNotFound  = "WORKLOAD_NOT_FOUND"
	ErrContainerNotFound = "CONTAINER_NOT_FOUND"
	ErrInvalidInput      = "INVALID_INPUT"
	ErrProcessingFailed  = "PROCESSING_FAILED"

	// Infrastructure errors
	ErrPrometheusQueryFailed = "PROMETHEUS_QUERY_FAILED"
	ErrKubernetesAPIError    = "KUBERNETES_API_ERROR"
	ErrBillingAPIError       = "BILLING_API_ERROR"
	ErrDatabaseError         = "DATABASE_ERROR"
)
