package tenant

import (
	"testing"

	"github.com/joshbarros/golang-carflow-api/internal/domain"
)

func TestIsValidPlan(t *testing.T) {
	tests := []struct {
		name string
		plan string
		want bool
	}{
		{"Basic plan", domain.PlanBasic, true},
		{"Pro plan", domain.PlanPro, true},
		{"Enterprise plan", domain.PlanEnterprise, true},
		{"Invalid plan", "invalid", false},
		{"Empty plan", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.IsValidPlan(tt.plan); got != tt.want {
				t.Errorf("IsValidPlan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"Active status", domain.StatusActive, true},
		{"Inactive status", domain.StatusInactive, true},
		{"Suspended status", domain.StatusSuspended, true},
		{"Invalid status", "invalid", false},
		{"Empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.IsValidStatus(tt.status); got != tt.want {
				t.Errorf("IsValidStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{"Valid domain", "example.com", false},
		{"Valid subdomain", "sub.example.com", false},
		{"Invalid domain - no TLD", "example", true},
		{"Invalid domain - special chars", "example!.com", true},
		{"Invalid domain - spaces", "example .com", true},
		{"Empty domain", "", true},
		{"Too long domain", "a.example.com", false},
		{"IP address", "192.168.1.1", true},
		{"Localhost", "localhost", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateDomain(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDomain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDefaultResourceLimits(t *testing.T) {
	tests := []struct {
		name string
		plan string
		want domain.ResourceLimits
	}{
		{
			name: "Basic plan limits",
			plan: domain.PlanBasic,
			want: domain.ResourceLimits{
				MaxUsers:        10,
				MaxCars:         50,
				APIRateLimit:    100,
				StorageLimit:    1024,
				BackupRetention: 7,
			},
		},
		{
			name: "Pro plan limits",
			plan: domain.PlanPro,
			want: domain.ResourceLimits{
				MaxUsers:        50,
				MaxCars:         200,
				APIRateLimit:    500,
				StorageLimit:    5120,
				BackupRetention: 30,
			},
		},
		{
			name: "Enterprise plan limits",
			plan: domain.PlanEnterprise,
			want: domain.ResourceLimits{
				MaxUsers:        -1,
				MaxCars:         -1,
				APIRateLimit:    1000,
				StorageLimit:    -1,
				BackupRetention: 90,
			},
		},
		{
			name: "Invalid plan defaults to basic",
			plan: "invalid",
			want: domain.ResourceLimits{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.GetDefaultResourceLimits(tt.plan)
			if got != tt.want {
				t.Errorf("GetDefaultResourceLimits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDefaultFeatures(t *testing.T) {
	baseFeatures := []domain.Feature{
		{Name: "basic_reporting", Description: "Basic reporting and analytics", Enabled: true},
		{Name: "email_support", Description: "Email support during business hours", Enabled: true},
	}

	proFeatures := []domain.Feature{
		{Name: "advanced_reporting", Description: "Advanced reporting and analytics", Enabled: true},
		{Name: "priority_support", Description: "Priority email and phone support", Enabled: true},
		{Name: "api_access", Description: "API access for integration", Enabled: true},
	}

	enterpriseFeatures := []domain.Feature{
		{Name: "custom_branding", Description: "Custom branding and white-labeling", Enabled: true},
		{Name: "dedicated_support", Description: "24/7 dedicated support", Enabled: true},
		{Name: "audit_logs", Description: "Detailed audit logs", Enabled: true},
	}

	tests := []struct {
		name     string
		plan     string
		wantLen  int
		features []domain.Feature
	}{
		{
			name:     "Basic plan features",
			plan:     domain.PlanBasic,
			wantLen:  len(baseFeatures),
			features: baseFeatures,
		},
		{
			name:     "Pro plan features",
			plan:     domain.PlanPro,
			wantLen:  len(baseFeatures) + len(proFeatures),
			features: append(baseFeatures, proFeatures...),
		},
		{
			name:     "Enterprise plan features",
			plan:     domain.PlanEnterprise,
			wantLen:  len(baseFeatures) + len(proFeatures) + len(enterpriseFeatures),
			features: append(append(baseFeatures, proFeatures...), enterpriseFeatures...),
		},
		{
			name:     "Invalid plan defaults to basic",
			plan:     "invalid",
			wantLen:  0,
			features: []domain.Feature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.GetDefaultFeatures(tt.plan)
			if len(got) != tt.wantLen {
				t.Errorf("GetDefaultFeatures() returned %d features, want %d", len(got), tt.wantLen)
			}

			// Check if all expected features are present
			for i, feature := range tt.features {
				if got[i] != feature {
					t.Errorf("GetDefaultFeatures()[%d] = %v, want %v", i, got[i], feature)
				}
			}
		})
	}
}
