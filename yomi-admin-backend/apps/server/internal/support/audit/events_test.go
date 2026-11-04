package audit

import (
	"testing"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// TestOutcomeFromError 校验业务结果映射
func TestOutcomeFromError(t *testing.T) {
	success := OutcomeFromError(nil)
	if success.Result != enums.ResultSuccess || success.BizCode != "OK" {
		t.Fatalf("成功结果 = %+v", success)
	}

	biz := OutcomeFromError(errors.Sentinel(errors.NewBiz("ADMIN_INVALID_CREDENTIAL", "用户名或密码错误")))
	if biz.Result != enums.ResultFailure || biz.BizCode != "ADMIN_INVALID_CREDENTIAL" {
		t.Fatalf("业务错误结果 = %+v", biz)
	}

	unknown := OutcomeFromError(errors.New("boom"))
	if unknown.Result != enums.ResultFailure || unknown.BizCode != "INTERNAL_ERROR" {
		t.Fatalf("未知错误结果 = %+v", unknown)
	}
}
