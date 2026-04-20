package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "strings"
    "sync"

    "github.com/rix4uni/tldscan/banner"
    "github.com/spf13/pflag"
)

// getConfigDir returns the config directory for tldscan
func getConfigDir() (string, error) {
    usr, err := user.Current()
    if err != nil {
        return "", err
    }
    configDir := filepath.Join(usr.HomeDir, ".config", "tldscan")
    return configDir, nil
}

// ensureConfigDir creates the config directory if it doesn't exist
func ensureConfigDir() error {
    configDir, err := getConfigDir()
    if err != nil {
        return err
    }
    return os.MkdirAll(configDir, 0755)
}

// getWordlistPath returns the full path for a wordlist
func getWordlistPath(filename string) (string, error) {
    configDir, err := getConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(configDir, filename), nil
}

func main() {
    concurrency := pflag.IntP("concurrency", "c", 50, "Set the concurrency level")
    orgName := pflag.String("org", "", "Organization name to prepend to domains")
    wordlistType := pflag.StringP("wordlist", "w", "small", "Wordlist type to use (small or large)")
    outputFilePath := pflag.StringP("output", "o", "", "File path to save resolved domains")
    silent := pflag.Bool("silent", false, "Silent mode.")
    version := pflag.Bool("version", false, "Print the version of the tool and exit.")
    verbose := pflag.BoolP("verbose", "v", false, "Enable verbose output for debugging purposes.")
    pflag.Parse()

    if *version {
        banner.PrintBanner()
        banner.PrintVersion()
        return
    }

    if !*silent {
        banner.PrintBanner()
    }

    // Ensure config directory exists
    if err := ensureConfigDir(); err != nil {
        fmt.Fprintf(os.Stderr, "failed to create config directory: %s\n", err)
        return
    }

    // Determine which wordlist to use
    var wordlistFile string
    switch *wordlistType {
    case "small":
        wordlistFile = "tld-small-wordlist.txt"
    case "large":
        wordlistFile = "tld-large-wordlist.txt"
    default:
        fmt.Fprintf(os.Stderr, "invalid wordlist type: %s (must be 'small' or 'large')\n", *wordlistType)
        return
    }

    wordlistPath, err := getWordlistPath(wordlistFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to get wordlist path: %s\n", err)
        return
    }

    // Check if the selected wordlist exists, if not, download/generate it
    if _, err := os.Stat(wordlistPath); os.IsNotExist(err) {
        if *verbose {
            fmt.Printf("%s does not exist. Setting up wordlists...\n", wordlistFile)
        }
        // Download TLDs and generate combinations if needed
        if err := downloadTLDList(*verbose); err != nil {
            fmt.Fprintf(os.Stderr, "Error setting up wordlists: %s\n", err)
            return
        }
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

    // Read wordlist into memory
    wordlistFileHandle, err := os.Open(wordlistPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to open wordlist file: %s\n", err)
        return
    }
    defer wordlistFileHandle.Close()

    var tlds []string
    scanner := bufio.NewScanner(wordlistFileHandle)
    for scanner.Scan() {
        tlds = append(tlds, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintf(os.Stderr, "failed to read wordlist: %s\n", err)
        return
    }

    // Check if we're reading from stdin
    stat, _ := os.Stdin.Stat()
    hasStdin := (stat.Mode() & os.ModeCharDevice) == 0

    if hasStdin {
        // Read base domains from stdin and combine with TLDs
        stdinScanner := bufio.NewScanner(os.Stdin)
        for stdinScanner.Scan() {
            baseDomain := strings.TrimSpace(stdinScanner.Text())
            if baseDomain == "" {
                continue
            }
            for _, tld := range tlds {
                domain := baseDomain + tld
                jobs <- domain
            }
        }
        if err := stdinScanner.Err(); err != nil {
            fmt.Fprintf(os.Stderr, "failed to read stdin: %s\n", err)
        }
    } else if *orgName != "" {
        // Use org name with TLDs
        for _, tld := range tlds {
            domain := *orgName + tld
            jobs <- domain
        }
    } else {
        fmt.Fprintf(os.Stderr, "error: either provide --org flag or pipe input via stdin\n")
        fmt.Fprintf(os.Stderr, "usage examples:\n")
        fmt.Fprintf(os.Stderr, "  tldscan --org google --wordlist small\n")
        fmt.Fprintf(os.Stderr, "  echo 'google' | tldscan --wordlist small\n")
        return
    }

    close(jobs)
    wg.Wait()
}

// Function to download tld-small-wordlist.txt using curl
func downloadTLDList(verbose bool) error {
    configDir, err := getConfigDir()
    if err != nil {
        return fmt.Errorf("failed to get config directory: %w", err)
    }

    smallWordlistPath := filepath.Join(configDir, "tld-small-wordlist.txt")

    if verbose {
        fmt.Println("Downloading TLD list...")
    }

    // Download the IANA TLD database page
    cmd := exec.Command("curl", "-s", "https://www.iana.org/domains/root/db")
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to download TLD list: %w", err)
    }

    // Parse TLDs from HTML
    var tlds []string
    content := string(output)

    // Look for TLDs in the HTML pattern: <span class="domain tld"><a href="/domains/root/db/...
    const spanStart = `<span class="domain tld"><a href="/domains/root/db/`
    for {
        idx := strings.Index(content, spanStart)
        if idx == -1 {
            break
        }
        content = content[idx+len(spanStart):]

        // Find the next '>' which ends the <a> tag
        gtIdx := strings.Index(content, ">")
        if gtIdx == -1 {
            continue
        }

        // Extract TLD (starts with .)
        content = content[gtIdx+1:]
        ltIdx := strings.Index(content, "<")
        if ltIdx == -1 {
            continue
        }

        tld := strings.TrimSpace(content[:ltIdx])
        if strings.HasPrefix(tld, ".") && tld != "." {
            tlds = append(tlds, tld)
        }
    }

    if len(tlds) == 0 {
        return fmt.Errorf("no TLDs found in downloaded content")
    }

    // Write TLDs to file
    file, err := os.Create(smallWordlistPath)
    if err != nil {
        return fmt.Errorf("failed to create wordlist file: %w", err)
    }
    defer file.Close()

    for _, tld := range tlds {
        _, err := file.WriteString(tld + "\n")
        if err != nil {
            return fmt.Errorf("failed to write to wordlist file: %w", err)
        }
    }

    if verbose {
        fmt.Printf("TLDs saved to %s (%d TLDs)\n", smallWordlistPath, len(tlds))
    }

    // Generate large wordlist after downloading small one
    if err := generateCombinations(verbose); err != nil {
        return fmt.Errorf("failed to generate combinations: %w", err)
    }

    return nil
}

// Function to generate TLD combinations
func generateCombinations(verbose bool) error {
    configDir, err := getConfigDir()
    if err != nil {
        return fmt.Errorf("failed to get config directory: %w", err)
    }

    smallWordlistPath := filepath.Join(configDir, "tld-small-wordlist.txt")
    largeWordlistPath := filepath.Join(configDir, "tld-large-wordlist.txt")

    file, err := os.Open(smallWordlistPath)
    if err != nil {
        return fmt.Errorf("error opening file: %w", err)
    }
    defer file.Close()

    var tlds []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        tlds = append(tlds, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error reading file: %w", err)
    }

    outfile, err := os.Create(largeWordlistPath)
    if err != nil {
        return fmt.Errorf("error creating output file: %w", err)
    }
    defer outfile.Close()

    var wg sync.WaitGroup
    chunkSize := 1000
    builder := &strings.Builder{}

    if verbose {
        fmt.Println("Generating TLD combinations...")
    }

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

    // Append small wordlist to large wordlist
    appendToOutput(largeWordlistPath, smallWordlistPath)
    
    if verbose {
        fmt.Printf("Combinations saved to %s\n", largeWordlistPath)
    }
    
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