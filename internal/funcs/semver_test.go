package funcs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name:       "matched constraint",
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
				ctx: context.Background(),
			}
			got, err := s.CheckConstraint(tt.constraint, tt.in)
			if tt.wantErr {
				require.Errorf(t, err, "SemverFuncs.CheckConstraint() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				require.NoErrorf(t, err, "SemverFuncs.CheckConstraint() error = %v, wantErr %v", err, tt.wantErr)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
