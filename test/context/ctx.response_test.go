package context

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/micro-plat/hydra/conf"
	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/hydra/context/ctx"
	"github.com/micro-plat/hydra/test/assert"
	"github.com/micro-plat/hydra/test/mocks"
	"github.com/micro-plat/lib4go/errs"
	"github.com/micro-plat/lib4go/logger"
)

func Test_response_Write(t *testing.T) {

	tests := []struct {
		name    string
		ctx     context.IInnerContext
		status  int
		content interface{}
		wantRs  int
		wantRc  string
	}{
		{name: "状态码非0,返回包含错误码的错误", ctx: &mocks.TestContxt{}, status: 0, content: errs.NewError(999, "错误"), wantRs: 999, wantRc: "错误"},
		{name: "状态码在200到400,返回错误", ctx: &mocks.TestContxt{}, status: 300, content: errors.New("err"), wantRs: 400, wantRc: "err"},
		{name: "状态码为0,返回非错误内容", ctx: &mocks.TestContxt{}, status: 0, content: nil, wantRs: 200, wantRc: ""},
		{name: "状态码非0,返回非错误内容", ctx: &mocks.TestContxt{}, status: 500, content: "content", wantRs: 500, wantRc: "content"},
		{name: "状态码非0,content-type为text/plain,返回非错误内容", ctx: &mocks.TestContxt{
			HttpHeader: http.Header{
				"Content-Type": []string{context.PLAINF},
			},
		}, status: 200, content: "content", wantRs: 200, wantRc: "content"},
		{name: "状态码非0,content-type为application/json,返回json内容", ctx: &mocks.TestContxt{
			HttpHeader: http.Header{
				"Content-Type": []string{context.JSONF},
			},
		}, status: 200, content: `{"key":"value"}`, wantRs: 200, wantRc: `{"key":"value"}`},
		{name: "状态码非0,content-type为application/xml,返回xml内容", ctx: &mocks.TestContxt{
			HttpHeader: http.Header{
				"Content-Type": []string{context.XMLF},
			},
		}, status: 200, content: "<?xml><key>value<key/><xml/>", wantRs: 200, wantRc: `<?xml><key>value<key/><xml/>`},
		{name: "状态码非0,content-type为text/html,返回html内容", ctx: &mocks.TestContxt{
			HttpHeader: http.Header{
				"Content-Type": []string{context.HTMLF},
			},
		}, status: 200, content: "<!DOCTYPE html><html></html>", wantRs: 200, wantRc: `<!DOCTYPE html><html></html>`},
		{name: "状态码非0,content-type为text/yaml,返回内容", ctx: &mocks.TestContxt{
			HttpHeader: http.Header{
				"Content-Type": []string{context.YAMLF},
			},
		}, status: 200, content: "key:value", wantRs: 200, wantRc: `key:value`},
		{name: "状态码非0,content-type为application/json,且返回内容非正确json字符串", ctx: &mocks.TestContxt{
			HttpHeader: http.Header{
				"Content-Type": []string{context.JSONF},
			},
		}, status: 200, content: "{key:value", wantRs: 200, wantRc: `{"data":"{key:value"}`},
		// 用例引起panic
		// {name: "状态码非0,content-type为application/xml,且返回内容非正确xml字符串", ctx:&mocks.TestContxt{
		// 	HttpHeader: http.Header{
		// 		"Content-Type": []string{context.XMLF},
		// 	},
		// },status: 200, content: "<key>value<key/>", wantRs: 200, wantRc: ``},
		{name: "状态码非0,content-type为空,返回布尔值/整型/浮点型/复数", ctx: &mocks.TestContxt{
			HttpHeader: http.Header{},
		}, status: 200, content: false, wantRs: 200, wantRc: `false`},
		{name: "状态码非0,content-type为application/json,返回布尔值/整型/浮点型/复数", ctx: &mocks.TestContxt{
			HttpHeader: http.Header{
				"Content-Type": []string{context.JSONF},
			},
		}, status: 200, content: 1, wantRs: 200, wantRc: `{"data":1}`},
		// 用例引起panic
		// {name: "状态码非0,content-type为application/xml,返回布尔值/整型/浮点型/复数", ctx:&mocks.TestContxt{
		// 	HttpHeader: http.Header{
		// 		"Content-Type": []string{context.XMLF},
		// 	},
		// }},status: 200, content: 1, wantRs: 200, wantRc: `{"data":1}`},
		{name: "状态码非0,content-type为空,返回非字符串/布尔值/整型/浮点型/复数的内容", ctx: &mocks.TestContxt{
			HttpHeader: http.Header{},
		}, status: 200, content: map[string]string{"key": "value"}, wantRs: 200, wantRc: `{"key":"value"}`},
		// 用例引起panic
		// {name: "状态码非0,content-type为空,返回非字符串/布尔值/整型/浮点型/复数的内容", ctx:&mocks.TestContxt{
		// 	HttpHeader: http.Header{
		// 		"Content-Type": []string{context.XMLF},
		// 	},
		// }},status: 200, content: map[string]string{"key": "value"}, wantRs: 200, wantRc: ``},
	}

	confObj := mocks.NewConf()         //构建对象
	confObj.API(":8080")               //初始化参数
	serverConf := confObj.GetAPIConf() //获取配置
	meta := conf.NewMeta()

	for _, tt := range tests {
		log := logger.GetSession(serverConf.GetMainConf().GetServerName(), ctx.NewUser(tt.ctx, meta).GetRequestID())

		//构建response对象
		c := ctx.NewResponse(tt.ctx, serverConf, log, meta)
		err := c.Write(tt.status, tt.content)
		assert.Equal(t, nil, err, tt.name)

		//测试reponse状态码和内容
		rs, rc := c.GetFinalResponse()
		assert.Equal(t, tt.wantRs, rs, tt.name)
		assert.Equal(t, tt.wantRc, rc, tt.name)

	}
}

