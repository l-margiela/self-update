package check

import (
	"os"
	"testing"
)

func Test_updateCandidates(t *testing.T) {
	singleDir := []os.FileInfo{
		fileInfo{
			name:  "alice",
			isDir: true,
		},
	}
	sameFile := []os.FileInfo{
		fileInfo{
			name: os.Args[0],
		},
	}
	validFile := []os.FileInfo{
		fileInfo{
			name: "whatever",
			mode: 0111,
		},
	}
	empty := []os.FileInfo{}
	type args struct {
		fs []os.FileInfo
	}
	tests := []struct {
		name string
		args args
		want []os.FileInfo
	}{
		{
			name: "no files",
			args: args{
				fs: empty,
			},
			want: empty,
		},
		{
			name: "directory",
			args: args{
				fs: singleDir,
			},
			want: empty,
		},
		{
			name: "same file",
			args: args{
				fs: sameFile,
			},
			want: empty,
		},
		{
			name: "same file and a directory",
			args: args{
				fs: append(sameFile, singleDir...),
			},
			want: empty,
		},
		{
			name: "same file and a directory",
			args: args{
				fs: append(sameFile, append(validFile, singleDir...)...),
			},
			want: validFile,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateCandidates(tt.args.fs)
			if len(got) != len(tt.want) {
				t.Errorf("filter() = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if tt.want[i] == got[i] {
					continue
				}
				t.Errorf("filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
