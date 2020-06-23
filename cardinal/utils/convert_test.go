package utils

import "testing"

func TestPascalCase(t *testing.T) {
    type args struct {
        underscore string
        lf         []bool
    }
    tests := []struct {
        name string
        args args
        want string
    }{
        {name: "1", args: args{"get_member_by_id", nil}, want: "GetMemberById"},
        {name: "2", args: args{"get_member_by_id", []bool{true}}, want: "getMemberById"},
        {name: "3", args: args{"member_id", nil}, want: "MemberId"},
        {name: "4", args: args{"member_id", []bool{true}}, want: "memberId"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := PascalCase(tt.args.underscore, tt.args.lf...); got != tt.want {
                t.Errorf("CamelCase() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestUnderScoreCase(t *testing.T) {
    type args struct {
        camel string
    }
    tests := []struct {
        name string
        args args
        want string
    }{
        {name: "1", args: args{"GetMemberById"}, want: "get_member_by_id"},
        {name: "2", args: args{"getMemberById"}, want: "get_member_by_id"},
        {name: "3", args: args{"MemberId"}, want: "member_id"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := UnderScoreCase(tt.args.camel); got != tt.want {
                t.Errorf("UnderScoreCase() = %v, want %v", got, tt.want)
            }
        })
    }
}
