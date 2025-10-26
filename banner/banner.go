package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.3"

func PrintVersion() {
	fmt.Printf("Current tldscan version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `
   __   __     __                        
  / /_ / /____/ /_____ _____ ____ _ ____ 
 / __// // __  // ___// ___// __  // __ \
/ /_ / // /_/ /(__  )/ /__ / /_/ // / / /
\__//_/ \__,_//____/ \___/ \__,_//_/ /_/ 
`
    fmt.Printf("%s\n%50s\n\n", banner, "Current tldscan version "+version)
}
