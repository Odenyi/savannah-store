package controllers

import (
	"database/sql"
	"net/http"
	"savannah-store/catalog-service/internal/models"
	"strconv"

	"github.com/labstack/echo/v4"
)

// CreateCategory inserts a new category
func CreateCategory(c echo.Context, db *sql.DB) error {

	req := new(models.CategoryRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	var parentExists bool
	if *req.ParentID != 0 {
		// Check if parent category exists
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE id = ?)", req.ParentID).Scan(&parentExists)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}

		if !parentExists {
			// Parent does not exist, ignore parent_id
			*req.ParentID = 0
		}
	}

	// Insert category
	var res sql.Result
	var err error
	if *req.ParentID != 0 {
		// Insert with parent_id
		res, err = db.Exec("INSERT INTO categories (name, parent_id) VALUES (?, ?)", req.Name, req.ParentID)
	} else {
		// Insert without parent_id
		res, err = db.Exec("INSERT INTO categories (name) VALUES (?)", req.Name)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	categoryID, _ := res.LastInsertId()
	return c.JSON(http.StatusCreated, echo.Map{
		"id":        categoryID,
		"name":      req.Name,
		"parent_id": req.ParentID, // 0 if not set
	})
}

func GetAveragePrice(c echo.Context, db *sql.DB) error {
	// Parse category ID from path
	catIDStr := c.Param("id")
	catID, err := strconv.Atoi(catIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid category id"})
	}

	// Recursive query with category name
	query := `
	WITH RECURSIVE category_hierarchy AS (
		SELECT id, name
		FROM categories
		WHERE id = ?

		UNION ALL

		SELECT c.id, c.name
		FROM categories c
		INNER JOIN category_hierarchy ch ON c.parent_id = ch.id
	)
	SELECT (SELECT name FROM categories WHERE id = ?) AS category_name,
	       AVG(p.price) AS avg_price
	FROM products p
	INNER JOIN category_hierarchy ch ON p.category_id = ch.id;
	`

	var avgPrice sql.NullFloat64
	var categoryName sql.NullString

	err = db.QueryRow(query, catID, catID).Scan(&categoryName, &avgPrice)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	price := 0.0
	if avgPrice.Valid {
		price = avgPrice.Float64
	}

	name := ""
	if categoryName.Valid {
		name = categoryName.String
	}

	return c.JSON(http.StatusOK, echo.Map{
		"category_id":   catID,
		"category_name": name,
		"average_price": price,
	})
}

// ViewCategories retrieves all categories
func ViewCategories(c echo.Context, db *sql.DB) error {
	rows, err := db.Query(`SELECT id, name, parent_id FROM categories`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var id int64
		var name string
		var parentID sql.NullInt64
		if err := rows.Scan(&id, &name, &parentID); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		categories = append(categories, echo.Map{
			"id":        id,
			"name":      name,
			"parent_id": parentID.Int64,
		})
	}

	return c.JSON(http.StatusOK, categories)
}

// UpdateCategory modifies a category
func UpdateCategory(c echo.Context, db *sql.DB) error {

	id := c.Param("id")
	req := new(models.CategoryRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	query := `UPDATE categories SET name = ?, parent_id = ? WHERE id = ?`
	_, err := db.Exec(query, req.Name, req.ParentID, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "category updated"})
}

// DeleteCategory removes a category
func DeleteCategory(c echo.Context, db *sql.DB) error {
	id := c.Param("id")

	// Check if category has children or products before deleting
	var childCount, productCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM categories WHERE parent_id = ?`, id).Scan(&childCount); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM products WHERE category_id = ?`, id).Scan(&productCount); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	if childCount > 0 || productCount > 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "cannot delete category with subcategories or products",
		})
	}

	_, err := db.Exec(`DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "category deleted"})
}

// CreateProduct inserts a new product
func CreateProduct(c echo.Context, db *sql.DB) error {

	req := new(models.ProductRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	query := `INSERT INTO products (name, price, category_id) VALUES (?, ?, ?)`
	res, err := db.Exec(query, req.Name, req.Price, req.CategoryID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	id, _ := res.LastInsertId()
	return c.JSON(http.StatusCreated, echo.Map{"id": id, "name": req.Name, "price": req.Price, "category_id": req.CategoryID})
}

// ViewProducts retrieves all products
func ViewProducts(c echo.Context, db *sql.DB) error {
	rows, err := db.Query(`SELECT id, name, price, category_id FROM products`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id int64
		var name string
		var price float64
		var categoryID int64
		if err := rows.Scan(&id, &name, &price, &categoryID); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		products = append(products, echo.Map{
			"id":          id,
			"name":        name,
			"price":       price,
			"category_id": categoryID,
		})
	}

	return c.JSON(http.StatusOK, products)
}

// UpdateProduct modifies a product
func UpdateProduct(c echo.Context, db *sql.DB) error {

	id := c.Param("id")
	req := new(models.ProductUpdateRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	query := `UPDATE products SET name = ?, price = ?, category_id = ? WHERE id = ?`
	_, err := db.Exec(query, req.Name, req.Price, req.CategoryID, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "product updated"})
}

// DeleteProduct removes a product
func DeleteProduct(c echo.Context, db *sql.DB) error {
	id := c.Param("id")

	_, err := db.Exec(`DELETE FROM products WHERE id = ?`, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "product deleted"})
}
