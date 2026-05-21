package models

import "time"

type Options struct {
	Concurrency int
	Count       int
	Url         string
	Timeout     time.Duration
}
