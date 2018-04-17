package store

import (
	"reflect"
	"testing"
)

func Test_uint64ToBytes(t *testing.T) {
	type args struct {
		i uint64
	}
	tests := []struct {
		args args
		want []byte
	}{
		{args{0}, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{args{1}, []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{args{1234567654321234567}, []byte{17, 34, 16, 189, 151, 1, 34, 135}},
	}
	for ti, tt := range tests {
		t.Run(string(ti), func(t *testing.T) {
			if got := uint64ToBytes(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("uint64ToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bytesToUint64(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		args args
		want uint64
	}{
		{args{[]byte{0, 0, 0, 0, 0, 0, 0, 0}}, 0},
		{args{[]byte{0, 0, 0, 0, 0, 0, 0, 1}}, 1},
		{args{[]byte{17, 34, 16, 189, 151, 1, 34, 135}}, 1234567654321234567},
	}
	for ti, tt := range tests {
		t.Run(string(ti), func(t *testing.T) {
			if got := bytesToUint64(tt.args.b); got != tt.want {
				t.Errorf("bytesToUint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_structToBytes(t *testing.T) {
	type args struct {
		s interface{}
	}
	tests := []struct {
		args    args
		want    []byte
		wantErr bool
	}{
		{
			args:    args{struct{ Foo string }{"bar"}},
			want:    []byte{20, 255, 129, 3, 1, 2, 255, 130, 0, 1, 1, 1, 3, 70, 111, 111, 1, 12, 0, 0, 0, 8, 255, 130, 1, 3, 98, 97, 114, 0},
			wantErr: false,
		},
	}
	for ti, tt := range tests {
		t.Run(string(ti), func(t *testing.T) {
			got, err := structToBytes(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("structToBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("structToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
