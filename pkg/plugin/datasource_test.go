package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const testReportingXML = `<Deployments>
  <Deployment>
    <DeploymentId>Deployments-1</DeploymentId>
    <DeploymentName>Deploy to Dev</DeploymentName>
    <ProjectId>Projects-1</ProjectId>
    <ProjectName>Web</ProjectName>
    <ProjectSlug>web</ProjectSlug>
    <TenantId></TenantId>
    <TenantName></TenantName>
    <ChannelId>Channels-1</ChannelId>
    <ChannelName>Default</ChannelName>
    <EnvironmentId>Environments-1</EnvironmentId>
    <EnvironmentName>Dev</EnvironmentName>
    <ReleaseId>Releases-1</ReleaseId>
    <ReleaseVersion>1.0.0</ReleaseVersion>
    <TaskId>ServerTasks-1</TaskId>
    <TaskState>Failed</TaskState>
    <Created>2020-06-01T09:55:00</Created>
    <QueueTime>2020-06-01T09:56:00</QueueTime>
    <StartTime>2020-06-01T10:00:00</StartTime>
    <CompletedTime>2020-06-01T10:05:00</CompletedTime>
    <DurationSeconds>300</DurationSeconds>
    <DeployedBy>alex</DeployedBy>
  </Deployment>
  <Deployment>
    <DeploymentId>Deployments-2</DeploymentId>
    <DeploymentName>Deploy to Dev</DeploymentName>
    <ProjectId>Projects-1</ProjectId>
    <ProjectName>Web</ProjectName>
    <ProjectSlug>web</ProjectSlug>
    <TenantId></TenantId>
    <TenantName></TenantName>
    <ChannelId>Channels-1</ChannelId>
    <ChannelName>Default</ChannelName>
    <EnvironmentId>Environments-1</EnvironmentId>
    <EnvironmentName>Dev</EnvironmentName>
    <ReleaseId>Releases-2</ReleaseId>
    <ReleaseVersion>1.0.1</ReleaseVersion>
    <TaskId>ServerTasks-2</TaskId>
    <TaskState>Success</TaskState>
    <Created>2020-06-01T11:55:00</Created>
    <QueueTime>2020-06-01T11:56:00</QueueTime>
    <StartTime>2020-06-01T12:00:00</StartTime>
    <CompletedTime>2020-06-01T12:05:00</CompletedTime>
    <DurationSeconds>300</DurationSeconds>
    <DeployedBy>alex</DeployedBy>
  </Deployment>
</Deployments>`

// newMockOctopusServer serves the subset of the Octopus API the plugin uses.
func newMockOctopusServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	writeJSONBody := func(rw http.ResponseWriter, body string) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(body))
	}

	mux.HandleFunc("/api", func(rw http.ResponseWriter, req *http.Request) {
		writeJSONBody(rw, `{"Application":"Octopus Deploy"}`)
	})
	mux.HandleFunc("/api/users/me", func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("X-Octopus-ApiKey") != "API-TEST" {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		writeJSONBody(rw, `{"Id":"Users-1"}`)
	})
	mux.HandleFunc("/api/spaces/all", func(rw http.ResponseWriter, req *http.Request) {
		writeJSONBody(rw, `[{"Name":"Default","Id":"Spaces-1","IsDefault":true}]`)
	})
	mux.HandleFunc("/api/Spaces-1/projects/all", func(rw http.ResponseWriter, req *http.Request) {
		writeJSONBody(rw, `[{"Name":"Web","Id":"Projects-1"}]`)
	})
	mux.HandleFunc("/api/Spaces-1/environments/all", func(rw http.ResponseWriter, req *http.Request) {
		writeJSONBody(rw, `[{"Name":"Dev","Id":"Environments-1"}]`)
	})
	mux.HandleFunc("/api/Spaces-1/reporting/deployments/xml", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/xml")
		_, _ = rw.Write([]byte(testReportingXML))
	})
	mux.HandleFunc("/api/Spaces-1/deployments", func(rw http.ResponseWriter, req *http.Request) {
		writeJSONBody(rw, `{"Items":[{"Name":"Deploy Web release 1.0.1 to Dev","Id":"Deployments-2","Created":"2020-06-01T12:00:00.000+00:00"}]}`)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func newTestDatasource(t *testing.T, serverURL string) *Datasource {
	t.Helper()

	instance, err := NewDatasource(context.Background(), backend.DataSourceInstanceSettings{
		JSONData:                []byte(`{"server":"` + serverURL + `"}`),
		DecryptedSecureJSONData: map[string]string{"apiKey": "API-TEST"},
	})
	if err != nil {
		t.Fatalf("could not create the datasource: %v", err)
	}

	ds, ok := instance.(*Datasource)
	if !ok {
		t.Fatalf("unexpected instance type %T", instance)
	}
	return ds
}

func testTimeRange() backend.TimeRange {
	return backend.TimeRange{
		From: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func runQuery(t *testing.T, ds *Datasource, queryJSON string) backend.DataResponse {
	t.Helper()

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{
			RefID:         "A",
			JSON:          []byte(queryJSON),
			TimeRange:     testTimeRange(),
			MaxDataPoints: 100,
		}},
	})
	if err != nil {
		t.Fatalf("QueryData returned an error: %v", err)
	}
	return resp.Responses["A"]
}

func TestQueryTimeSeries(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"timeseries","spaceName":"Default","projectName":"Web","successField":true,"failureField":true}`)
	if response.Error != nil {
		t.Fatalf("unexpected query error: %v", response.Error)
	}
	if len(response.Frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(response.Frames))
	}

	frame := response.Frames[0]
	// time + success + failure
	if len(frame.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(frame.Fields))
	}

	totalSuccess := uint32(0)
	totalFailure := uint32(0)
	for i := 0; i < frame.Fields[1].Len(); i++ {
		totalSuccess += frame.Fields[1].At(i).(uint32)
		totalFailure += frame.Fields[2].At(i).(uint32)
	}
	if totalSuccess != 1 || totalFailure != 1 {
		t.Errorf("expected 1 success and 1 failure, got %d and %d", totalSuccess, totalFailure)
	}
}

func TestQueryTable(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"table","spaceName":"Default"}`)
	if response.Error != nil {
		t.Fatalf("unexpected query error: %v", response.Error)
	}

	frame := response.Frames[0]
	if frame.Rows() != 2 {
		t.Fatalf("expected 2 rows, got %d", frame.Rows())
	}
	if frame.Fields[1].At(0).(string) != "Deployments-1" {
		t.Errorf("unexpected first deployment: %v", frame.Fields[1].At(0))
	}
}

