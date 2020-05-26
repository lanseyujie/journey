package cardinal

import "fmt"

const (
    version = "1.0.0"
    website = "https://lanseyujie.com"
    banner  = `
     ______               ___             __
    / ____/___  ____  ___/ (_)___  ____  / /
   / /   / __ \/ ___\/ __ / / __ \/ __ \/ /
  / /___/ /_/ / /  / /_/ / / / / / /_/ / /
  \____/\__,_/_/   \__,_/_/_/ /_/\__,_/_/    %s
high performance, minimalist blog framework
%s

`
)

func init() {
    fmt.Printf(banner, version, website)
}
