package websocket

import "testing"

func TestAttemptCounter(t *testing.T) {
	t.Run("starts at attempt 1", func(t *testing.T) {
		c := newAttemptCounter(3)
		if got := c.attempt(); got != 1 {
			t.Errorf("expected: 1, got: %d", got)
		}
	})

	t.Run("advances on failure until attempts are exhausted", func(t *testing.T) {
		c := newAttemptCounter(3)

		if exhausted := c.recordFailure(); exhausted {
			t.Fatal("expected not exhausted on the first failure")
		}
		if got := c.attempt(); got != 2 {
			t.Errorf("expected: 2, got: %d", got)
		}

		if exhausted := c.recordFailure(); exhausted {
			t.Fatal("expected not exhausted on the second failure")
		}
		if got := c.attempt(); got != 3 {
			t.Errorf("expected: 3, got: %d", got)
		}

		if exhausted := c.recordFailure(); !exhausted {
			t.Fatal("expected exhausted on the third failure")
		}
	})

	t.Run("resets after a success, so an earlier failure streak does not carry over", func(t *testing.T) {
		c := newAttemptCounter(2)

		c.recordFailure()
		c.recordSuccess()

		if got := c.attempt(); got != 1 {
			t.Fatalf("expected the counter to reset to 1, got: %d", got)
		}

		if exhausted := c.recordFailure(); exhausted {
			t.Fatal("expected a fresh failure streak to get the full budget again")
		}
		if exhausted := c.recordFailure(); !exhausted {
			t.Fatal("expected exhausted on the second failure of the new streak")
		}
	})
}
