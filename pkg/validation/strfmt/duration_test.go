// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package strfmt

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuration(t *testing.T) {
	pp := Duration(0)

	err := pp.UnmarshalText([]byte("0ms"))
	require.NoError(t, err)
	err = pp.UnmarshalText([]byte("yada"))
	require.Error(t, err)

	orig := "2ms"
	b := []byte(orig)
	bj := []byte("\"" + orig + "\"")

	err = pp.UnmarshalText(b)
	require.NoError(t, err)

	err = pp.UnmarshalText([]byte("three week"))
	require.Error(t, err)

	err = pp.UnmarshalText([]byte("9999999999999999999999999999999999999999999999999999999 weeks"))
	require.Error(t, err)

	txt, err := pp.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, orig, string(txt))

	err = pp.UnmarshalJSON(bj)
	require.NoError(t, err)
	assert.Equal(t, orig, pp.String())

	err = pp.UnmarshalJSON([]byte("yada"))
	require.Error(t, err)

	err = pp.UnmarshalJSON([]byte(`"12 parsecs"`))
	require.Error(t, err)

	err = pp.UnmarshalJSON([]byte(`"12 y"`))
	require.Error(t, err)

	b, err = pp.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, bj, b)
}

func testDurationParser(t *testing.T, toParse string, expected time.Duration) {
	t.Helper()

	r, e := ParseDuration(toParse)
	require.NoError(t, e)
	assert.Equal(t, expected, r)
}

func TestDurationParser_Failed(t *testing.T) {
	_, e := ParseDuration("45 wekk")
	require.Error(t, e)
}

func TestIsDuration_Failed(t *testing.T) {
	e := IsDuration("45 weeekks")
	assert.False(t, e)
}

func TestDurationParser(t *testing.T) {
	testcases := map[string]time.Duration{
		// parse the short forms without spaces
		"1ns": 1 * time.Nanosecond,
		"1us": 1 * time.Microsecond,
		"1µs": 1 * time.Microsecond,
		"1ms": 1 * time.Millisecond,
		"1s":  1 * time.Second,
		"1m":  1 * time.Minute,
		"1h":  1 * time.Hour,
		"1hr": 1 * time.Hour,
		"1d":  24 * time.Hour,
		"1w":  7 * 24 * time.Hour,
		"1wk": 7 * 24 * time.Hour,

		// parse the long forms without spaces
		"1nanoseconds":  1 * time.Nanosecond,
		"1nanos":        1 * time.Nanosecond,
		"1microseconds": 1 * time.Microsecond,
		"1micros":       1 * time.Microsecond,
		"1millis":       1 * time.Millisecond,
		"1milliseconds": 1 * time.Millisecond,
		"1second":       1 * time.Second,
		"1sec":          1 * time.Second,
		"1min":          1 * time.Minute,
		"1minute":       1 * time.Minute,
		"1hour":         1 * time.Hour,
		"1day":          24 * time.Hour,
		"1week":         7 * 24 * time.Hour,

		// parse the short forms with spaces
		"1  ns": 1 * time.Nanosecond,
		"1  us": 1 * time.Microsecond,
		"1  µs": 1 * time.Microsecond,
		"1  ms": 1 * time.Millisecond,
		"1  s":  1 * time.Second,
		"1  m":  1 * time.Minute,
		"1  h":  1 * time.Hour,
		"1  hr": 1 * time.Hour,
		"1  d":  24 * time.Hour,
		"1  w":  7 * 24 * time.Hour,
		"1  wk": 7 * 24 * time.Hour,

		// parse the long forms without spaces
		"1  nanoseconds":  1 * time.Nanosecond,
		"1  nanos":        1 * time.Nanosecond,
		"1  microseconds": 1 * time.Microsecond,
		"1  micros":       1 * time.Microsecond,
		"1  millis":       1 * time.Millisecond,
		"1  milliseconds": 1 * time.Millisecond,
		"1  second":       1 * time.Second,
		"1  sec":          1 * time.Second,
		"1  min":          1 * time.Minute,
		"1  minute":       1 * time.Minute,
		"1  hour":         1 * time.Hour,
		"1  day":          24 * time.Hour,
		"1  week":         7 * 24 * time.Hour,

		// parse composite forms
		"1m45s":                time.Minute + 45*time.Second,
		"1 m45 s":              time.Minute + 45*time.Second,
		"1m 45s":               time.Minute + 45*time.Second,
		"1  minute 45 seconds": time.Minute + 45*time.Second,
	}

	for str, dur := range testcases {
		t.Run(str, func(t *testing.T) {
			testDurationParser(t, str, dur)

			// negative duration
			testDurationParser(t, "-"+str, -dur)
			testDurationParser(t, "- "+str, -dur)
		})
	}
}

