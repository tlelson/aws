package cloudwatch

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type CloudWatch struct {
	Client *cloudwatchlogs.Client
}

func New(ctx context.Context) (CloudWatch, error) {
	cw := CloudWatch{}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return cw, fmt.Errorf("unable to load SDK config, %w", err)
	}

	cw.Client = cloudwatchlogs.NewFromConfig(cfg)
	return cw, nil
}

type QueryParameters struct {
	GroupFilter    *string
	StreamPrefix   *string
	LookBackPeriod *time.Duration
	Pattern        *string
}

func ptr[T any](x T) *T { return &x }

func (cw CloudWatch) Query(ctx context.Context, p QueryParameters) (chan types.FilteredLogEvent, chan error) {

	if p.LookBackPeriod == nil {
		p.LookBackPeriod = ptr(24 * time.Hour)
	}

	results := make(chan types.FilteredLogEvent, 1)
	errors := make(chan error, 1)

	go func() {
		defer close(results)
		defer close(errors)
		paginator := cloudwatchlogs.NewFilterLogEventsPaginator(
			cw.Client,
			ptr(cloudwatchlogs.FilterLogEventsInput{
				LogGroupName:  p.GroupFilter,
				EndTime:       ptr(time.Now().UnixMilli()),
				FilterPattern: p.Pattern,
				//Interleaved:         new(bool),
				//Limit:               new(int32),
				LogStreamNamePrefix: p.StreamPrefix,
				//LogStreamNames:      []string{},
				//NextToken:           new(string)
				StartTime: ptr(time.Now().Add(-*p.LookBackPeriod).UnixMilli()),
			}))
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				errors <- err
				return
			}
			for _, event := range page.Events {
				results <- event
			}
		}
	}()
	return results, errors
}
