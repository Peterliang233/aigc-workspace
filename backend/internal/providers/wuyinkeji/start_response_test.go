package wuyinkeji

import "testing"

func TestParseStartJobIDSuccess(t *testing.T) {
	jobID, err := parseStartJobID([]byte(`{"code":200,"msg":"ok","data":{"id":"job_123"}}`))
	if err != nil {
		t.Fatalf("parseStartJobID returned error: %v", err)
	}
	if jobID != "job_123" {
		t.Fatalf("parseStartJobID=%q want %q", jobID, "job_123")
	}
}

func TestParseStartJobIDReturnsBusinessError(t *testing.T) {
	_, err := parseStartJobID([]byte(`{"code":500,"msg":"转发请求失败","data":null}`))
	if err == nil {
		t.Fatal("parseStartJobID should return business error")
	}
	if got := err.Error(); got != "wuyinkeji start failed: code=500 msg=转发请求失败" {
		t.Fatalf("unexpected error %q", got)
	}
}
