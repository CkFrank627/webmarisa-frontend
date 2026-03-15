package Controllers

import (
	"fmt"
	"github.com/kataras/iris"
	"net"
	"net/http"
	"server/Models"
	"server/Services"
)

func Add(ctx iris.Context, service Services.IMemoriseService) ModelAndView {
	memory := Models.Memorise{}

	err := ctx.ReadForm(&memory)

	if err != nil {
		fmt.Printf("Controller Add() error: %v\n", err)
		return ModelAndView{
			Code: http.StatusBadRequest,
			Data: err.Error(),
		}
	}
	if memory.Ip == "" {
		host, _, splitErr := net.SplitHostPort(ctx.RemoteAddr())
		if splitErr == nil {
			memory.Ip = host
		} else {
			memory.Ip = ctx.RemoteAddr()
		}
	}
	if data := service.Add(memory); data != nil {
		return ModelAndView{
			Code: http.StatusOK,
			Data: data,
		}
	}
	return ModelAndView{
		Code: http.StatusBadGateway,
		Data: "服务器繁忙",
	}
}

func Reply(ctx iris.Context, service Services.IMemoriseService) ModelAndView {
	memory := Models.Memorise{}

	err := ctx.ReadForm(&memory)
	if err != nil {
		fmt.Printf("Controller Reply() error: %v\n", err)
		return ModelAndView{
			Code: http.StatusBadRequest,
			Data: err.Error(),
		}
	}
	code, data := service.Reply(memory)
	return ModelAndView{
		Code: code,
		Data: data,
	}
}

func Forget(ctx iris.Context, service Services.IMemoriseService) ModelAndView {
	memory := Models.Memorise{}

	err := ctx.ReadForm(&memory)
	if err != nil {
		fmt.Printf("Controller Forget() error: %v\n", err)
		return ModelAndView{
			Code: http.StatusBadRequest,
			Data: err.Error(),
		}
	}
	if flag := service.Forget(memory.Answer); flag {
		return ModelAndView{
			Code: http.StatusOK,
			Data: "success",
		}
	}
	return ModelAndView{
		Code: http.StatusBadGateway,
		Data: "服务器繁忙",
	}
}

func Status(service Services.IMemoriseService) ModelAndView {
	count := service.Status()
	return ModelAndView{
		Code: http.StatusOK,
		Data: count,
	}
}