// TestDurationParser_EdgeCases covers ParseDuration branches that the happy-path
// tests don't exercise: the "0" shortcut, empty input left after a sign or
// spaces, and inputs that have a decimal point but no digit on either side.
func TestDurationParser_EdgeCases(t *testing.T) {
	t.Run("zero shortcut returns 0 with no error", func(t *testing.T) {
		for _, in := range []string{"0", "-0", "+0", "- 0", "+ 0"} {
			input := in
			t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
				t.Parallel()

				d, err := ParseDuration(input)
				require.NoError(t, err)
				assert.Equal(t, time.Duration(0), d)
			})
		}
	})

	t.Run("empty payload after sign or spaces is rejected", func(t *testing.T) {
		for _, in := range []string{"", "-", "+", "   ", "- ", "+    "} {
			input := in
			t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
				t.Parallel()

				_, err := ParseDuration(input)
				require.Error(t, err)
			})
		}
	})

	t.Run("decimal point without digits is rejected", func(t *testing.T) {
		// A leading '.' passes the first numeric check but produces neither an
		// integer nor a fractional part, exercising the "I dare you" branch.
		for _, in := range []string{".s", ".h", ".d", "-.s", "+.h", "1m .s"} {
			input := in
			t.Run(input, func(t *testing.T) {
				t.Parallel()

				_, err := ParseDuration(input)
				require.Error(t, err)
			})
		}
	})
}

// TestDurationParser_Overflow covers every numerical-overflow branch in
// ParseDuration, leadingInt, and leadingFraction.
//
// The boundary values are derived from maxUint64 = 1<<63 (the magnitude of
// math.MinInt64). The fractional cases hinge on (1<<63 - 1)/10 = 922337203685477580.
func TestDurationParser_Overflow(t *testing.T) {
	overflows := []struct {
		name  string
		input string
	}{
		{
			name: "leadingInt overflow after multiply-add",
			// 19 digits where the last one pushes x past 1<<63 even though
			// x <= maxUint64/10 still held before the final multiply-add.
			input: "9223372036854775809ns",
		},
		{
			name: "v*unit fits but adding fraction overflows",
			// 2562047*hours = 9223369200000000000 (under 1<<63), then
			// +0.9*hours = +3240000000000 pushes the sum past 1<<63.
			input: "2562047.9h",
		},
		{
			name: "running total d exceeds maxUint64 across tokens",
			// Each token alone fits (9e18 < 1<<63 ~= 9.223e18), but the sum
			// (1.8e19) overflows in the in-loop d > maxUint64 check.
			input: "9000000000000ms 9000000000000ms",
		},
		{
			name: "single positive value equals maxUint64",
			// 1<<63 ns parses through leadingInt successfully, then fails the
			// final d > maxUint64-1 check (only the negative form fits, as
			// time.Duration(math.MinInt64)).
			input: "9223372036854775808ns",
		},
	}

	for _, tt := range overflows {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d, err := ParseDuration(tc.input)
			require.Errorf(t, err, "parsed %q as %v, expected overflow error", tc.input, d)
		})
	}
}