func TestQueryResourceTable(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"environments","spaceName":"Default"}`)
	if response.Error != nil {
		t.Fatalf("unexpected query error: %v", response.Error)
	}

	frame := response.Frames[0]
	if frame.Rows() != 1 || frame.Fields[0].At(0).(string) != "Dev" {
		t.Errorf("unexpected environments table: %v", frame.Fields[0])
	}
}

func TestQueryTableAddsNoticeWhenNothingMatches(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"table","spaceName":"Default","projectName":"NoSuchProject"}`)
	if response.Error != nil {
		t.Fatalf("unexpected query error: %v", response.Error)
	}

	frame := response.Frames[0]
	if frame.Rows() != 0 {
		t.Fatalf("expected 0 rows, got %d", frame.Rows())
	}
	if frame.Meta == nil || len(frame.Meta.Notices) != 1 {
		t.Fatalf("expected 1 notice, got %+v", frame.Meta)
	}

	notice := frame.Meta.Notices[0]
	if !strings.Contains(notice.Text, "No deployments completed in the selected time range") {
		t.Errorf("unexpected notice text: %q", notice.Text)
	}
	// The mock deployments endpoint reports a deployment created 2020-06-01.
	if !strings.Contains(notice.Text, "2020-06-01") {
		t.Errorf("expected the newest deployment date in the notice, got: %q", notice.Text)
	}
}

func TestQueryTableHasNoNoticeWhenDeploymentsMatch(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"table","spaceName":"Default"}`)
	if response.Error != nil {
		t.Fatalf("unexpected query error: %v", response.Error)
	}

	frame := response.Frames[0]
	if frame.Rows() != 2 {
		t.Fatalf("expected 2 rows, got %d", frame.Rows())
	}
	if frame.Meta != nil && len(frame.Meta.Notices) != 0 {
		t.Errorf("expected no notices, got %+v", frame.Meta.Notices)
	}
}

func TestQueryDeploymentsResourceTable(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"deployments","spaceName":"Default"}`)
	if response.Error != nil {
		t.Fatalf("unexpected query error: %v", response.Error)
	}

	frame := response.Frames[0]
	if frame.Rows() != 1 || frame.Fields[0].At(0).(string) != "Deploy Web release 1.0.1 to Dev" {
		t.Errorf("unexpected deployments table: %v", frame.Fields[0])
	}
}

func TestQueryTableNoticeOmitsDateWithExtraFilters(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	// The latest-deployment lookup only understands the project and
	// environment filters, so a channel filter must suppress the timestamp.
	response := runQuery(t, ds, `{"format":"table","spaceName":"Default","channelName":"NoSuchChannel"}`)
	if response.Error != nil {
		t.Fatalf("unexpected query error: %v", response.Error)
	}

	frame := response.Frames[0]
	if frame.Meta == nil || len(frame.Meta.Notices) != 1 {
		t.Fatalf("expected 1 notice, got %+v", frame.Meta)
	}
	if strings.Contains(frame.Meta.Notices[0].Text, "most recent matching deployment") {
		t.Errorf("the notice should not claim a matching deployment date: %q", frame.Meta.Notices[0].Text)
	}
}

