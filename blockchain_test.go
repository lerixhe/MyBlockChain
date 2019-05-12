//定义链

package main

import (
	"github.com/boltdb/bolt"
	"reflect"
	"testing"
)

func TestNewBlockChain(t *testing.T) {
	tests := []struct {
		name string
		want *BlockChain
	}{
		// TODO: Add test cases.

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockChain(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockChain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_AddBlock(t *testing.T) {
	mydb, _ := bolt.Open(dbFile, 0600, nil)
	type args struct {
		data string
	}
	tests := []struct {
		name string
		bc   *BlockChain
		args args
	}{
		// TODO: Add test cases.
		{
			name: "case",
			bc: &BlockChain{
				db:   mydb,
				tail: []byte("0000010000000000000000000000000000000000000000000000000000000000"),
			},
			args: args{
				"lalalla",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.bc.AddBlock(tt.args.data)
		})
	}
}
