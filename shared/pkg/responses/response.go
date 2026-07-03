package response

import (
	"github.com/gofiber/fiber/v3"
)

//custom
func Response(c fiber.Ctx,code int,message string,data any,result bool)error{
	return c.Status(code).JSON(fiber.Map{
		"success": result,
		"message":message,
		"data":data,
	})
}

//sucess - 200
func OK(c fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

//created - 201
func Created(c fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

//accepted - 202
func Accepted(c fiber.Ctx,message string,data any)error{
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}


//not valid input - 400
func BadRequest(c fiber.Ctx, data any, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"data":    data,
		"error":   message,
	})
}


//internal error - 500
func InternalServerError(c fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"data":    nil,
		"error":   "An unexpected server error occurred. Please try again later.",
	})
}


//not loggedin - 401
func Unauthorized(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"success": false,
		"data":    nil,
		"error":   message, 
	})
}


//notfound 404
func NotFound(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}


//conflict - 409
func Conflict(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}


//dont have acess - 403
func Forbidden(c fiber.Ctx, data any, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"success": false,
		"data":    data,
		"error":   message,
	})
}


//too many requests - 429
func TooManyRequests(c fiber.Ctx, message string, retryAfter int) error {
    return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
        "success":    false,
        "error":      message,
        "retryAfter": retryAfter, // seconds — frontend drives the countdown from this
    })
}


//un processable input - 422
func UnprocessableEntity(c fiber.Ctx, message string) error {
    return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
        "success": false,
        "error":   message,
    })
}

