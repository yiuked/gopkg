package xerr

import (
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
)

const (
	DBDuplicate = 1062
)

type XError struct {
	OutErr error
	InErr  error
}

func NewXErr(in, out error) *XError {
	return &XError{OutErr: out, InErr: in}
}

func (err *XError) Error() string {
	var errString string
	if err.InErr != nil {
		errString = fmt.Sprintf("%s in err:%s", errString, err.InErr.Error())
	}
	if err.OutErr != nil {
		errString = fmt.Sprintf("%s out err:%s", errString, err.OutErr.Error())
	}
	return errString
}

var (
	DBErr     = errors.New("数据操作失败")
	ParamsErr = errors.New("参数错误")
)

func IsDBDuplicate(err error) bool {
	if mysqlError, ok := err.(*mysql.MySQLError); ok && mysqlError.Number == DBDuplicate {
		return true
	}
	return false
}
