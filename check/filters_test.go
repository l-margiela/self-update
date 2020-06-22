package check

import (
	"os"
	"testing"
	"time"
)

type fileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	isDir   bool
	modTime time.Time
	sys     interface{}
}

func (f fileInfo) Name() string {
	return f.name
}

func (f fileInfo) Size() int64 {
	return f.size
}

func (f fileInfo) Mode() os.FileMode {
	return f.mode
}

func (f fileInfo) IsDir() bool {
	return f.isDir
}

func (f fileInfo) ModTime() time.Time {
	return f.modTime
}

func (f fileInfo) Sys() interface{} {
	return f.sys
}

func Test_executableFilter(t *testing.T) {
	type args struct {
		f os.FileInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "zeros",
			args: args{
				f: fileInfo{
					mode: 0,
				},
			},
			want: false,
		},
		{
			name: "ones",
			args: args{
				f: fileInfo{
					mode: 0111,
				},
			},
			want: true,
		},
		{
			name: "user executable",
			args: args{
				f: fileInfo{
					mode: 0100,
				},
			},
			want: true,
		},
		{
			name: "group executable",
			args: args{
				f: fileInfo{
					mode: 0010,
				},
			},
			want: true,
		},
		{
			name: "everyone executable",
			args: args{
				f: fileInfo{
					mode: 0001,
				},
			},
			want: true,
		},
		{
			name: "777",
			args: args{
				f: fileInfo{
					mode: 0777,
				},
			},
			want: true,
		},
		{
			name: "666",
			args: args{
				f: fileInfo{
					mode: 0666,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := executableFilter(tt.args.f); got != tt.want {
				t.Errorf("executableFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sameFileFilter(t *testing.T) {
	type args struct {
		f os.FileInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "same",
			args: args{
				f: fileInfo{
					name: os.Args[0],
				},
			},
			want: false,
		},
		{
			name: "different",
			args: args{
				f: fileInfo{
					name: "whatever",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sameFileFilter(tt.args.f); got != tt.want {
				t.Errorf("sameFileFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dirFilter(t *testing.T) {
	type args struct {
		f os.FileInfo
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "directory",
			args: args{
				f: fileInfo{
					isDir: true,
				},
			},
			want: false,
		},
		{
			name: "not directory",
			args: args{
				f: fileInfo{
					isDir: false,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dirFilter(tt.args.f); got != tt.want {
				t.Errorf("dirFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filter(t *testing.T) {
	sampleFs := []os.FileInfo{fileInfo{name: "bob"}, fileInfo{name: "alice"}}
	empty := []os.FileInfo{}
	type args struct {
		fs   []os.FileInfo
		pred filterPredicate
	}
	tests := []struct {
		name string
		args args
		want []os.FileInfo
	}{
		{
			name: "empty list always true predicate",
			args: args{
				fs:   empty,
				pred: func(f os.FileInfo) bool { return true },
			},
			want: empty,
		},
		{
			name: "empty list always false predicate",
			args: args{
				fs:   empty,
				pred: func(f os.FileInfo) bool { return false },
			},
			want: empty,
		},
		{
			name: "always true predicate",
			args: args{
				fs:   sampleFs,
				pred: func(f os.FileInfo) bool { return true },
			},
			want: sampleFs,
		},
		{
			name: "always false predicate",
			args: args{
				fs:   sampleFs,
				pred: func(f os.FileInfo) bool { return false },
			},
			want: empty,
		},
		{
			name: "only alice predicate",
			args: args{
				fs:   sampleFs,
				pred: func(f os.FileInfo) bool { return f.Name() == "alice" },
			},
			want: []os.FileInfo{fileInfo{name: "alice"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter(tt.args.fs, tt.args.pred)
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