func TestQueryAnnotationDeploymentsRejectsUnknownProject(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"annotation-deployments","spaceName":"Default","projectName":"NoSuchProject"}`)
	if response.Error == nil {
		t.Fatal("expected an error for an unresolvable project filter")
	}
	if !strings.Contains(response.Error.Error(), "NoSuchProject") {
		t.Errorf("unexpected error: %v", response.Error)
	}
}

func TestQueryRejectsUnknownFormat(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"../../users","spaceName":"Default"}`)
	if response.Error == nil {
		t.Fatalf("expected an error for an unknown format")
	}
}

func TestQueryRejectsUnknownSpace(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"table","spaceName":"Nope"}`)
	if response.Error == nil {
		t.Fatalf("expected an error for an unknown space")
	}
}

func TestQueryAnnotationReport(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"annotation-deploymentreport","spaceName":"Default","projectName":"Web"}`)
	if response.Error != nil {
		t.Fatalf("unexpected query error: %v", response.Error)
	}

	frame := response.Frames[0]
	if frame.Rows() != 2 {
		t.Fatalf("expected 2 annotation rows, got %d", frame.Rows())
	}

	fieldNames := []string{}
	for _, field := range frame.Fields {
		fieldNames = append(fieldNames, field.Name)
	}
	want := []string{"time", "timeEnd", "title", "text", "tags"}
	for i, name := range want {
		if fieldNames[i] != name {
			t.Fatalf("field %d = %q, want %q", i, fieldNames[i], name)
		}
	}

	if frame.Fields[2].At(0).(string) != "Web 1.0.0" {
		t.Errorf("unexpected title: %v", frame.Fields[2].At(0))
	}
}

func TestQueryAnnotationDeployments(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	response := runQuery(t, ds, `{"format":"annotation-deployments","spaceName":"Default","projectName":"Web","environmentName":"Dev"}`)
	if response.Error != nil {
		t.Fatalf("unexpected query error: %v", response.Error)
	}

	frame := response.Frames[0]
	if frame.Rows() != 1 {
		t.Fatalf("expected 1 annotation row, got %d", frame.Rows())
	}
	if frame.Fields[1].At(0).(string) != "Deploy Web release 1.0.1 to Dev" {
		t.Errorf("unexpected title: %v", frame.Fields[1].At(0))
	}
}

func TestCheckHealth(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != backend.HealthStatusOk {
		t.Errorf("expected OK, got %v: %s", result.Status, result.Message)
	}
}

func TestCheckHealthWarnsOnPlainHTTP(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != backend.HealthStatusOk {
		t.Fatalf("expected OK, got %v: %s", result.Status, result.Message)
	}
	if !strings.Contains(result.Message, "plain HTTP") {
		t.Errorf("expected a plain HTTP warning in the message, got: %s", result.Message)
	}
}

func TestCheckHealthRejectsBadKey(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)
	ds.client.apiKey = "API-WRONG"

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != backend.HealthStatusError {
		t.Errorf("expected an error status for a rejected API key")
	}
}

type capturedResponse struct {
	status int
	body   []byte
}

func callResource(t *testing.T, ds *Datasource, path string) capturedResponse {
	t.Helper()

	captured := capturedResponse{}
	sender := backend.CallResourceResponseSenderFunc(func(resp *backend.CallResourceResponse) error {
		captured.status = resp.Status
		captured.body = append(captured.body, resp.Body...)
		return nil
	})

	err := ds.CallResource(context.Background(), &backend.CallResourceRequest{
		Method: http.MethodGet,
		Path:   path,
		URL:    "/" + path,
	}, sender)
	if err != nil {
		t.Fatalf("CallResource returned an error: %v", err)
	}
	return captured
}

func TestResourceSpaces(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	resp := callResource(t, ds, "spaces/nameid")
	if resp.status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.status, resp.body)
	}

	spaces := map[string]string{}
	if err := json.Unmarshal(resp.body, &spaces); err != nil {
		t.Fatalf("could not parse the response: %v", err)
	}
	if spaces["Default"] != "Spaces-1" {
		t.Errorf("unexpected spaces: %v", spaces)
	}
}

func TestResourceEntities(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	resp := callResource(t, ds, "Spaces-1/nameid/projects")
	if resp.status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.status, resp.body)
	}

	projects := map[string]string{}
	if err := json.Unmarshal(resp.body, &projects); err != nil {
		t.Fatalf("could not parse the response: %v", err)
	}
	if projects["Web"] != "Projects-1" {
		t.Errorf("unexpected projects: %v", projects)
	}
}

func TestResourceRejectsInvalidInput(t *testing.T) {
	server := newMockOctopusServer(t)
	ds := newTestDatasource(t, server.URL)

	tests := []struct {
		name string
		path string
	}{
		{name: "bad space ID", path: "NotASpace/nameid/projects"},
		{name: "bad resource type", path: "Spaces-1/nameid/nonexistent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := callResource(t, ds, tt.path)
			if resp.status != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", resp.status, resp.body)
			}
		})
	}
}
