package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/exec"
    "strings"
    "sync"

    "github.com/spf13/pflag"
)

// prints the version message
const version = "0.0.2"

func printVersion() {
	fmt.Printf("Current tldscan version %s\n", version)
}

// Prints the Colorful banner
func printBanner() {
	banner := `
   __   __     __                        
  / /_ / /____/ /_____ _____ ____ _ ____ 
 / __// // __  // ___// ___// __  // __ \
/ /_ / // /_/ /(__  )/ /__ / /_/ // / / /
\__//_/ \__,_//____/ \___/ \__,_//_/ /_/ 
`
fmt.Printf("%s\n%50s\n\n", banner, "Current tldscan version "+version)
}

func main() {
    concurrency := pflag.IntP("concurrency", "c", 50, "Set the concurrency level")
    orgName := pflag.String("org", "", "Organization name to prepend to domains")
    wordlist := pflag.StringP("wordlist", "w", "", "Path to the wordlist file")
    outputFilePath := pflag.StringP("output", "o", "", "File path to save resolved domains")
    silent := pflag.Bool("silent", false, "silent mode.")
    version := pflag.Bool("version", false, "Print the version of the tool and exit.")
    verbose := pflag.BoolP("verbose", "v", false, "Enable verbose output with IP addresses")
    pflag.Parse()

    // Print version and exit if -version flag is provided
	if *version {
		printBanner()
		printVersion()
		return
	}

	// Don't Print banner if -silnet flag is provided
	if !*silent {
		printBanner()
	}

    // Check if tld-small-wordlist.txt exists
    if _, err := os.Stat("tld-small-wordlist.txt"); os.IsNotExist(err) {
        fmt.Println("tld-small-wordlist.txt does not exist. Downloading TLDs...")

        // Download TLDs and generate combinations
        if err := downloadTLDList(); err != nil {
            fmt.Println("Error downloading tld-small-wordlist.txt:", err)
            return
        }
        if err := generateCombinations(); err != nil {
            fmt.Println("Error generating TLD combinations:", err)
            return
        }
    } else {
        fmt.Println("tld-small-wordlist.txt exists. Proceeding with domain resolution...")
    }

    jobs := make(chan string)
    var wg sync.WaitGroup

    // Create output file if specified
    var outputFile *os.File
    if *outputFilePath != "" {
        var err error
        outputFile, err = os.Create(*outputFilePath)
        if err != nil {
            fmt.Fprintf(os.Stderr, "failed to create output file: %s\n", err)
            return
        }
        defer outputFile.Close()
    }

    // Start the domain resolution workers
    for i := 0; i < *concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for domain := range jobs {
                addr, err := net.ResolveIPAddr("ip4", domain)
                if err != nil {
                    continue
                }

                var resolvedDomain string
		        if *verbose {
		            resolvedDomain = fmt.Sprintf("%s -> %s", domain, addr.IP.String()) // Verbose output
		        } else {
		            resolvedDomain = domain // Simple output
		        }

		        // Print the resolved domain to the console
		        fmt.Println(resolvedDomain)

                // If output file is specified, write the result to the file
                if outputFile != nil {
                    _, err := outputFile.WriteString(resolvedDomain + "\n")
                    if err != nil {
                        fmt.Fprintf(os.Stderr, "failed to write to output file: %s\n", err)
                    }
                }
            }
        }()
    }

    // Read from the wordlist or stdin
    if *wordlist != "" {
        file, err := os.Open(*wordlist)
        if err != nil {
            fmt.Fprintf(os.Stderr, "failed to open wordlist file: %s\n", err)
            return
        }
        defer file.Close()

        sc := bufio.NewScanner(file)
        for sc.Scan() {
            domain := sc.Text()
            if *orgName != "" {
                domain = *orgName + domain
            }
            jobs <- domain
        }
        if err := sc.Err(); err != nil {
            fmt.Fprintf(os.Stderr, "failed to read wordlist: %s\n", err)
        }
    } else {
        sc := bufio.NewScanner(os.Stdin)
        for sc.Scan() {
            domain := sc.Text()
            if *orgName != "" {
                domain = *orgName + domain
            }
            jobs <- domain
        }
        if err := sc.Err(); err != nil {
            fmt.Fprintf(os.Stderr, "failed to read input: %s\n", err)
        }
    }

    close(jobs)
    wg.Wait()
}

// Function to download tld-small-wordlist.txt using curl
func downloadTLDList() error {
    cmd := exec.Command("bash", "-c", `curl -s "https://www.iana.org/domains/root/db" | grep '<span class="domain tld"><a href="/domains/root/db/' | grep -oP '\.\w+(?=<\/a>)' | unew -q tld-small-wordlist.txt`)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to run command: %w", err)
    }

    fmt.Println("TLDs saved to tld-small-wordlist.txt")
    return nil
}

// Function to generate TLD combinations
func generateCombinations() error {
    file, err := os.Open("tld-small-wordlist.txt")
    if err != nil {
        return fmt.Errorf("Error opening file: %w", err)
    }
    defer file.Close()

    var tlds []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        tlds = append(tlds, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        return fmt.Errorf("Error reading file: %w", err)
    }

    outputFile := "tld-large-wordlist.txt"
    outfile, err := os.Create(outputFile)
    if err != nil {
        return fmt.Errorf("Error creating output file: %w", err)
    }
    defer outfile.Close()

    var wg sync.WaitGroup
    chunkSize := 1000
    builder := &strings.Builder{}

    for i := 0; i < len(tlds); i++ {
        for j := 0; j < len(tlds); j++ {
            if i != j {
                builder.WriteString(tlds[i] + tlds[j] + "\n")

                if builder.Len() >= chunkSize {
                    wg.Add(1)
                    go func(data string) {
                        defer wg.Done()
                        _, err := outfile.WriteString(data)
                        if err != nil {
                            fmt.Println("Error writing to output file:", err)
                        }
                    }(builder.String())
                    builder.Reset()
                }
            }
        }
    }

    if builder.Len() > 0 {
        wg.Add(1)
        go func(data string) {
            defer wg.Done()
            _, err := outfile.WriteString(builder.String())
            if err != nil {
                fmt.Println("Error writing to output file:", err)
            }
        }(builder.String())
    }

    wg.Wait()

    appendToOutput(outputFile, "tld-small-wordlist.txt")
    fmt.Println("Combinations saved to", outputFile)
    return nil
}

// Function to append contents of one file to another
func appendToOutput(outputFile, inputFile string) {
    file, err := os.Open(inputFile)
    if err != nil {
        fmt.Println("Error opening input file:", err)
        return
    }
    defer file.Close()

    outfile, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("Error opening output file for appending:", err)
        return
    }
    defer outfile.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        _, err := outfile.WriteString(scanner.Text() + "\n")
        if err != nil {
            fmt.Println("Error writing to output file:", err)
            return
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading input file:", err)
    }
}
