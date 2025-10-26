## tldscan

A high-performance domain scanner that discovers active domains by testing multiple Top-Level Domains (TLDs) for given domain names.

## Installation
```
go install github.com/rix4uni/tldscan@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/tldscan/releases/download/v0.0.3/tldscan-linux-amd64-0.0.3.tgz
tar -xvzf tldscan-linux-amd64-0.0.3.tgz
rm -rf tldscan-linux-amd64-0.0.3.tgz
mv tldscan ~/go/bin/tldscan
```
Or download [binary release](https://github.com/rix4uni/tldscan/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/tldscan.git
cd tldscan; go install
```

## Usage
```yaml
Usage of tldscan:
  -c, --concurrency int   Set the concurrency level (default 50)
      --org string        Organization name to prepend to domains
  -o, --output string     File path to save resolved domains
      --silent            Silent mode.
  -v, --verbose           Enable verbose output for debugging purposes.
      --version           Print the version of the tool and exit.
  -w, --wordlist string   Wordlist type to use (small or large) (default "small")
```

## Usage Examples

Single URL:
```yaml
# Using stdin with small wordlist
$ echo "google" | tldscan --wordlist small

google.org
google.ac
google.ad
google.ae

# Using stdin with large wordlist
$ echo "google" | tldscan --wordlist large
google.org.ag
google.org.amsterdam
google.org.app
google.org.arab
google.org.aw
google.org.best
```

Multiple URLs:
```yaml
$ cat org.txt | tldscan --wordlist small

google.org
google.ac
google.ad
google.ae
dell.org
dell.ac
dell.ad
dell.ae
```
