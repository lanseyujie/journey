package utils

import (
    "github.com/lanseyujie/journey/log"
    "testing"
)

func TestGetFunctionName(t *testing.T) {
    type args struct {
        fn   interface{}
        seps []rune
    }
    tests := []struct {
        name string
        args args
        want string
    }{
        {name: "1", args: args{fn: log.Println, seps: []rune{'/', '.'}}, want: "Println"},
        {name: "2", args: args{fn: func() {}, seps: []rune{'/', '.'}}, want: "func1"},
        {name: "3", args: args{fn: func() {}, seps: []rune{'/', '.'}}, want: "func2"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := GetFunctionName(tt.args.fn, tt.args.seps...); got != tt.want {
                t.Errorf("GetFunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
