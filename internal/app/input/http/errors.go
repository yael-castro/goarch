package http

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/yael-castro/goarch/internal/app/business"
	"net/http"
)

func ErrorHandler(handler echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var userErr business.Error

		if !errors.As(err, &userErr) {
			handler(err, c)
			return
		}

		code := http.StatusInternalServerError
		response := echo.Map{
			"code":    userErr.Error(),
			"message": err.Error(),
		}

		//goland:noinspection ALL
		switch userErr {
		case
			business.ErrInvalidUserName,
			business.ErrInvalidUserEmail,
			business.ErrInvalidUserAge:
			code = http.StatusBadRequest
		case
			business.ErrDuplicateUserEmail:
			code = http.StatusConflict
		case
			business.ErrUserNotFound:
			code = http.StatusNotFound
		}

		_ = c.JSON(code, response)
	}
}
