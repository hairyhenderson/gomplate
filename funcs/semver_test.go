package funcs

import (
	"context"
	"testing"
)

func TestSemverFuncs_MatchConstraint(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		in         string
		want       bool
		wantErr    bool
	}{
		{
			name:       "mached constraint",
			constraint: ">=1.0.0",
			in:         "v1.1.1",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "not matched constraint",
			constraint: "<1.0.0",
			in:         "v1.1.1",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "wrong constraint",
			constraint: "abc",
			in:         "v1.1.1",
			want:       false,
			wantErr:    true,
		},
		{
			name:       "wrong in",
			constraint: ">1.0.0",
			in:         "va.b.c",
			want:       false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SemverFuncs{
				ctx: context.TODO(),
			}
			got, err := s.MatchConstraint(tt.constraint, tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("SemverFuncs.MatchConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SemverFuncs.MatchConstraint() = %v, want %v", got, tt.want)
			}
		})
	}
}
