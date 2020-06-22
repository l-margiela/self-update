package check

import (
	"testing"

	"github.com/Masterminds/semver"
)

func Test_isNewer(t *testing.T) {
	type args struct {
		new  *semver.Version
		curr *semver.Version
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "same version",
			args: args{
				new:  semver.MustParse("1.0.0"),
				curr: semver.MustParse("1.0.0"),
			},
			want: false,
		},
		{
			name: "curr newer",
			args: args{
				new:  semver.MustParse("1.0.0"),
				curr: semver.MustParse("1.1.0"),
			},
			want: false,
		},
		{
			name: "new newer",
			args: args{
				new:  semver.MustParse("1.1.0"),
				curr: semver.MustParse("1.0.0"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNewer(tt.args.new, tt.args.curr); got != tt.want {
				t.Errorf("isNewer() = %v, want %v", got, tt.want)
			}
		})
	}
}