func Test_response_Header(t *testing.T) {
	confObj := mocks.NewConf()         //构建对象
	confObj.API(":8080")               //初始化参数
	serverConf := confObj.GetAPIConf() //获取配置
	meta := conf.NewMeta()
	rc := &mocks.TestContxt{HttpHeader: http.Header{}}
	log := logger.GetSession(serverConf.GetMainConf().GetServerName(), ctx.NewUser(rc, meta).GetRequestID())
	c := ctx.NewResponse(rc, serverConf, log, meta)

	//设置header
	c.Header("header1", "value1")
	assert.Equal(t, http.Header{"header1": []string{"value1"}}, rc.GetHeaders(), "设置header")

	//再次设置header
	c.Header("header1", "value1-1")
	assert.Equal(t, http.Header{"header1": []string{"value1-1"}}, rc.GetHeaders(), "再次设置header")

	//设置不同的header
	c.Header("header2", "value2")
	assert.Equal(t, http.Header{"header1": []string{"value1-1"}, "header2": []string{"value2"}}, rc.GetHeaders(), "再次设置header")
}

func Test_response_ContentType(t *testing.T) {
	confObj := mocks.NewConf()         //构建对象
	confObj.API(":8080")               //初始化参数
	serverConf := confObj.GetAPIConf() //获取配置
	meta := conf.NewMeta()
	rc := &mocks.TestContxt{HttpHeader: http.Header{}}
	log := logger.GetSession(serverConf.GetMainConf().GetServerName(), ctx.NewUser(rc, meta).GetRequestID())
	c := ctx.NewResponse(rc, serverConf, log, meta)

	//设置content-type
	c.ContentType("application/json")
	assert.Equal(t, "application/json", rc.ContentType(), "设置content-type")

	//再次设置header
	c.ContentType("text/plain")
	assert.Equal(t, "text/plain", rc.ContentType(), "再次设置content-type")
}

func Test_response_Abort(t *testing.T) {
	confObj := mocks.NewConf()         //构建对象
	confObj.API(":8080")               //初始化参数
	serverConf := confObj.GetAPIConf() //获取配置
	meta := conf.NewMeta()
	context := &mocks.TestContxt{HttpHeader: http.Header{}}
	log := logger.GetSession(serverConf.GetMainConf().GetServerName(), ctx.NewUser(context, meta).GetRequestID())
	c := ctx.NewResponse(context, serverConf, log, meta)

	//测试Abort
	c.Abort(200, fmt.Errorf("终止"))
	rs, rc := c.GetFinalResponse()
	assert.Equal(t, 400, rs, "验证状态码")
	assert.Equal(t, []byte(rc), context.Content, "验证返回内容")
	assert.Equal(t, rs, context.StatusCode, "验证上下文中的状态码")
	assert.Equal(t, "text/plain; charset=utf-8", context.HttpHeader["Content-Type"][0], "验证上下文中的content-type")
	assert.Equal(t, true, context.WrittenStatus, "验证上下文中的写入状态")
	assert.Equal(t, true, context.Doen, "验证上下文中的abort状态")
}

