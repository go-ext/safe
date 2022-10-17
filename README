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