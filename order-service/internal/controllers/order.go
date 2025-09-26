package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis"

	"savannah-store/order-service/internal/library"
	"savannah-store/order-service/internal/models"

	"time"

	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Add item to cart (Redis)
func AddToCart(c echo.Context, db *sql.DB, redisConn *redis.Client, req *models.CartItem) error {
	// Validate that product exists and fetch current price
	var exists int
	var currentPrice float64
	err := db.QueryRow(`SELECT COUNT(*), price FROM catalogdb.products WHERE id = ?`, req.ProductID).Scan(&exists, &currentPrice)
	if err != nil || exists == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "product does not exist"})
	}

	// Force the correct price from catalog
	req.Price = currentPrice

	// Redis key per user
	key := fmt.Sprintf("cart:%v", req.UserID)

	// Get existing cart
	_, cartData := library.GetAllKeys(redisConn, key+"*")

	// Convert cartData (map) -> slice
	var cart []models.CartItem
	for _, v := range cartData {
		var item models.CartItem
		_ = json.Unmarshal([]byte(v), &item)
		cart = append(cart, item)
	}

	// Add or update product in cart
	found := false
	for i, item := range cart {
		if item.ProductID == req.ProductID {
			cart[i].Quantity += req.Quantity
			cart[i].Price = currentPrice // Update price in case it changed
			found = true
			break
		}
	}
	if !found {
		cart = append(cart, *req)
	}

	// Save back to Redis
	for _, item := range cart {
		val, _ := json.Marshal(item)
		library.SetRedisKey(redisConn, fmt.Sprintf("%s:%d", key, item.ProductID), string(val))
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "item added to cart"})
}

// View Cart
func ViewCart(c echo.Context, db *sql.DB, redisConn *redis.Client, userID int64, role string) error {

	results := map[string]string{}

	if role == "admin" {
		// Admin: fetch all carts
		err, allCarts := library.GetAllKeys(redisConn, "cart:*")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch carts"})
		}
		results = allCarts
	} else {
		// Normal user: fetch only their cart
		err, userCart := library.GetAllKeys(redisConn, fmt.Sprintf("cart:%d*", userID))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch cart"})
		}
		results = userCart
	}

	var cart []models.CartItem
	for _, v := range results {
		var item models.CartItem
		if err := json.Unmarshal([]byte(v), &item); err != nil {
			continue
		}
		cart = append(cart, item)
	}

	return c.JSON(http.StatusOK, cart)
}

// Update cart item
func UpdateCart(c echo.Context, db *sql.DB, redisConn *redis.Client) error {
	claims := c.Get("claims").(*models.JwtCustomClaims)
	role := c.Get("role").(string)

	productID := c.Param("id") // assuming /cart/:id

	req := new(models.UpdateCartRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	userID := claims.UserID
	if role == "admin" && req.UserID != 0 {
		userID = req.UserID
	}

	key := fmt.Sprintf("cart:%v:%v", userID, productID)
	val, _ := json.Marshal(req)
	if err := library.SetRedisKey(redisConn, key, string(val)); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "cart updated"})
}

// Delete cart item
func DeleteCart(c echo.Context, db *sql.DB, redisConn *redis.Client) error {
	claims := c.Get("claims").(*models.JwtCustomClaims)
	role := c.Get("role").(string)

	productID := c.Param("id") // /cart/:id
	userID := claims.UserID

	// Admin can delete for another user
	reqUserID := c.QueryParam("user_id")
	if role == "admin" && reqUserID != "" {
		userID = library.ParseUserID(reqUserID)
	}

	key := fmt.Sprintf("cart:%v:%v", userID, productID)
	if err := library.DeleteRedisKey(redisConn, key); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "item deleted"})
}

// Place Order
func PlaceOrder(c echo.Context, db *sql.DB, redisConn *redis.Client, mq *amqp.Connection) error {
	claims := c.Get("claims").(*models.JwtCustomClaims)
	role := c.Get("role").(string)

	userID := claims.UserID
	if role == "admin" {
		reqUserID := c.Param("user_id")
		if reqUserID != "" {
			userID = library.ParseUserID(reqUserID)
		}
	}

	cartKeyPattern := fmt.Sprintf("cart:%v:*", userID)

	// Fetch cart items
	keys, _ := redisConn.Keys(cartKeyPattern).Result()
	if len(keys) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "cart is empty"})
	}

	var items []models.CartItem
	var total float64
	for _, key := range keys {
		data, err := library.GetRedisKey(redisConn, key)
		if err != nil {
			continue
		}
		var item models.CartItem
		if err := json.Unmarshal([]byte(data), &item); err != nil {
			continue
		}
		total += float64(item.Quantity) * item.Price
		items = append(items, item)
	}

	// Insert order
	res, err := db.Exec(`INSERT INTO orders (user_id, total_amount, status, created_at) VALUES (?, ?, ?, ?)`,
		userID, total, "Pending", time.Now())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	orderID, _ := res.LastInsertId()

	// Insert order items
	for _, item := range items {
		_, err := db.Exec(`INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)`,
			orderID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
	}

	// Clear cart
	for _, key := range keys {
		_ = library.DeleteRedisKey(redisConn, key)
	}

	go func() {
		_ = SendSMS(db, mq, userID, orderID)
		_ = SendEmailToAdmin(db, mq, orderID, total, items)
	}()

	return c.JSON(http.StatusCreated, echo.Map{"order_id": orderID, "total": total, "status": "Pending"})
}

