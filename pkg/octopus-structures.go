package main

import (
	"encoding/xml"
	"time"
)

type IdResource struct {
	Name string `json:Name`
	Id   string `json:Id`
}

type SpaceResource struct {
	Name      string `json:Name`
	Id        string `json:Id`
	IsDefault bool   `json:IsDefault`
}

type Release struct {
	Name          string `json:Name`
	Id            string `json:Id`
	Assembled     string `json:Assembled`
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
