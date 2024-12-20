package svg

import (
	"testing"

	mt "github.com/rustyoz/Mtransform"
)

func TestPolyLine(t *testing.T) {
	tests := []struct {
		name      string
		polyline  PolyLine
		wantValid bool
	}{
		{
			name: "valid polyline",
			polyline: PolyLine{
				ID:        "test-polyline",
				Transform: "translate(10,20)",
				Style:     "fill:none;stroke:black",
				Points:    "0,0 10,10 20,20",
			},
			wantValid: true,
		},
		{
			name: "empty polyline",
			polyline: PolyLine{
				ID: "empty-polyline",
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic structure validation
			if tt.polyline.ID == "" && tt.wantValid {
				t.Error("expected non-empty ID for valid polyline")
			}

			// Test if Points string is properly formatted (when present)
			if tt.polyline.Points != "" {
				// Could add more specific point format validation here
				if len(tt.polyline.Points) < 3 { // At least one point (x,y)
					t.Error("invalid points format")
				}
			}

			// Test Transform parsing (when present)
			if tt.polyline.Transform != "" {
				transform, err := parseTransform(tt.polyline.Transform)
				if err != nil {
					t.Errorf("failed to parse transform: %v", err)
				}
				tt.polyline.transform = transform
				t.Logf("transform: %v", transform)
				ttransform := mt.Identity()
				ttransform.Translate(10, 20)
				if !transform.Equals(&ttransform) {
					t.Errorf("transforms do not match")
				}
			}
		})
	}
}
