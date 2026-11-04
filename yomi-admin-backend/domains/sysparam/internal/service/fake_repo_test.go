package service

import (
	"context"
	"strings"
	"yomi-admin-backend/domains/sysparam/internal/repo_contract"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// fakeRepo 内存版参数仓储：按存储顺序维护，支持注入查询故障
type fakeRepo struct {
	rows     []*entity.Param
	getErr   error
	dupOnNew bool // 下一次 CreateParam 强制返回唯一键冲突
}

var _ repo_contract.SysParamRepo = (*fakeRepo)(nil)

func (f *fakeRepo) Page(_ context.Context, filter *repo_contract.ParamFilter) ([]*entity.Param, int64, error) {
	if f.getErr != nil {
		return nil, 0, f.getErr
	}
	items := make([]*entity.Param, 0, len(f.rows))
	for _, row := range f.rows {
		if filter.Status != "" && row.Status != filter.Status {
			continue
		}
		if filter.ValueType != "" && row.ValueType != filter.ValueType {
			continue
		}
		items = append(items, row)
	}
	return items, int64(len(items)), nil
}

func (f *fakeRepo) FindByID(_ context.Context, id int64) (*entity.Param, bool, error) {
	for _, row := range f.rows {
		if row.ID == id {
			return row, true, nil
		}
	}
	return nil, false, nil
}

func (f *fakeRepo) FindByKey(_ context.Context, key string) (*entity.Param, bool, error) {
	for _, row := range f.rows {
		if row.Key == key {
			return row, true, nil
		}
	}
	return nil, false, nil
}

func (f *fakeRepo) GetEnabledByKey(_ context.Context, key string) (*entity.Param, bool, error) {
	if f.getErr != nil {
		return nil, false, f.getErr
	}
	for _, row := range f.rows {
		if row.Key == key && row.Status == enums.StatusEnabled {
			return row, true, nil
		}
	}
	return nil, false, nil
}

func (f *fakeRepo) FindEnabledByKeys(_ context.Context, keys []string) ([]*entity.Param, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	wanted := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		wanted[key] = struct{}{}
	}
	items := make([]*entity.Param, 0, len(keys))
	for _, row := range f.rows {
		if row.Status != enums.StatusEnabled {
			continue
		}
		if _, ok := wanted[row.Key]; ok {
			items = append(items, row)
		}
	}
	return items, nil
}

func (f *fakeRepo) CreateParam(_ context.Context, param *entity.Param, _ int64) error {
	if f.dupOnNew {
		f.dupOnNew = false
		return errors.Sentinel(repo_contract.ErrDuplicatedKey)
	}
	for _, row := range f.rows {
		if row.Key == param.Key {
			return errors.Sentinel(repo_contract.ErrDuplicatedKey)
		}
	}
	f.rows = append(f.rows, param)
	return nil
}

func (f *fakeRepo) UpdateParam(_ context.Context, param *entity.Param, _ int64) error {
	for i, row := range f.rows {
		if row.ID == param.ID {
			updated := *param
			f.rows[i] = &updated
			return nil
		}
	}
	return errors.New("参数不存在: " + param.Key)
}

func (f *fakeRepo) UpdateParamStatus(_ context.Context, id int64, status enums.Status, _ int64) error {
	for _, row := range f.rows {
		if row.ID == id {
			row.Status = status
			return nil
		}
	}
	return errors.New("参数不存在")
}

func (f *fakeRepo) DeleteParam(_ context.Context, id int64) error {
	for i, row := range f.rows {
		if row.ID == id {
			f.rows = append(f.rows[:i], f.rows[i+1:]...)
			return nil
		}
	}
	return errors.New("参数不存在")
}

// newTestIDs 递增生成实体 ID
type newTestIDs struct {
	next int64
}

func (g *newTestIDs) NextID() (int64, error) {
	g.next++
	return g.next, nil
}

// longString 生成长度为 n 的字符串
func longString(n int) string {
	return strings.Repeat("x", n)
}
