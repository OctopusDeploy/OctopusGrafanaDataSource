package plugin

import (
	"encoding/xml"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// BaseResource represents the object holding the common details for any table
// query. Almost all resources are a name/ID combination, but some others like
// releases are version/ID.
type BaseResource struct {
	Name    string `json:"Name"`
	Id      string `json:"Id"`
	Version string `json:"Version"`
}

// PagedResources is the shape of paginated collection endpoints such as
// /api/{space}/deployments and /api/{space}/releases.
type PagedResources struct {
	Items []BaseResource `json:"Items"`
}

type PlainDeploymentItems struct {
	Items []PlainDeployment `json:"Items"`
}

type PlainDeployment struct {
	Name          string `json:"Name"`
	Id            string `json:"Id"`
	Created       string `json:"Created"`
	CreatedParsed time.Time
}

type SpaceResource struct {
	Name      string `json:"Name"`
	Id        string `json:"Id"`
	IsDefault bool   `json:"IsDefault"`
}

type Release struct {
	Name          string `json:"Name"`
	Id            string `json:"Id"`
	Assembled     string `json:"Assembled"`
	AssembledDate time.Time
}

type Deployments struct {
	XMLName     xml.Name     `xml:"Deployments" json:"-"`
	Deployments []Deployment `xml:"Deployment"`
}

type Deployment struct {
	XMLName              xml.Name `xml:"Deployment" json:"-"`
	DeploymentId         string   `xml:"DeploymentId"`
	DeploymentName       string   `xml:"DeploymentName"`
	ProjectId            string   `xml:"ProjectId"`
	ProjectName          string   `xml:"ProjectName"`
	ProjectSlug          string   `xml:"ProjectSlug"`
	TenantId             string   `xml:"TenantId"`
	TenantName           string   `xml:"TenantName"`
	ChannelId            string   `xml:"ChannelId"`
	ChannelName          string   `xml:"ChannelName"`
	EnvironmentId        string   `xml:"EnvironmentId"`
	EnvironmentName      string   `xml:"EnvironmentName"`
	ReleaseId            string   `xml:"ReleaseId"`
	ReleaseVersion       string   `xml:"ReleaseVersion"`
	TaskId               string   `xml:"TaskId"`
	TaskState            string   `xml:"TaskState"`
	Created              string   `xml:"Created"`
	QueueTime            string   `xml:"QueueTime"`
	StartTime            string   `xml:"StartTime"`
	StartTimeParsed      time.Time
	CompletedTime        string `xml:"CompletedTime"`
	CompletedTimeRounded time.Time
	CompletedTimeParsed  time.Time
	DurationSeconds      uint32 `xml:"DurationSeconds"`
	DeployedBy           string `xml:"DeployedBy"`
}

// Query formats handled by QueryData. Any other format is treated as a
// resource type listed in allowedResourceTypes.
const (
	formatTimeSeries            = "timeseries"
	formatTable                 = "table"
	formatAnnotationReport      = "annotation-deploymentreport"
	formatAnnotationDeployments = "annotation-deployments"
)

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
	Query                      backend.DataQuery
}

// usesReportingFeed returns true when the query is answered from the XML
// reporting feed rather than a JSON collection endpoint.
func (qm *queryModel) usesReportingFeed() bool {
	return qm.Format == formatTable || qm.Format == formatTimeSeries || qm.Format == formatAnnotationReport
}
