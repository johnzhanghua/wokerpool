package pool

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type test struct {
	name string
	fn   ProcessFn
	res  []any
	err  error
}

var ts = []test{
	{
		name: "sqr",
		fn:   sqr,
		err:  nil,
		res:  []any{0, 1, 4, 9, 16, 25, 36, 49, 64, 81},
	},
	{
		name: "sqrErr",
		fn:   sqrErr,
		err:  ErrInvalidValue,
	},
	{
		name: "sqrTimeout",
		fn:   sqrTimeout,
		err:  context.DeadlineExceeded,
	},
}

func TestPool(t *testing.T) {
	jobs := []any{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	for _, tc := range ts {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			p := NewJobPool(ctx, jobs, 2, tc.fn)
			err := p.Process()
			if err != nil {
				if tc.err == nil {
					t.Errorf("%s unexpected error: %v", tc.name, err)
				} else if !errors.Is(err, tc.err) {
					t.Errorf("%s error wrong type :%v\n", tc.name, err)
				}
			} else {
				if tc.err != nil {
					t.Errorf("%s expected error: %v", tc.name, tc.err)
				}
			}
			if tc.res != nil {
				if rs := p.Results(); !reflect.DeepEqual(rs, tc.res) {
					t.Errorf("%s wrong results: %v\n", tc.name, rs)
				}
			} else {
				t.Logf("%s results: %v\n", tc.name, p.Results())
			}
		})
	}

}

func sqr(ctx context.Context, i any) (any, error) {
	// time.Sleep(1 * time.Second)
	d, ok := i.(int)
	if !ok {
		return nil, fmt.Errorf("invalid value: %v, %w", i, ErrInvalidValue)
	}
	return any(d * d), nil
}

func sqrErr(ctx context.Context, i any) (any, error) {
	d, ok := i.(int)
	if !ok {
		return nil, fmt.Errorf("invalid value: %v, %w", i, ErrInvalidValue)
	}
	if d > 0 && d%3 == 0 {
		return nil, fmt.Errorf("invalid value :%d, %w", i, ErrInvalidValue)
	}
	return any(d * d), nil
}

func sqrTimeout(ctx context.Context, i any) (any, error) {
	d, ok := i.(int)
	if !ok {
		return nil, fmt.Errorf("invalid value: %v, %w", i, ErrInvalidValue)
	}
	if d > 0 && d%3 == 0 {
		time.Sleep(6 * time.Second)
	}
	return any(d * d), nil
}
