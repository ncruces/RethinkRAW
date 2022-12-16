package main

import (
	"context"
	"log"
	"strconv"

	"github.com/ncruces/rethinkraw/internal/dngconv"
	"golang.org/x/sync/semaphore"
)

var semDNGConverter = semaphore.NewWeighted(3)

func runDNGConverter(ctx context.Context, input, output string, side int, exp *exportSettings) error {
	args := []string{}
	if exp != nil && exp.DNG {
		if exp.Preview != "" {
			args = append(args, "-"+exp.Preview)
		}
		if exp.Lossy {
			args = append(args, "-lossy")
		}
		if exp.Embed {
			args = append(args, "-e")
		}
	} else {
		if side > 0 {
			args = append(args, "-lossy", "-side", strconv.Itoa(side))
		}
		args = append(args, "-p2")
	}

	if err := semDNGConverter.Acquire(ctx, 1); err != nil {
		return err
	}
	defer semDNGConverter.Release(1)

	log.Print("dng converter...")
	return dngconv.Convert(ctx, input, output, args...)
}
