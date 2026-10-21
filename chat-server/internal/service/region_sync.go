package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"app/internal/model"
	"app/pkg/logger"
)

const (
	xzqhSourceBase      = "https://dmfw.mca.gov.cn/9095"
	xzqhMaxLevel        = 3
	xzqhRequestInterval = 500 * time.Millisecond
	xzqhHTTPTimeout     = 120 * time.Second
	xzqhMaxAttempts     = 3
)

type xzqhNode struct {
	Code     string     `json:"code"`
	Name     string     `json:"name"`
	Level    int        `json:"level"`
	Type     string     `json:"type"`
	Children []xzqhNode `json:"children"`
}

var (
	xzqhHTTP   = &http.Client{Timeout: xzqhHTTPTimeout}
	xzqhMu     sync.Mutex
	xzqhLastAt time.Time
)

func (s *RegionService) Sync(ctx context.Context) error {
	startedAt := time.Now()
	dataYear := strconv.Itoa(startedAt.Year())

	nodes, partial, err := s.fetchAllRegions(ctx)
	if err != nil {
		return err
	}

	flat := flattenRegionNodes(nodes, "", "", dataYear)
	flat = mergeRegions(flat, supplementalRegions(dataYear))
	codes := make([]string, 0, len(flat))
	for _, item := range flat {
		codes = append(codes, item.Code)
	}

	inserted, updated, err := s.regionRepo.UpsertBatch(ctx, flat)
	if err != nil {
		return err
	}

	revoked := 0
	if !partial {
		rows, err := s.regionRepo.RevokeExcept(ctx, codes)
		if err != nil {
			return err
		}
		revoked = int(rows)
	}

	s.log.Info(ctx, "行政区划同步完成",
		logger.Int("fetched", len(flat)),
		logger.Int("inserted", inserted),
		logger.Int("updated", updated),
		logger.Int("revoked", revoked),
		logger.Bool("partial", partial),
	)
	return nil
}

func (s *RegionService) fetchAllRegions(ctx context.Context) ([]xzqhNode, bool, error) {
	provinces, err := fetchXZQHTree(ctx, "", 1)
	if err != nil {
		return nil, false, err
	}

	out := make([]xzqhNode, 0, len(provinces))
	partial := false
	remainLevel := xzqhMaxLevel - 1
	for _, province := range provinces {
		resolved := resolveProvinceCode(province)
		if resolved == "" {
			continue
		}
		province.Code = resolved
		if isSupplementalProvince(resolved) {
			out = append(out, province)
			continue
		}
		sub, err := fetchXZQHRoot(ctx, province.Code, remainLevel)
		if err != nil {
			s.log.Warn(ctx, "拉取省级下级失败，跳过该省下级",
				logger.String("province", province.Name),
				logger.String("code", province.Code),
				logger.Error(err),
			)
			out = append(out, province)
			partial = true
			continue
		}
		sub.Code = province.Code
		sub.Name = province.Name
		if sub.Level == 0 {
			sub.Level = province.Level
		}
		out = append(out, sub)
	}
	return out, partial, nil
}

func flattenRegionNodes(nodes []xzqhNode, parentCode, parentPath, dataYear string) []model.Region {
	out := make([]model.Region, 0, 64)
	walkRegionNodes(nodes, parentCode, parentPath, dataYear, &out)
	return out
}

func walkRegionNodes(nodes []xzqhNode, parentCode, parentPath, dataYear string, out *[]model.Region) {
	for _, node := range nodes {
		rawCode := strings.TrimSpace(node.Code)
		code := normalizeRegionCode(rawCode)
		if !isNumericRegionCode(rawCode) && parentCode == "" {
			if resolved := resolveProvinceCode(node); resolved != "" {
				code = resolved
			}
		}
		if code == "" || node.Name == "" {
			continue
		}
		level := detectRegionLevel(code)
		if node.Level > 0 {
			level = uint8(node.Level)
		}
		pcode := parentCode
		if pcode == "" {
			pcode = parentRegionCode(code)
		}
		fullPath := node.Name
		if parentPath != "" {
			fullPath = parentPath + "/" + node.Name
		}
		*out = append(*out, model.Region{
			Code:       code,
			Name:       node.Name,
			ShortName:  regionShortName(node.Name),
			ParentCode: pcode,
			Level:      level,
			FullPath:   fullPath,
			TiandituGB: toTiandituGB(code),
			AmapAdcode: toAmapAdcode(code),
			DataYear:   dataYear,
			Status:     model.RegionStatusActive,
		})
		if len(node.Children) > 0 {
			walkRegionNodes(node.Children, code, fullPath, dataYear, out)
		}
	}
}

