## tldscan

Scan all possible TLD's for a given domain name

## Installation
```
go install github.com/rix4uni/tldscan@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/tldscan/releases/download/v0.0.2/tldscan-linux-amd64-0.0.2.tgz
tar -xvzf tldscan-linux-amd64-0.0.2.tgz
rm -rf tldscan-linux-amd64-0.0.2.tgz
mv tldscan ~/go/bin/tldscan
```
Or download [binary release](https://github.com/rix4uni/tldscan/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/tldscan.git
cd tldscan; go install
```

## Usage
```
Usage of tldscan:
  -c, --concurrency int   Set the concurrency level (default 50)
      --org string        Organization name to prepend to domains
  -o, --output string     File path to save resolved domains
      --silent            silent mode.
  -v, --verbose           Enable verbose output with IP addresses
      --version           Print the version of the tool and exit.
  -w, --wordlist string   Path to the wordlist file
```

## Usage Examples
```bash
# Fast with normal small wordlist
$ tldscan --org google -w tld-small-wordlist.txt -o tldscan-output.txt
google.org
google.ac
google.ad
google.ae

# Slow with large combinations wordlist
$ tldscan --org google -w tld-large-wordlist.txt -o tldscan-output.txt
google.org.ag
google.org.amsterdam
google.org.app
google.org.arab
google.org.aw
google.org.best
```
