package errors

import (
	"fmt"
	"testing"
)

func TestIsError(t *testing.T) {
	t.Parallel()
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestIsError",
			args: args{
				err: &Error{},
			},
			want: true,
		},
		{
			name: "TestIsError",
			args: args{
				err: fmt.Errorf("TestError"),
			},
			want: false,
		},
		{
			name: "TestIsError",
			args: args{
				err: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsError(tt.args.err); got != tt.want {
				t.Errorf("IsError() = %v, want %v", got, tt.want)
			}
		})
	}
}
