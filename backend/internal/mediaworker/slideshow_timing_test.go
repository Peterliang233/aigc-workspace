package mediaworker

import "testing"

func TestStretchDurationsCompresses(t *testing.T) {
	got := stretchDurations([]int{2000, 3000}, 4500)
	want := []int{1800, 2700}
	assertDurations(t, got, want, 4500)
}

func TestStretchDurationsStretches(t *testing.T) {
	got := stretchDurations([]int{2000, 3000, 5000}, 12400)
	want := []int{2480, 3720, 6200}
	assertDurations(t, got, want, 12400)
}

func TestStretchDurationsHandlesSingleShot(t *testing.T) {
	got := stretchDurations([]int{1800}, 4200)
	want := []int{4200}
	assertDurations(t, got, want, 4200)
}

func TestStretchDurationsNoChange(t *testing.T) {
	got := stretchDurations([]int{3000, 2000}, 5000)
	want := []int{3000, 2000}
	assertDurations(t, got, want, 5000)
}

func TestStretchDurationsMinFloor(t *testing.T) {
	got := stretchDurations([]int{500, 500, 9000}, 2000)
	// 500 would scale to 100, but floor is 1000
	// last element gets remainder: 2000 - 1000 - 1000 = 0 → floor 1000
	if got[0] < 1000 || got[1] < 1000 || got[2] < 1000 {
		t.Fatalf("expected all >= 1000, got %v", got)
	}
}

func assertDurations(t *testing.T, got, want []int, total int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len=%d want=%d", len(got), len(want))
	}
	sum := 0
	for i := range got {
		sum += got[i]
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%d want=%d", i, got[i], want[i])
		}
	}
	if sum != total {
		t.Fatalf("sum=%d want=%d", sum, total)
	}
}
