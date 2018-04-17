package store

import (
	"reflect"
	"testing"
)

func Test_intersect(t *testing.T) {
	type args struct {
		smallerList   []uint64
		largerListMap map[uint64]struct{}
	}
	tests := []struct {
		name  string
		args  args
		want1 []uint64
		want2 map[uint64]struct{}
	}{
		{
			name:  "Empty sets",
			args:  args{[]uint64{}, map[uint64]struct{}{}},
			want1: []uint64{},
			want2: map[uint64]struct{}{},
		},
		{
			name:  "Empty first set",
			args:  args{[]uint64{}, map[uint64]struct{}{1: struct{}{}}},
			want1: []uint64{},
			want2: map[uint64]struct{}{},
		},
		{
			name:  "Empty second set",
			args:  args{[]uint64{1}, map[uint64]struct{}{}},
			want1: []uint64{},
			want2: map[uint64]struct{}{},
		},
		{
			name:  "Matching",
			args:  args{[]uint64{1}, map[uint64]struct{}{1: struct{}{}}},
			want1: []uint64{1},
			want2: map[uint64]struct{}{1: struct{}{}},
		},
		{
			name:  "Intersecting",
			args:  args{[]uint64{1, 2, 4}, map[uint64]struct{}{2: struct{}{}, 3: struct{}{}, 4: struct{}{}}},
			want1: []uint64{2, 4},
			want2: map[uint64]struct{}{2: struct{}{}, 4: struct{}{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got2 := intersect(tt.args.smallerList, tt.args.largerListMap)
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("intersect() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("intersect() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
