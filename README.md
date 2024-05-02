# tldscan

## Installation
```
go install -v github.com/rix4uni/unew@latest
go install github.com/tomnomnom/hacks/filter-resolved@latest

git clone https://github.com/rix4uni/tldscan.git
cd tldscan && chmod +x tldscan && mv tldscan /usr/bin/
tldscan -h
```

## Usage
```
Quick-Mode Run:
   tldscan -q google

Verbose-Mode Run:
   tldscan -v google

Show Help:
   tldscan -h
```

## Output
```
#tldscan -q google
google.org
google.ac
google.ad
google.ae

#tldscan -v google

```