func supplementalRegions(dataYear string) []model.Region {
	type row struct {
		code, name, parent string
		level              uint8
	}

	rows := []row{
		{"710000000000", "台湾省", "", model.RegionLevelProvince},
		{"710100000000", "台北市", "710000000000", model.RegionLevelCity},
		{"710200000000", "高雄市", "710000000000", model.RegionLevelCity},
		{"710300000000", "台南市", "710000000000", model.RegionLevelCity},
		{"710400000000", "台中市", "710000000000", model.RegionLevelCity},
		{"710500000000", "金门县", "710000000000", model.RegionLevelCity},
		{"710600000000", "南投县", "710000000000", model.RegionLevelCity},
		{"710700000000", "基隆市", "710000000000", model.RegionLevelCity},
		{"710800000000", "新竹市", "710000000000", model.RegionLevelCity},
		{"710900000000", "嘉义市", "710000000000", model.RegionLevelCity},
		{"711100000000", "新北市", "710000000000", model.RegionLevelCity},
		{"711200000000", "宜兰县", "710000000000", model.RegionLevelCity},
		{"711300000000", "新竹县", "710000000000", model.RegionLevelCity},
		{"711400000000", "桃园市", "710000000000", model.RegionLevelCity},
		{"711500000000", "苗栗县", "710000000000", model.RegionLevelCity},
		{"711700000000", "彰化县", "710000000000", model.RegionLevelCity},
		{"711900000000", "云林县", "710000000000", model.RegionLevelCity},
		{"712100000000", "屏东县", "710000000000", model.RegionLevelCity},
		{"712400000000", "台东县", "710000000000", model.RegionLevelCity},
		{"712500000000", "花莲县", "710000000000", model.RegionLevelCity},
		{"712600000000", "澎湖县", "710000000000", model.RegionLevelCity},
		{"810000000000", "香港特别行政区", "", model.RegionLevelProvince},
		{"810101000000", "中西区", "810000000000", model.RegionLevelCity},
		{"810102000000", "湾仔区", "810000000000", model.RegionLevelCity},
		{"810103000000", "东区", "810000000000", model.RegionLevelCity},
		{"810104000000", "南区", "810000000000", model.RegionLevelCity},
		{"810105000000", "油尖旺区", "810000000000", model.RegionLevelCity},
		{"810106000000", "深水埗区", "810000000000", model.RegionLevelCity},
		{"810107000000", "九龙城区", "810000000000", model.RegionLevelCity},
		{"810108000000", "黄大仙区", "810000000000", model.RegionLevelCity},
		{"810109000000", "观塘区", "810000000000", model.RegionLevelCity},
		{"810110000000", "荃湾区", "810000000000", model.RegionLevelCity},
		{"810111000000", "屯门区", "810000000000", model.RegionLevelCity},
		{"810112000000", "元朗区", "810000000000", model.RegionLevelCity},
		{"810113000000", "北区", "810000000000", model.RegionLevelCity},
		{"810114000000", "大埔区", "810000000000", model.RegionLevelCity},
		{"810115000000", "西贡区", "810000000000", model.RegionLevelCity},
		{"810116000000", "沙田区", "810000000000", model.RegionLevelCity},
		{"810117000000", "葵青区", "810000000000", model.RegionLevelCity},
		{"810118000000", "离岛区", "810000000000", model.RegionLevelCity},
		{"820000000000", "澳门特别行政区", "", model.RegionLevelProvince},
		{"820101000000", "花地玛堂区", "820000000000", model.RegionLevelCity},
		{"820102000000", "圣安多尼堂区", "820000000000", model.RegionLevelCity},
		{"820103000000", "大堂区", "820000000000", model.RegionLevelCity},
		{"820104000000", "望德堂区", "820000000000", model.RegionLevelCity},
		{"820105000000", "风顺堂区", "820000000000", model.RegionLevelCity},
		{"820106000000", "嘉模堂区", "820000000000", model.RegionLevelCity},
		{"820107000000", "路氹填海区", "820000000000", model.RegionLevelCity},
		{"820108000000", "圣方济各堂区", "820000000000", model.RegionLevelCity},
	}

	names := make(map[string]string, len(rows))
	for _, item := range rows {
		names[item.code] = item.name
	}

	out := make([]model.Region, 0, len(rows))
	for _, item := range rows {
		fullPath := item.name
		if item.parent != "" {
			if parentName := names[item.parent]; parentName != "" {
				fullPath = parentName + "/" + item.name
			}
		}
		out = append(out, model.Region{
			Code:       item.code,
			Name:       item.name,
			ShortName:  regionShortName(item.name),
			ParentCode: item.parent,
			Level:      item.level,
			FullPath:   fullPath,
			TiandituGB: toTiandituGB(item.code),
			AmapAdcode: toAmapAdcode(item.code),
			DataYear:   dataYear,
			Status:     model.RegionStatusActive,
		})
	}
	return out
}

func mergeRegions(base, extra []model.Region) []model.Region {
	byCode := make(map[string]model.Region, len(base)+len(extra))
	for _, item := range base {
		byCode[item.Code] = item
	}
	for _, item := range extra {
		byCode[item.Code] = item
	}
	out := make([]model.Region, 0, len(byCode))
	for _, item := range byCode {
		out = append(out, item)
	}
	return out
}

