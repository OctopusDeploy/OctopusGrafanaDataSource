package plugin

import (
	"testing"
	"time"
)

func TestGetBucketDurationNeverReturnsZero(t *testing.T) {
	tests := []struct {
		name          string
		queryDuration time.Duration
		maxDataPoints int64
	}{
		{name: "zero length range", queryDuration: 0, maxDataPoints: 100},
		{name: "range shorter than point count", queryDuration: 10 * time.Nanosecond, maxDataPoints: 100},
		{name: "zero max data points", queryDuration: time.Hour, maxDataPoints: 0},
		{name: "normal range", queryDuration: time.Hour, maxDataPoints: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buckets, bucketDuration := getBucketDuration(tt.queryDuration, tt.maxDataPoints)
			if buckets < 1 {
				t.Errorf("expected at least 1 bucket, got %d", buckets)
			}
			if bucketDuration <= 0 {
				t.Errorf("expected a positive bucket duration, got %v", bucketDuration)
			}
		})
	}
}

func TestBucketAssignmentIncludesRangeBoundaries(t *testing.T) {
	from := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	buckets, bucketDuration := getBucketDuration(to.Sub(from), 4)

	deployments := Deployments{Deployments: []Deployment{
		{CompletedTime: "2020-06-01T00:00:00"}, // exactly From
		{CompletedTime: "2020-06-01T00:40:00"}, // mid range
		{CompletedTime: "2020-06-01T01:00:00"}, // exactly To
		{CompletedTime: "2020-06-01T02:00:00"}, // outside the range
	}}

	setCompletedTimeRounded(deployments, from, to, buckets, bucketDuration)

	if got := deployments.Deployments[0].CompletedTimeRounded; !got.Equal(from) {
		t.Errorf("a completion at From should land in the first bucket, got %v", got)
	}
	if got := deployments.Deployments[1].CompletedTimeRounded; !got.Equal(from.Add(2 * bucketDuration)) {
		t.Errorf("a mid range completion landed in the wrong bucket: %v", got)
	}
	lastBucket := from.Add(bucketDuration * time.Duration(buckets-1))
	if got := deployments.Deployments[2].CompletedTimeRounded; !got.Equal(lastBucket) {
		t.Errorf("a completion at To should land in the last bucket, got %v", got)
	}
	if got := deployments.Deployments[3].CompletedTimeRounded; !got.IsZero() {
		t.Errorf("a completion outside the range should not be bucketed, got %v", got)
	}

	// Every bucket start must be inside [From, To).
	for i := int64(0); i < buckets; i++ {
		start := from.Add(bucketDuration * time.Duration(i))
		if start.Before(from) || !start.Before(to) {
			t.Errorf("bucket %d starts outside the query range: %v", i, start)
		}
	}
}
