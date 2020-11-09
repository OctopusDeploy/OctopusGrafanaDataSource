package main

import (
	"encoding/xml"
	"time"
)

type Deployments struct {
	XMLName     xml.Name     `xml:"Deployments"`
	Deployments []Deployment `xml:"Deployment"`
}

type Deployment struct {
	XMLName             xml.Name `xml:"Deployment"`
	DeploymentId        string   `xml:"DeploymentId"`
	DeploymentName      string   `xml:"DeploymentName"`
	ProjectId           string   `xml:"ProjectId"`
	ProjectName         string   `xml:"ProjectName"`
	ProjectSlug         string   `xml:"ProjectSlug"`
	TenantId            string   `xml:"TenantId"`
	TenantName          string   `xml:"TenantName"`
	ChannelId           string   `xml:"ChannelId"`
	ChannelName         string   `xml:"ChannelName"`
	EnvironmentId       string   `xml:"EnvironmentId"`
	EnvironmentName     string   `xml:"EnvironmentName"`
	ReleaseId           string   `xml:"ReleaseId"`
	ReleaseVersion      string   `xml:"ReleaseVersion"`
	TaskId              string   `xml:"TaskId"`
	TaskState           string   `xml:"TaskState"`
	Created             string   `xml:"Created"`
	QueueTime           string   `xml:"QueueTime"`
	StartTime           string   `xml:"StartTime"`
	CompletedTime       string   `xml:"CompletedTime"`
	CompetedTimeRounded time.Time
	DurationSeconds     uint8  `xml:"DurationSeconds"`
	DeployedBy          string `xml:"DeployedBy"`
}
