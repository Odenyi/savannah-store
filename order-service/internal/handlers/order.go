package handlers

import (
	
	"savannah-store/order-service/internal/controllers"

	"github.com/labstack/echo/v4"
)


// AddToCart godoc
// @Summary      Add item to cart
// @Description  Adds a product to the user cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Param        body  body  models.AddToCartRequest  true  "Cart item"
// @Success      200   {object} map[string]interface{}
// @Failure      400   {object} map[string]string
// @Router       /cart [post]
func (a *App) AddToCart(c echo.Context) error {
	return controllers.AddToCart(c, a.DB, a.RedisConnection)
}

// ViewCart godoc
// @Summary      View cart
// @Description  Retrieves all items from the user cart
// @Tags         Cart
// @Produce      json
// @Success      200  {array}  map[string]interface{}
// @Router       /cart [get]
func (a *App) ViewCart(c echo.Context) error {
	return controllers.ViewCart(c, a.DB, a.RedisConnection)
}

// UpdateCart godoc
// @Summary      Update cart
// @Description  Updates quantity of an item in the cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Param        body  body  models.UpdateCartRequest  true  "Updated cart item"
// @Success      200   {object} map[string]interface{}
// @Failure      400   {object} map[string]string
// @Router       /cart [put]
func (a *App) UpdateCart(c echo.Context) error {
	return controllers.UpdateCart(c, a.DB, a.RedisConnection)
}

// DeleteCart godoc
// @Summary      Delete item from cart
// @Description  Removes a product from the cart
// @Tags         Cart
// @Produce      json
// @Param        product_id query int true "Product ID"
// @Success      200 {object} map[string]interface{}
// @Failure      404 {object} map[string]string
// @Router       /cart [delete]
func (a *App) DeleteCart(c echo.Context) error {
	return controllers.DeleteCart(c, a.DB, a.RedisConnection)
}


// PlaceOrder godoc
// @Summary      Place an order
// @Description  Places a new order for the user
// @Tags         Order
// @Accept       json
// @Produce      json
// @Param        body  body  models.PlaceOrderRequest  true  "Order details"
// @Success      201   {object} map[string]interface{}
// @Failure      400   {object} map[string]string
// @Router       /orders [post]
func (a *App) PlaceOrder(c echo.Context) error {
	return controllers.PlaceOrder(c, a.DB, a.RedisConnection,a.RabbitMQConn)
}

// ViewOrders godoc
// @Summary      View orders
// @Description  Retrieves all orders for the user
// @Tags         Order
// @Produce      json
// @Success      200 {array} map[string]interface{}
// @Router       /orders [get]
func (a *App) ViewOrders(c echo.Context) error {
	return controllers.ViewOrders(c, a.DB)
}

// DeleteOrder godoc
// @Summary      Delete order
// @Description  Deletes an order by ID
// @Tags         Order
// @Produce      json
// @Param        order_id query int true "Order ID"
// @Success      200 {object} map[string]interface{}
// @Failure      404 {object} map[string]string
// @Router       /orders [delete]
func (a *App) DeleteOrder(c echo.Context) error {
	return controllers.DeleteOrder(c, a.DB)
}
