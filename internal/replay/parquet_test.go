package replay

import "testing"

func TestAdmitTickMixedPrecisionAndSequence(t *testing.T) {
	first, err := admitTick(0, false, 1_735_689_599_999_000, 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := admitTick(first.TimestampMS, true, 1_735_689_600_999_999, 1)
	if err != nil {
		t.Fatal(err)
	}
	if first.TimestampMS != 1_735_689_600_000 || second.TimestampMS != 1_735_689_601_000 {
		t.Fatalf("unexpected timestamps: %d %d", first.TimestampMS, second.TimestampMS)
	}
	if _, err := admitTick(second.TimestampMS, true, 1_735_689_602_999_999, 1); err == nil {
		t.Fatal("gap was accepted")
	}
}
