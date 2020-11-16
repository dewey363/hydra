package hydra

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/micro-plat/hydra/conf"
	octx "github.com/micro-plat/hydra/context/ctx"
	"github.com/micro-plat/lib4go/encoding/base64"

	"github.com/micro-plat/hydra/conf/server/auth/basic"
	"github.com/micro-plat/hydra/hydra/servers/pkg/middleware"
	"github.com/micro-plat/hydra/test/assert"
	"github.com/micro-plat/hydra/test/mocks"
)

var authUserKey = "userName"

//author:taoshouyin
//time:2020-11-11
//desc:测试basic验证中间件逻辑
func TestAuthBasic(t *testing.T) {
	type testCase struct {
		name        string
		basicOpts   []basic.Option
		requestPath string
		reqHeadVal  string
		isSet       bool
		wantStatus  int
		wantSpecial string
		user        string
		repHeadVal  string
	}

	tests := []*testCase{
		{name: "basic-未配置", isSet: false, requestPath: "", wantStatus: 200, wantSpecial: "", basicOpts: []basic.Option{}},
		{name: "basic-配置未启动", isSet: true, requestPath: "", wantStatus: 200, wantSpecial: "", basicOpts: []basic.Option{basic.WithDisable()}},
		{name: "basic-配置启动-路由存在,被排除", isSet: true, requestPath: "/basic/test", wantStatus: 200, wantSpecial: "",
			basicOpts: []basic.Option{basic.WithExcludes("/basic/test")}},
		{name: "basic-配置启动-不排除,认证信息为空", isSet: true, requestPath: "/basic/test", wantStatus: 200, wantSpecial: "",
			reqHeadVal: "suibianchuan", basicOpts: []basic.Option{basic.WithExcludes("/basic/test1")}},
		{name: "basic-配置启动-不排除,认证失败", isSet: true, requestPath: "/basic/test", wantStatus: 401, wantSpecial: "basic",
			reqHeadVal: "suibianchuan", repHeadVal: "Basic realm=" + strconv.Quote("Authorization Required"),
			basicOpts: []basic.Option{basic.WithExcludes("/basic/test1"), basic.WithUP("taosy", "tpwd")}},
		{name: "basic-配置启动-不排除,认证成功", isSet: true, requestPath: "/basic/test", wantStatus: 200, wantSpecial: "basic",
			user: "taosy", reqHeadVal: "Basic " + base64.Encode("taosy:tpwd"), basicOpts: []basic.Option{basic.WithExcludes("/basic/test1"), basic.WithUP("taosy", "tpwd")}},
	}

	for _, tt := range tests {
		mockConf := mocks.NewConf()
		//初始化测试用例参数
		confB := mockConf.GetAPI()
		if tt.isSet {
			confB.Basic(tt.basicOpts...)
		}
		serverConf := mockConf.GetAPIConf()
		ctx := &mocks.MiddleContext{
			MockMeta:     conf.NewMeta(),
			MockUser:     &mocks.MockUser{MockClientIP: "192.168.0.1", MockAuth: &octx.Auth{}},
			MockResponse: &mocks.MockResponse{MockStatus: 200, MockHeader: map[string][]string{}},
			MockRequest: &mocks.MockRequest{
				MockPath: &mocks.MockPath{
					MockHeader:      http.Header{"Authorization": []string{tt.reqHeadVal}},
					MockRequestPath: tt.requestPath,
				},
			},
			MockAPPConf: serverConf,
		}

		//获取中间件
		handler := middleware.BasicAuth()
		//调用中间件
		handler(ctx)
		//断言结果
		gotStatus, _ := ctx.Response().GetFinalResponse()
		assert.Equalf(t, tt.wantStatus, gotStatus, tt.name, tt.wantStatus, gotStatus)
		gotSpecial := ctx.Response().GetSpecials()
		assert.Equalf(t, tt.wantSpecial, gotSpecial, tt.name, tt.wantSpecial, gotSpecial)
		gotUser := ctx.Meta().GetString(authUserKey)
		assert.Equalf(t, tt.user, gotUser, tt.name, tt.user, gotUser)
		if tt.user != "" {
			quthReq := (ctx.User().Auth().Request()).(map[string]interface{})
			assert.Equalf(t, tt.user, quthReq[authUserKey].(string), tt.name, tt.user, quthReq[authUserKey])
		}

		if tt.repHeadVal != "" {
			header := ctx.Response().GetHeaders()
			assert.Equalf(t, []string{tt.repHeadVal}, header["WWW-Authenticate"], tt.name, tt.repHeadVal, header["WWW-Authenticate"])
		}
	}
}
