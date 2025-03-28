package http

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/yael-castro/goarch/internal/app/business"
	"net/http"
	"strconv"
)

func NewUserHandler(cases business.UserCases) (UserHandler, error) {
	if cases == nil {
		return UserHandler{}, errors.New("business logic is not provided")
	}

	return UserHandler{
		cases: cases,
	}, nil
}

type UserHandler struct {
	cases business.UserCases
}

func (u UserHandler) PostUser(c echo.Context) error {
	var user User

	if err := c.Bind(&user); err != nil {
		return err
	}

	user.ID, _ = strconv.ParseInt(c.Param("id"), 10, 64)

	err := u.cases.CreateUser(c.Request().Context(), user.ToBusiness())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, user)
}

func (u UserHandler) PutUser(c echo.Context) error {
	var user User

	if err := c.Bind(&user); err != nil {
		return err
	}

	user.ID, _ = strconv.ParseInt(c.Param("id"), 10, 64)

	err := u.cases.UpdateUser(c.Request().Context(), user.ToBusiness())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

func (u UserHandler) GetUser(c echo.Context) error {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	user, err := u.cases.QueryUser(c.Request().Context(), business.UserID(userID))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, NewUser(&user))
}
