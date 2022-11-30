[![Go Report Card](https://goreportcard.com/badge/github.com/go-ext/syncsafe)](https://goreportcard.com/report/github.com/go-ext/syncsafe)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/eff211f4bfa14af0ac69c8e0c08f1c90)](https://www.codacy.com/gh/go-ext/syncsafe/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=go-ext/syncsafe&amp;utm_campaign=Badge_Grade)
![example workflow](https://github.com/go-ext/syncsafe/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/go-ext/syncsafe/branch/main/graph/badge.svg?token=ZNB6FL3YOD)](https://codecov.io/gh/go-ext/syncsafe)
[![GoDoc](https://godoc.org/github.com/askretov/timex?status.svg)](https://godoc.org/github.com/askretov/timex)
[![Licenses](https://img.shields.io/badge/license-mit-brightgreen.svg)](https://opensource.org/licenses/BSD-3-Clause)

# syncsafe

## Introduction
syncsafe package provides synchronization mechanisms similar to native sync package but in more defensive way

## Usage
### Installation
```go
go get github.com/go-ext/syncsafe
```
### WaitGroup examples
```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
defer cancel()
wg := NewWaitGroup()
for i := 0; i < 3; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        time.Sleep(time.Second * time.Duration(i))
    }()
}
if err := wg.WaitContext(ctx); err != nil {
    log.Fatal(err, err.StackTrace())
}
```