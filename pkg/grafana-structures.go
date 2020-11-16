package main

import "github.com/grafana/grafana-plugin-sdk-go/backend"

type queryModel struct {
	SpaceName                  string `json:"spaceName"`
	ProjectName                string `json:"projectName"`
	TenantName                 string `json:"tenantName"`
	EnvironmentName            string `json:"environmentName"`
	ChannelName                string `json:"channelName"`
	ReleaseVersion             string `json:"releaseVersion"`
	TaskState                  string `json:"TaskState"`
	Format                     string `json:"format"`
	SuccessField               bool   `json:"successField"`
	FailureField               bool   `json:"failureField"`
	CancelledField             bool   `json:"cancelledField"`
	TimedOutField              bool   `json:"timedOutField"`
	TotalDurationField         bool   `json:"totalDurationField"`
	AverageDurationField       bool   `json:"averageDurationField"`
	TotalTimeToRecoveryField   bool   `json:"totalTimeToRecoveryField"`
	AverageTimeToRecoveryField bool   `json:"averageTimeToRecoveryField"`
	TotalCycleTimeField        bool   `json:"totalCycleTimeField"`
	AverageCycleTimeField      bool   `json:"averageCycleTimeField"`
	OctopusQueryUrl            string
	Query                      backend.DataQuery
}

type datasourceModel struct {
	Server         string
	BucketDuration string
	Format         string
}
