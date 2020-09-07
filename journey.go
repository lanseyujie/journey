package journey

import "fmt"

const (
    banner = `
       __
      / /___  __  ___________  ___  __  __
 __  / / __ \/ / / / ___/ __ \/ _ \/ / / /
/ /_/ / /_/ / /_/ / /  / / / /  __/ /_/ /
\____/\____/\__,_/_/  /_/ /_/\___/\__, /
                                 /____/    v%s (%s)

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
