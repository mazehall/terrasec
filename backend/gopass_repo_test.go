package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGopass_setConfig(t *testing.T) {
	type fields struct {
		configFile string
		config     GopassHcl
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		state   string
	}{
		{
			name:    "incorrect reading",
			fields:  fields{configFile: "./_nothing"},
			wantErr: true,
			state:   "",
		},
		{
			name:    "first reading",
			fields:  fields{configFile: "../fixture/terrasec.hcl"},
			wantErr: false,
			state:   "aws.private/terrasec.eks-cluster-ci",
		},
		{
			name: "cached reading",
			fields: fields{
				configFile: "../fixture/terrasec.hcl",
				config: GopassHcl{
					Repo: GopassConfig{Kind: "gopass", State: "/foo", Secrets: map[string]string{}},
				},
			},
			wantErr: false,
			state:   "/foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := GopassRepo{
				configFile: tt.fields.configFile,
				synced:     false,
				config:     tt.fields.config,
			}
			if err := r.setConfig(); (err != nil) != tt.wantErr {
				t.Errorf("Gopass.setConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.state, r.config.Repo.State, "they should be equal")
		})
	}
}
