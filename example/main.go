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
