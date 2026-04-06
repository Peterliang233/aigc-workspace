package gptbest

import "testing"

func TestPickVideoURLFromOutputField(t *testing.T) {
	raw := []byte(`{
		"task_id":"video_4cb59223-d745-4aaa-a7d5-af33cc47c04b",
		"status":"SUCCESS",
		"data":{"output":"https://midjourney-plus.oss-us-west-1.aliyuncs.com/flow/demo.mp4"}
	}`)

	got := pickVideoURL(raw)
	want := "https://midjourney-plus.oss-us-west-1.aliyuncs.com/flow/demo.mp4"
	if got != want {
		t.Fatalf("pickVideoURL() = %q, want %q", got, want)
	}
}

func TestParseVideoStatusAndErrorSuccess(t *testing.T) {
	raw := []byte(`{"status":"SUCCESS","fail_reason":"","data":{"output":"https://example.com/demo.mp4"}}`)

	status, jobErr := parseVideoStatusAndError(raw)
	if status != "succeeded" {
		t.Fatalf("status = %q, want succeeded", status)
	}
	if jobErr != "" {
		t.Fatalf("jobErr = %q, want empty", jobErr)
	}
}
