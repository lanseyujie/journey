package utils

import (
    "testing"
)

func TestNewUuidV4(t *testing.T) {
    zero := "00000000-0000-4000-8000-000000000000"
    u1 := NewUuidV4().String()
    u2 := NewUuidV4().String()
    if u1 == zero || u2 == zero || u1 == u2 {
        t.FailNow()
    }
}

func TestNewUuidV5(t *testing.T) {
    type args struct {
        namespace []byte
        name      []byte
    }
    tests := []struct {
        name    string
        args    args
        want    string
        wantErr bool
    }{
        {name: "1", args: args{namespace: []byte("test"), name: []byte("1")}, want: "b444ac06-613f-58d6-b795-be9ad0beaf55", wantErr: false},
        {name: "2", args: args{namespace: []byte("test"), name: []byte("2")}, want: "109f4b3c-50d7-50df-b29d-299bc6f8e9ef", wantErr: false},
        {name: "3", args: args{namespace: []byte("dev"), name: []byte("2")}, want: "87835653-f7f7-5d15-820f-c69988bdadcd", wantErr: false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewUuidV5(tt.args.namespace, tt.args.name)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewUuidV5() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if err == nil && got.String() != tt.want {
                t.Errorf("NewUuidV5() got = %v, want %v", got, tt.want)
            }
        })
    }
}
