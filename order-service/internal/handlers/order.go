package handlers

import (
	"net/http"
	"savannah-store/order-service/internal/controllers"
	"savannah-store/order-service/internal/models"

	"github.com/labstack/echo/v4"
)

// AddToCart godoc
// @Summary      Add item to cart
// @Description  Adds a product to the user cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Param        api-key header string true "API Key"
// @Param        body  body  models.AddToCartRequest  true  "Cart item"
// @Success      201   {object} map[string]interface{}
// @Failure      400   {object} map[string]string
// @Failure      401   {object} map[string]string
// @Router       /cart [post]
func (a *App) AddToCart(c echo.Context) error {
	claims := c.Get("claims").(*models.JwtCustomClaims) // Get user ID from JWT
	req := new(models.CartItem)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	// Override any UserID sent by client
	req.UserID = claims.UserID

	return controllers.AddToCart(c, a.DB, a.RedisConnection, req)
}


// ViewCart godoc
// @Summary      View cart
// @Description  Retrieves all items from the user's cart.
//               Admins see all carts, regular users see only their own.
// @Tags         Cart
// @Produce      json
// @Param        api-key header string true "API Key"
// @Success      200  {array}  map[string]interface{}
// @Failure      401  {object} map[string]string "unauthorized or invalid token"
// @Failure      500  {object} map[string]string "server error"
// @Router       /cart [get]
func (a *App) ViewCart(c echo.Context) error {
	claims := c.Get("claims").(*models.JwtCustomClaims)
	role := c.Get("role").(string)

	// Pass claims.UserID and role to controller
	return controllers.ViewCart(c, a.DB, a.RedisConnection, claims.UserID, role)
}

// UpdateCart godoc
// @Summary      Update cart
// @Description  Updates quantity of an item in the cart. Users can update their own cart. Admins can update any user's cart by specifying user_id in request body.
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Param        body  body  models.UpdateCartRequest  true  "Updated cart item"
// @Success      200   {object} map[string]interface{}
// @Failure      400   {object} map[string]string
// @Param        api-key header string true "API Key"
// @Router       /cart [put]
func (a *App) UpdateCart(c echo.Context) error {
	return controllers.UpdateCart(c, a.DB, a.RedisConnection)
}

// DeleteCart godoc
// @Summary      Delete item from cart
// @Description  Removes a product from the cart. Users can delete items from their own cart. Admins can delete items from any user's cart by specifying user_id in query.
// @Tags         Cart
// @Produce      json
// @Param        product_id query int true "Product ID"
// @Param        user_id    query int false "User ID (admin only)"
// @Success      200 {object} map[string]interface{}
// @Failure      404 {object} map[string]string
// @Param        api-key header string true "API Key"
// @Router       /cart [delete]
func (a *App) DeleteCart(c echo.Context) error {
	return controllers.DeleteCart(c, a.DB, a.RedisConnection)
}

// PlaceOrder godoc
// @Summary      Place an order
// @Description  Places a new order for the user. Admins can place orders for other users by specifying user_id in request body.
// @Tags         Order
// @Accept       json
// @Produce      json
// @Param        body  body  models.PlaceOrderRequest  true  "Order details"
// @Success      201   {object} map[string]interface{}
// @Failure      400   {object} map[string]string
// @Param        api-key header string true "API Key"
// @Router       /orders [post]
func (a *App) PlaceOrder(c echo.Context) error {
	return controllers.PlaceOrder(c, a.DB, a.RedisConnection,a.RabbitMQConn)
}

// ViewOrders godoc
// @Summary      View orders
// @Description  Retrieves all orders for the user. Admins can view all orders or specify a user_id query to view orders of a specific user.
// @Tags         Order
// @Produce      json
// @Param        user_id query int false "User ID (admin only)"
// @Success      200 {array} map[string]interface{}
// @Param        api-key header string true "API Key"
// @Router       /orders [get]
func (a *App) ViewOrders(c echo.Context) error {
	return controllers.ViewOrders(c, a.DB)
}

// DeleteOrder godoc
// @Summary      Delete order
// @Description  Deletes an order by ID. Users can delete their own orders. Admins can delete any order.
// @Tags         Order
// @Produce      json
// @Param        order_id query int true "Order ID"
// @Param        user_id  query int false "User ID (admin only)"
// @Success      200 {object} map[string]interface{}
// @Failure      404 {object} map[string]string
// @Param        api-key header string true "API Key"
// @Router       /orders [delete]
func (a *App) DeleteOrder(c echo.Context) error {
	return controllers.DeleteOrder(c, a.DB)
}
