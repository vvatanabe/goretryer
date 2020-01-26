# goretryer

Retry exponential backoff algorithm in Go. Generalized HTTP Request retry logic in [aws/aws-sdk-go](https://github.com/aws/aws-sdk-go).

## Requires

- Go 1.13+

## Installation

This package can be installed with the go get command:

```
$ go get github.com/vvatanabe/goretryer
```

## Usage

### Basically

```go
package main

import (
	"context"
	"errors"
	"github.com/vvatanabe/goretryer/exponential"
	"log"
	"time"
)

var ErrTemporary = errors.New("temporary")

func main() {

	log.SetFlags(log.Lmicroseconds)

	retryer := exponential.Retryer{
		NumMaxRetries: 5,
		MinRetryDelay: 300 * time.Millisecond,
		MaxRetryDelay: 300 * time.Second,
	}

	var cnt int
	operation := func(ctx context.Context) error {
		cnt++
		log.Println("cnt", cnt)
		return ErrTemporary
	}

	isErrorRetryable := func(err error) bool {
		return err == ErrTemporary
	}

	over, err := retryer.Do(context.Background(), operation, isErrorRetryable)
	if err != nil {
		log.Printf("retry over: %v, error: %v\n", over, err)
	}
}
```

## Acknowledgments

See [aws-sdk-go/aws/client/default_retryer.go](https://github.com/aws/aws-sdk-go/blob/84fbd57ef75762a07aade079776907d01be3891d/aws/client/default_retryer.go) for great origins.

## Bugs and Feedback

For bugs, questions and discussions please use the Github Issues.

