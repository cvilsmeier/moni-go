package monibot

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cvilsmeier/monibot-go/internal/assert"
)

func TestApi(t *testing.T) {
	str := func(a any) string {
		switch x := a.(type) {
		case Watchdog:
			return fmt.Sprintf("Id=%v, Name=%v, IntervalMillis=%v", x.Id, x.Name, x.IntervalMillis)
		case Machine:
			return fmt.Sprintf("Id=%v, Name=%v", x.Id, x.Name)
		case Metric:
			return fmt.Sprintf("Id=%v, Name=%v, Type=%v", x.Id, x.Name, x.Type)
		}
		return "???"
	}
	ass := assert.New(t)
	// this test uses a fake HTTP sender
	sender := &fakeSender{}
	// create Api
	api := &Api{sender}
	// GET ping
	{
		sender.calls = nil
		sender.responses = append(sender.responses, fakeResponse{})
		err := api.GetPing()
		ass.Nil(err)
		ass.Eq(1, len(sender.calls))
		ass.Eq("GET ping", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// GET watchdogs
	{
		sender.calls = nil
		resp := `[
			{"id":"0001", "name":"Cronjob 1", "intervalMillis": 72000000},
			{"id":"0002", "name":"Cronjob 2", "intervalMillis": 36000000}
		]`
		sender.responses = append(sender.responses, fakeResponse{data: []byte(resp)})
		watchdogs, err := api.GetWatchdogs()
		ass.Nil(err)
		ass.Eq(2, len(watchdogs))
		ass.Eq("Id=0001, Name=Cronjob 1, IntervalMillis=72000000", str(watchdogs[0]))
		ass.Eq("Id=0002, Name=Cronjob 2, IntervalMillis=36000000", str(watchdogs[1]))
		ass.Eq(1, len(sender.calls))
		ass.Eq("GET watchdogs", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// GET watchdog/00000001
	{
		sender.calls = nil
		resp := `{"id":"0001", "name":"Cronjob 1", "intervalMillis": 72000000}`
		sender.responses = append(sender.responses, fakeResponse{data: []byte(resp)})
		watchdog, err := api.GetWatchdog("00000001")
		ass.Nil(err)
		ass.Eq("Id=0001, Name=Cronjob 1, IntervalMillis=72000000", str(watchdog))
		ass.Eq(1, len(sender.calls))
		ass.Eq("GET watchdog/00000001", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// POST watchdog/00000001/heartbeat
	{
		sender.calls = nil
		sender.responses = append(sender.responses, fakeResponse{})
		err := api.PostWatchdogHeartbeat("00000001")
		ass.Nil(err)
		ass.Eq(1, len(sender.calls))
		ass.Eq("POST watchdog/00000001/heartbeat", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// GET machines
	{
		sender.calls = nil
		resp := `[
			{"id":"01", "name":"Server 1"},
			{"id":"02", "name":"Server 2"}
		]`
		sender.responses = append(sender.responses, fakeResponse{data: []byte(resp)})
		machines, err := api.GetMachines()
		ass.Nil(err)
		ass.Eq(2, len(machines))
		ass.Eq("Id=01, Name=Server 1", str(machines[0]))
		ass.Eq("Id=02, Name=Server 2", str(machines[1]))
		ass.Eq(1, len(sender.calls))
		ass.Eq("GET machines", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// GET machine/01
	{
		sender.calls = nil
		resp := `{"id":"01", "name":"Server 1"}`
		sender.responses = append(sender.responses, fakeResponse{data: []byte(resp)})
		machine, err := api.GetMachine("01")
		ass.Nil(err)
		ass.Eq("Id=01, Name=Server 1", str(machine))
		ass.Eq(1, len(sender.calls))
		ass.Eq("GET machine/01", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// POST machine/00000001/sample
	{
		sender.calls = nil
		sender.responses = append(sender.responses, fakeResponse{})
		tstamp := time.Date(2023, 10, 27, 10, 0, 0, 0, time.UTC)
		sample := MachineSample{
			Tstamp:      tstamp.UnixMilli(),
			Load1:       1.01,
			Load5:       0.78,
			Load15:      0.12,
			CpuPercent:  12,
			MemPercent:  34,
			DiskPercent: 12,
			DiskRead:    678,
			DiskWrite:   567,
			NetRecv:     13,
			NetSend:     14,
		}
		err := api.PostMachineSample("00000001", sample)
		ass.Nil(err)
		ass.Eq(1, len(sender.calls))
		ass.Eq("POST machine/00000001/sample tstamp=1698400800000&load1=1.010&load5=0.780&load15=0.120&cpu=12&mem=34&disk=12&diskRead=678&diskWrite=567&netRecv=13&netSend=14", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// POST machine/00000001/text
	{
		sender.calls = nil
		sender.responses = append(sender.responses, fakeResponse{})
		text := "line1\nline2\n\n"
		err := api.PostMachineText("00000001", text)
		ass.Nil(err)
		ass.Eq(1, len(sender.calls))
		ass.Eq("POST machine/00000001/text text=line1%0Aline2%0A%0A", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// GET metrics
	{
		sender.calls = nil
		resp := `[
			{"id":"01", "name":"Metric 1", "type": 0},
			{"id":"02", "name":"Metric 2", "type": 1}
		]`
		sender.responses = append(sender.responses, fakeResponse{data: []byte(resp)})
		metrics, err := api.GetMetrics()
		ass.Nil(err)
		ass.Eq(2, len(metrics))
		ass.Eq("Id=01, Name=Metric 1, Type=0", str(metrics[0]))
		ass.Eq("Id=02, Name=Metric 2, Type=1", str(metrics[1]))
		ass.Eq(1, len(sender.calls))
		ass.Eq("GET metrics", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// GET metric/01
	{
		sender.calls = nil
		resp := `{"id":"01", "name":"Metric 1", "type": 0}`
		sender.responses = append(sender.responses, fakeResponse{data: []byte(resp)})
		metric, err := api.GetMetric("01")
		ass.Nil(err)
		ass.Eq("Id=01, Name=Metric 1, Type=0", str(metric))
		ass.Eq(1, len(sender.calls))
		ass.Eq("GET metric/01", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// POST metric/00000001/inc
	{
		sender.calls = nil
		sender.responses = append(sender.responses, fakeResponse{nil, fmt.Errorf("connect timeout")})
		err := api.PostMetricInc("00000001", 42)
		ass.Eq("connect timeout", err.Error())
		ass.Eq(1, len(sender.calls))
		ass.Eq("POST metric/00000001/inc value=42", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// POST metric/00000001/set
	{
		sender.calls = nil
		sender.responses = append(sender.responses, fakeResponse{nil, fmt.Errorf("connect timeout")})
		err := api.PostMetricSet("00000001", 113)
		ass.Eq("connect timeout", err.Error())
		ass.Eq(1, len(sender.calls))
		ass.Eq("POST metric/00000001/set value=113", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
	// POST metric/00000042/values
	{
		sender.calls = nil
		sender.responses = append(sender.responses, fakeResponse{nil, fmt.Errorf("connect timeout")})
		err := api.PostMetricValues("010101", []int64{3, 5, 2, 5, 0, 3, 4, 3, 1})
		ass.Eq("connect timeout", err.Error())
		ass.Eq(1, len(sender.calls))
		// "0%2C1%2C2%2C3%3A3%2C4%2C5%3A2" = urlEncode("0,1,2,3:3,4,5:2")
		ass.Eq("POST metric/010101/values values=0%2C1%2C2%2C3%3A3%2C4%2C5%3A2", sender.calls[0])
		ass.Eq(0, len(sender.responses))
	}
}

type fakeSender struct {
	calls     []string
	responses []fakeResponse
}

func (f *fakeSender) Send(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	call := strings.TrimSpace(method + " " + path + " " + string(body))
	f.calls = append(f.calls, call)
	if len(f.responses) == 0 {
		return nil, fmt.Errorf("no response for %s %s", method, path)
	}
	re := f.responses[0]
	f.responses = f.responses[1:]
	return re.data, re.err
}

type fakeResponse struct {
	data []byte
	err  error
}
