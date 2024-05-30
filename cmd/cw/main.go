package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tlelson/aws/cloudwatch"
)

/*
go run main.go -group '/aws/lambda/rfs-dev-import-bom' -stream-prefix '2022/10/27/[$LATEST]76c9c30274f242ae90135334b03644e6' -search-pattern 'new predictions started'
*/
func main() {
	gf := flag.String("group", "", "group name filter pattern")
	sp := flag.String("stream-prefix", "", "log streams prefix")
	p := flag.String("search-pattern", "", "text pattern to search by")

	flag.Parse()

	fmt.Printf("group filter: '%s'\n", *gf)
	fmt.Printf("stream prefix: '%s'\n", *sp)
	fmt.Printf("search-pattern: '%s'\n", *p)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cw, err := cloudwatch.New(ctx)
	if err != nil {
		log.Fatalf("failed to initialise cloudwatch client: %v", err)
	}

	logs, errs := cw.Query(ctx, cloudwatch.QueryParameters{
		GroupFilter:  gf,
		StreamPrefix: sp,
		Pattern:      p,
	})

	select {
	case <-ctx.Done():
		log.Fatal("timed out")
	case err := <-errs:
		log.Fatalf("failed to query cloudwatch for logs: %v", err)
	case log := <-logs:
		fmt.Fprintf(os.Stdout, "%v", log)
	}

	fmt.Println("DONE")
}
