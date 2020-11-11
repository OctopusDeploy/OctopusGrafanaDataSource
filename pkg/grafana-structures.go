package main

type queryModel struct {
	ProjectName                string `json:"projectName"`
	TenantName                 string `json:"tenantName"`
	EnvironmentName            string `json:"environmentName"`
	ChannelName                string `json:"channelName"`
	ReleaseVersion             string `json:"releaseVersion"`
	Format                     string `json:"format"`
	SuccessField               bool   `json:"successField"`
	FailureField               bool   `json:"failureField"`
	CancelledField             bool   `json:"cancelledField"`
	TimedOutField              bool   `json:"timedOutField"`
	TotalDurationField         bool   `json:"totalDurationField"`
	AverageDurationField       bool   `json:"averageDurationField"`
	TotalTimeToRecoveryField   bool   `json:"totalTimeToRecoveryField"`
	AverageTimeToRecoveryField bool   `json:"averageTimeToRecoveryField"`
}

type datasourceModel struct {
	Server         string
	SpaceId        string
	BucketDuration string
	Format         string
}