// TestDurationParser_FractionPrecisionLoss covers the three overflow-handling
// branches in leadingFraction. Unlike integer overflow, fractional overflow is
// not an error: per the helper's contract it stops accumulating precision and
// returns the partial value, so the parser still produces a valid duration —
// just truncated near the 18th fractional digit.
func TestDurationParser_FractionPrecisionLoss(t *testing.T) {
	// All three inputs share the same first 18 fractional digits; precision
	// caps there regardless of which overflow branch fires, so the parsed
	// value is identical.
	const truncated = time.Duration(922337203) // 0.922337203 * 1s, rounded down

	cases := []struct {
		name  string
		input string
	}{
		{
			name: "y overflow on 19th digit",
			// 19th digit '9' makes y = x*10+9 > 1<<63.
			input: "0.9223372036854775809s",
		},
		{
			name: "x exceeds threshold on 20th digit",
			// 19th digit '8' lands y exactly at 1<<63 (still accepted),
			// 20th digit then trips the x > (1<<63-1)/10 guard.
			input: "0.92233720368547758089s",
		},
		{
			name: "trailing digits skipped once overflow is set",
			// 21st digit hits the early `if overflow` continue branch.
			input: "0.922337203685477580891s",
		},
	}

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d, err := ParseDuration(tc.input)
			require.NoError(t, err)
			assert.Equal(t, truncated, d)
		})
	}
}

// TestDurationParser_NegativeMinDuration verifies that the smallest
// representable duration parses successfully even though the equivalent
// positive value overflows.
func TestDurationParser_NegativeMinDuration(t *testing.T) {
	d, err := ParseDuration("-9223372036854775808ns")
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1<<63), d)
}

func TestIsDuration_Caveats(t *testing.T) {
	// This works too
	e := IsDuration("45 weeks")
	assert.True(t, e)

	// This no longer works
	e = IsDuration("45 weekz")
	assert.False(t, e)

	// This works too
	e = IsDuration("12 hours")
	assert.True(t, e)

	// This works too
	e = IsDuration("12 minutes")
	assert.True(t, e)

	// This does not work
	e = IsDuration("12 phours")
	assert.False(t, e)
}

func TestDeepCopyDuration(t *testing.T) {
	dur := Duration(42)
	in := &dur

	out := new(Duration)
	in.DeepCopyInto(out)
	assert.Equal(t, in, out)

	out2 := in.DeepCopy()
	assert.Equal(t, in, out2)

	var inNil *Duration
	out3 := inNil.DeepCopy()
	assert.Nil(t, out3)
}

func TestIssue169FractionalDuration(t *testing.T) {
	for _, tt := range []struct {
		Input       string
		Expected    string
		ExpectError bool
	}{
		{
			Input:    "1.5 h",
			Expected: "1h30m0s",
		},
		{
			Input:    "1.5 d",
			Expected: "36h0m0s",
		},
		{
			Input:    "3.14159 d",
			Expected: "75h23m53.376s",
		},
		{
			Input:    "- 3.14159 d",
			Expected: "-75h23m53.376s",
		},
		{
			Input:       "3.141.59 d",
			ExpectError: true,
		},
		{
			Input:    ".314159 d",
			Expected: "7h32m23.3376s",
		},
		{
			Input:       "314159. d",
			ExpectError: true,
		},
	} {
		fractionalDuration := tt

		if fractionalDuration.ExpectError {
			t.Run(fmt.Sprintf("invalid fractional duration %s should NOT parse", fractionalDuration.Input), func(t *testing.T) {
				t.Parallel()

				require.False(t, IsDuration(fractionalDuration.Input))
			})

			continue
		}

		t.Run(fmt.Sprintf("fractional duration %s should parse", fractionalDuration.Input), func(t *testing.T) {
			t.Parallel()

			require.True(t, IsDuration(fractionalDuration.Input))

			var d Duration
			require.NoError(t, d.UnmarshalText([]byte(fractionalDuration.Input)))

			require.Equal(t, fractionalDuration.Expected, d.String())

			dd, err := ParseDuration(fractionalDuration.Input)
			require.NoError(t, err)
			require.Equal(t, fractionalDuration.Expected, dd.String())
		})
	}
}
