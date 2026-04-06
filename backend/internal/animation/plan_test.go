package animation

import "testing"

func TestBuildPlanCoversRequestedDuration(t *testing.T) {
	plan, total, err := BuildPlan(23, []int{5, 10})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}
	if total != 25 {
		t.Fatalf("expected planned total 25, got %d", total)
	}
	if len(plan) != 3 || plan[0] != 10 || plan[1] != 10 || plan[2] != 5 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestBuildPlanExactMatch(t *testing.T) {
	plan, total, err := BuildPlan(20, []int{5, 10})
	if err != nil {
		t.Fatalf("BuildPlan error: %v", err)
	}
	if total != 20 {
		t.Fatalf("expected planned total 20, got %d", total)
	}
	if len(plan) != 2 || plan[0] != 10 || plan[1] != 10 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}
