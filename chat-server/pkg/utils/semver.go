package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseSemver(v string) (major, minor, patch int, err error) {
	v = strings.TrimSpace(v)
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver: %s", v)
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, e := strconv.Atoi(p)
		if e != nil || n < 0 {
			return 0, 0, 0, fmt.Errorf("invalid semver: %s", v)
		}
		nums[i] = n
	}
	return nums[0], nums[1], nums[2], nil
}

func ValidSemver(v string) bool {
	_, _, _, err := ParseSemver(v)
	return err == nil
}

func CompareSemver(a, b string) (int, error) {
	am, an, ap, err := ParseSemver(a)
	if err != nil {
		return 0, err
	}
	bm, bn, bp, err := ParseSemver(b)
	if err != nil {
		return 0, err
	}
	if am != bm {
		if am < bm {
			return -1, nil
		}
		return 1, nil
	}
	if an != bn {
		if an < bn {
			return -1, nil
		}
		return 1, nil
	}
	if ap != bp {
		if ap < bp {
			return -1, nil
		}
		return 1, nil
	}
	return 0, nil
}
