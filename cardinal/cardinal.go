package cardinal

import "fmt"

const (
    banner = `
     ______               ___             __
    / ____/___  ____  ___/ (_)___  ____  / /
   / /   / __ \/ ___\/ __ / / __ \/ __ \/ /
  / /___/ /_/ / /  / /_/ / / / / / /_/ / /
  \____/\__,_/_/   \__,_/_/_/ /_/\__,_/_/    v%s (%s)
high performance, minimalist blog framework
https://lanseyujie.com

`
)

var (
    Version    = "1.0.0"
    LastCommit = "0000000"
    BuildDate  = ""
)

func init() {
    fmt.Printf(banner, Version, LastCommit)
}
