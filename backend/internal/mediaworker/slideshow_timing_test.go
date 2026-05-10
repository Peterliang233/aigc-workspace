package mediaworker

import "testing"

func TestStretchDurationsKeepsOriginalWhenEnough(t *testing.T) {
	got := stretchDurations([]int{2000, 3000}, 4500)
	want := []int{2000, 3000}
	assertDurations(t, got, want, 5000)
}

func TestStretchDurationsMatchesTarget(t *testing.T) {
	got := stretchDurations([]int{2000, 3000, 5000}, 12400)
	want := []int{2480, 3720, 6200}
	assertDurations(t, got, want, 12400)
}

func TestStretchDurationsHandlesSingleShot(t *testing.T) {
	got := stretchDurations([]int{1800}, 4200)
	want := []int{4200}
	assertDurations(t, got, want, 4200)
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
