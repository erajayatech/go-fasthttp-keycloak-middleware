package keycloakmiddleware

import (
	"testing"
)

func Test_isScopesValid(t *testing.T) {
	type args struct {
		claims claims
		scopes []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "scope valid",
			args: args{
				claims: claims{
					Authorization: authorization{
						Permissions: []permission{
							{Scopes: []string{"foo", "bar", "baz"}},
							{Scopes: []string{"qux", "fred"}},
						},
					},
				},
				scopes: []string{"bar", "fred"},
			},
			want: true,
		},
		{
			name: "scope valid 2",
			args: args{
				claims: claims{
					Authorization: authorization{
						Permissions: []permission{
							{Scopes: []string{"foo", "bar", "baz"}},
							{Scopes: []string{"qux", "fred"}},
						},
					},
				},
				scopes: []string{"qux", "thud"},
			},
			want: true,
		},
		{
			name: "scope valid 2",
			args: args{
				claims: claims{
					Authorization: authorization{
						Permissions: []permission{
							{Scopes: []string{"foo", "bar", "baz"}},
							{Scopes: []string{"qux", "fred"}},
						},
					},
				},
				scopes: []string{"thud", "chips"},
			},
			want: false,
		},
	}
	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			if got := isScopesValid(tests[i].args.claims, tests[i].args.scopes); got != tests[i].want {
				t.Errorf("isScopesValid() = %v, want %v", got, tests[i].want)
			}
		})
	}
}