// ViewOrders retrieves all orders for a specific user
func ViewOrders(c echo.Context, db *sql.DB) error {
	claims := c.Get("claims").(*models.JwtCustomClaims)
	role := c.Get("role").(string)

	userID := claims.UserID
	if role == "admin" {
		reqUserID := c.Param("user_id")
		if reqUserID != "" {
			userID = library.ParseUserID(reqUserID)
		}
	}

	rows, err := db.Query(`SELECT id, user_id, total_amount, status, created_at FROM orders WHERE user_id = ?`, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	defer rows.Close()

	var orders []map[string]interface{}
	for rows.Next() {
		var id, uid int64
		var total float64
		var status, createdAt string

		if err := rows.Scan(&id, &uid, &total, &status, &createdAt); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}

		orders = append(orders, echo.Map{
			"id":           id,
			"user_id":      uid,
			"total_amount": total,
			"status":       status,
			"created_at":   createdAt,
		})
	}

	return c.JSON(http.StatusOK, orders)
}

// DeleteOrder removes an order by ID
func DeleteOrder(c echo.Context, db *sql.DB) error {
	claims := c.Get("claims").(*models.JwtCustomClaims)
	role := c.Get("role").(string)

	orderID := c.Param("id")

	// Admin can delete any order, users only their own
	if role != "admin" {
		var userID int64
		err := db.QueryRow(`SELECT user_id FROM orders WHERE id = ?`, orderID).Scan(&userID)
		if err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "order not found"})
		}
		if userID != claims.UserID {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "not authorized"})
		}
	}

	_, err := db.Exec(`DELETE FROM orders WHERE id = ?`, orderID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "order deleted"})
}

func SendSMS(db *sql.DB, rabbitConn *amqp.Connection, userID, orderID int64) error {
	// Fetch user phone from DB
	var phone string
	err := db.QueryRow(`SELECT phone FROM authdb.users WHERE id = ?`, userID).Scan(&phone)
	if err != nil {
		log.Printf("Failed to fetch user phone for userID %d: %v\n", userID, err)
		return err
	}

	if phone == "" {
		return fmt.Errorf("user %d has no phone number", userID)
	}

	// Prepare SMS notification
	message := fmt.Sprintf("Your order #%d has been placed successfully!", orderID)
	notif := models.Notification{
		UserID:  userID,
		Type:    "sms",
		To:      phone,
		Message: message,
	}

	// Push to notification queue
	if err := library.Notification(rabbitConn, notif); err != nil {
		log.Printf("Failed to push SMS notification to queue for user %d: %v\n", userID, err)
		return err
	}

	return nil
}

func SendEmailToAdmin(db *sql.DB, rabbitConn *amqp.Connection, orderID int64, total float64, items []models.CartItem) error {
	// Fetch admin emails (could be multiple)
	rows, err := db.Query(`
		SELECT u.email, 
		FROM authdb.users u
		JOIN authdb.roles r ON u.role_id = r.id
		WHERE r.name = 'admin'
	`)
	if err != nil {
		log.Println("Failed to fetch admin emails:", err)
		return err
	}
	defer rows.Close()

	var adminEmails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			log.Println("Failed to scan admin email:", err)
			continue
		}
		adminEmails = append(adminEmails, email)
	}

	if len(adminEmails) == 0 {
		return fmt.Errorf("no admin emails found")
	}

	// Prepare order details
	itemDetails := ""
	for _, item := range items {
		itemDetails += fmt.Sprintf("Product ID: %d, Quantity: %d, Price: %.2f\n", item.ProductID, item.Quantity, item.Price)
	}
	subject := fmt.Sprintf("New Order Placed: #%d", orderID)
	message := fmt.Sprintf("A new order has been placed.\n\nOrder ID: %d\nTotal Amount: %.2f\n\nItems:\n%s", orderID, total, itemDetails)

	// Send notification to each admin
	for _, adminEmail := range adminEmails {
		notif := models.Notification{
			Type:    "email",
			To:      adminEmail,
			Subject: subject,
			Message: message,
		}
		if err := library.Notification(rabbitConn, notif); err != nil {
			log.Printf("Failed to send notification to %s: %v\n", adminEmail, err)
		}
	}

	return nil
}