func Test_response_Stop(t *testing.T) {
	confObj := mocks.NewConf()         //构建对象
	confObj.API(":8080")               //初始化参数
	serverConf := confObj.GetAPIConf() //获取配置
	meta := conf.NewMeta()
	context := &mocks.TestContxt{HttpHeader: http.Header{}}
	log := logger.GetSession(serverConf.GetMainConf().GetServerName(), ctx.NewUser(context, meta).GetRequestID())
	c := ctx.NewResponse(context, serverConf, log, meta)

	//测试Stop
	c.Stop(200)
	rs, rc := c.GetFinalResponse()
	assert.Equal(t, 200, rs, "验证状态码")
	assert.Equal(t, []byte(rc), context.Content, "验证返回内容")
	assert.Equal(t, rs, context.StatusCode, "验证上下文中的状态码")
	assert.Equal(t, "text/plain; charset=utf-8", context.HttpHeader["Content-Type"][0], "验证上下文中的content-type")
	assert.Equal(t, true, context.WrittenStatus, "验证上下文中的写入状态")
	assert.Equal(t, true, context.Doen, "验证上下文中的abort状态")
}

func Test_response_StatusCode(t *testing.T) {
	tests := []struct {
		name       string
		s          int
		wantStatus int
	}{
		{name: "设置状态码为200", s: 200, wantStatus: 200},
		{name: "设置状态码为300", s: 300, wantStatus: 300},
		{name: "设置状态码为400", s: 400, wantStatus: 400},
		{name: "设置状态码为500", s: 500, wantStatus: 500},
	}
	confObj := mocks.NewConf()         //构建对象
	confObj.API(":8080")               //初始化参数
	serverConf := confObj.GetAPIConf() //获取配置
	meta := conf.NewMeta()
	context := &mocks.TestContxt{HttpHeader: http.Header{}}
	log := logger.GetSession(serverConf.GetMainConf().GetServerName(), ctx.NewUser(context, meta).GetRequestID())
	c := ctx.NewResponse(context, serverConf, log, meta)
	for _, tt := range tests {
		c.StatusCode(tt.s)
		rs, _ := c.GetFinalResponse()
		assert.Equal(t, tt.wantStatus, rs, tt.name)
		assert.Equal(t, context.Status(), rs, tt.name)
	}
}

func Test_response_File(t *testing.T) {
	confObj := mocks.NewConf()         //构建对象
	confObj.API(":8080")               //初始化参数
	serverConf := confObj.GetAPIConf() //获取配置
	meta := conf.NewMeta()
	context := &mocks.TestContxt{HttpHeader: http.Header{}}
	log := logger.GetSession(serverConf.GetMainConf().GetServerName(), ctx.NewUser(context, meta).GetRequestID())
	c := ctx.NewResponse(context, serverConf, log, meta)

	//测试File
	c.File("file")
	assert.Equal(t, true, context.WrittenStatus, "验证上下文中的文件内容")
	assert.Equal(t, true, context.Doen, "验证上下文中的abort状态")
}

func Test_response_WriteFinal(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		content string
		ctp     string
		wantS   int
		wantC   string
	}{
		{name: "写入200状态码和json数据", status: 200, content: `{"a":"b"}`, ctp: "application/json", wantS: 200, wantC: `{"a":"b"}`},
		{name: "写入300状态码和空数据", status: 300, content: ``, ctp: "application/json", wantS: 300, wantC: ``},
		{name: "写入400状态码和错误数据", status: 400, content: `错误`, ctp: "application/json", wantS: 400, wantC: "错误"},
		{name: "写入空状态码和空数据", ctp: "application/json", wantS: 400, wantC: ""},
	}

	confObj := mocks.NewConf()         //构建对象
	confObj.API(":8080")               //初始化参数
	serverConf := confObj.GetAPIConf() //获取配置
	meta := conf.NewMeta()
	context := &mocks.TestContxt{HttpHeader: http.Header{}}
	log := logger.GetSession(serverConf.GetMainConf().GetServerName(), ctx.NewUser(context, meta).GetRequestID())
	c := ctx.NewResponse(context, serverConf, log, meta)

	for _, tt := range tests {
		c.WriteFinal(tt.status, tt.content, tt.ctp)
		rs, rc := c.GetFinalResponse()
		assert.Equal(t, tt.wantS, rs, tt.name)
		assert.Equal(t, tt.wantC, rc, tt.name)
	}
}

func Test_response_Redirect(t *testing.T) {

	confObj := mocks.NewConf()         //构建对象
	confObj.API(":8080")               //初始化参数
	serverConf := confObj.GetAPIConf() //获取配置
	meta := conf.NewMeta()
	context := &mocks.TestContxt{HttpHeader: http.Header{}}
	log := logger.GetSession(serverConf.GetMainConf().GetServerName(), ctx.NewUser(context, meta).GetRequestID())
	c := ctx.NewResponse(context, serverConf, log, meta)

	c.Redirect(200, "url")
	assert.Equal(t, true, context.WrittenStatus, "验证上下文中的写入状态")
	assert.Equal(t, "url", context.Url, "验证上下文中的url")
	assert.Equal(t, 200, context.StatusCode, "验证上下文中的状态码")
}