func isNumericRegionCode(code string) bool {
	code = strings.TrimSpace(code)
	if len(code) < 6 {
		return false
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func resolveProvinceCode(node xzqhNode) string {
	code := strings.TrimSpace(node.Code)
	if isNumericRegionCode(code) {
		return normalizeRegionCode(code)
	}
	name := strings.TrimSpace(node.Name)
	switch {
	case strings.Contains(name, "台湾"):
		return "710000000000"
	case strings.Contains(name, "香港"):
		return "810000000000"
	case strings.Contains(name, "澳门"):
		return "820000000000"
	default:
		return ""
	}
}

func isSupplementalProvince(code string) bool {
	switch normalizeRegionCode(code) {
	case "710000000000", "810000000000", "820000000000":
		return true
	default:
		return false
	}
}

func fetchXZQHTree(ctx context.Context, code string, maxLevel int) ([]xzqhNode, error) {
	root, err := fetchXZQHRoot(ctx, code, maxLevel)
	if err != nil {
		return nil, err
	}
	return xzqhNodeChildren(root), nil
}

func fetchXZQHRoot(ctx context.Context, code string, maxLevel int) (xzqhNode, error) {
	var lastErr error
	for attempt := 1; attempt <= xzqhMaxAttempts; attempt++ {
		node, err := fetchXZQHRootOnce(ctx, code, maxLevel)
		if err == nil {
			return node, nil
		}
		lastErr = err
		if attempt == xzqhMaxAttempts || !xzqhRetryable(err) {
			break
		}
		wait := time.Duration(attempt) * time.Second
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return xzqhNode{}, ctx.Err()
		case <-timer.C:
		}
	}
	return xzqhNode{}, lastErr
}

func xzqhRetryable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "HTTP 5") || strings.Contains(msg, "请求民政部区划接口失败")
}

func fetchXZQHRootOnce(ctx context.Context, code string, maxLevel int) (xzqhNode, error) {
	if err := xzqhThrottle(ctx); err != nil {
		return xzqhNode{}, err
	}

	endpoint := xzqhSourceBase + "/xzqh/getList"
	query := url.Values{}
	if code != "" {
		query.Set("code", code)
	}
	query.Set("maxLevel", strconv.Itoa(maxLevel))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return xzqhNode{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; chat-server/1.0)")
	req.Header.Set("Accept", "application/json")

	resp, err := xzqhHTTP.Do(req)
	if err != nil {
		return xzqhNode{}, fmt.Errorf("请求民政部区划接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return xzqhNode{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return xzqhNode{}, fmt.Errorf("民政部区划接口 HTTP %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	var payload struct {
		Data xzqhNode `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return xzqhNode{}, fmt.Errorf("解析民政部区划响应失败: %w", err)
	}

	if code != "" {
		if payload.Data.Code == "" || payload.Data.Code == "00" {
			return xzqhNode{}, fmt.Errorf("民政部区划接口返回空数据")
		}
		return payload.Data, nil
	}
	if len(payload.Data.Children) == 0 {
		return xzqhNode{}, fmt.Errorf("民政部区划接口返回空数据")
	}
	return xzqhNode{Children: payload.Data.Children}, nil
}

func xzqhThrottle(ctx context.Context) error {
	xzqhMu.Lock()
	wait := xzqhRequestInterval - time.Since(xzqhLastAt)
	xzqhMu.Unlock()

	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}

	xzqhMu.Lock()
	xzqhLastAt = time.Now()
	xzqhMu.Unlock()
	return nil
}

func xzqhNodeChildren(data xzqhNode) []xzqhNode {
	if len(data.Children) > 0 {
		return data.Children
	}
	if data.Code != "" && data.Code != "00" {
		return []xzqhNode{data}
	}
	return nil
}

func detectRegionLevel(code string) uint8 {
	code = normalizeRegionCode(code)
	if len(code) != 12 {
		return 0
	}
	switch {
	case code[2:] == "0000000000":
		return 1
	case code[4:] == "00000000":
		return 2
	case code[6:] == "000000":
		return 3
	default:
		return 4
	}
}

func parentRegionCode(code string) string {
	code = normalizeRegionCode(code)
	switch detectRegionLevel(code) {
	case 1:
		return ""
	case 2:
		return code[:2] + "0000000000"
	case 3:
		return code[:4] + "00000000"
	case 4:
		return code[:6] + "000000"
	default:
		return ""
	}
}

func toTiandituGB(code string) string {
	code = normalizeRegionCode(code)
	if len(code) < 6 {
		return ""
	}
	return "156" + code[:6]
}

func toAmapAdcode(code string) string {
	code = normalizeRegionCode(code)
	if len(code) < 6 {
		return ""
	}
	return code[:6]
}

func regionShortName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	if len([]rune(name)) <= 2 {
		return name
	}
	for _, suffix := range []string{
		"特别行政区", "自治州", "自治县", "自治区", "地区", "盟", "省", "市", "区", "县",
	} {
		if strings.HasSuffix(name, suffix) && len(name) > len(suffix) {
			short := strings.TrimSuffix(name, suffix)
			if len([]rune(short)) >= 2 {
				return short
			}
			return name
		}
	}
	return name
}
