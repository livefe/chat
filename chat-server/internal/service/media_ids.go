package service

import (
	"encoding/json"
	"fmt"

	"app/pkg/errs"

	"gorm.io/datatypes"
)

func normalizeMediaIDs(ids []uint, max int) ([]uint, error) {
	if len(ids) > max {
		return nil, errs.Invalid(fmt.Sprintf("最多上传 %d 张图片", max))
	}
	out := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func encodeMediaIDs(ids []uint) (datatypes.JSON, error) {
	if len(ids) == 0 {
		return datatypes.JSON([]byte("[]")), nil
	}
	raw, err := json.Marshal(ids)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(raw), nil
}

func parseMediaIDs(raw datatypes.JSON) []uint {
	if len(raw) == 0 {
		return []uint{}
	}
	var ids []uint
	if err := json.Unmarshal(raw, &ids); err != nil {
		return []uint{}
	}
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			out = append(out, id)
		}
	}
	return out
}
