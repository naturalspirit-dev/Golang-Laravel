package exceptions

import (
	"github.com/goal-web/contracts"
	"github.com/qbhy/goal/http"
	"github.com/qbhy/goal/logs"
	"github.com/qbhy/goal/supports/utils"
	"reflect"
)

type ExceptionHandler struct {
	dontReportExceptions []reflect.Type
}

func NewHandler() contracts.ExceptionHandler {
	return &ExceptionHandler{utils.ConvertToTypes([]contracts.Exception{})}
}

func (handler *ExceptionHandler) Handle(exception contracts.Exception) {
	logs.WithException(exception).
		WithField("exception", reflect.TypeOf(exception).String()).
		Error("ExceptionHandler")

	if httpException, isHttpException := exception.(http.HttpException); isHttpException {
		logs.WithException(httpException).WithFields(contracts.Fields{})
	}

	if handler.ShouldReport(exception) {
		handler.Report(exception)
	}
}

func (handler *ExceptionHandler) Report(exception contracts.Exception) {
}

func (handler *ExceptionHandler) ShouldReport(exception contracts.Exception) bool {
	return !utils.IsInstanceIn(exception, handler.dontReportExceptions...)
}
