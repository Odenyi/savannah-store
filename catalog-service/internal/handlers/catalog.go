package handlers

import (
	
	"savannah-store/catalog-service/internal/controllers"

	"github.com/labstack/echo/v4"
)


// CreateCategory godoc
// @Summary      Create a category
// @Description  Creates a new catalog category
// @Tags         Catalog
// @Accept       json
// @Produce      json
// @Param        body  body  models.CategoryRequest  true  "Category info"
// @Success      201   {object} models.CategoryResponse
// @Failure      400   {object} map[string]string
// @Router       /catalog/categories [post]
func (a *App) CreateCategory(c echo.Context) error {
	return controllers.CreateCategory(c, a.DB)
}

// GetAveragePrice godoc
// @Summary Get average price of products in a category
// @Description Returns the average price of products for a given category, including subcategories
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{} "category_id, category_name, average_price"
// @Failure 400 {object} map[string]string "error message for invalid input"
// @Failure 500 {object} map[string]string "error message for server/database issues"
// @Router /categories/{id}/average-price [get]
func (a *App) GetAveragePrice(c echo.Context) error {
	return controllers.GetAveragePrice(c, a.DB)
}


// ViewCategories godoc
// @Summary      List categories
// @Description  Retrieves all catalog categories
// @Tags         Catalog
// @Produce      json
// @Success      200  {array} models.CategoryResponse
// @Router       /catalog/categories [get]
func (a *App) ViewCategories(c echo.Context) error {
	return controllers.ViewCategories(c, a.DB)
}

// UpdateCategory godoc
// @Summary      Update category
// @Description  Updates a catalog category by ID
// @Tags         Catalog
// @Accept       json
// @Produce      json
// @Param        id    path  int                        true  "Category ID"
// @Param        body  body  models.CategoryUpdateRequest true "Updated category info"
// @Success      200   {object} models.CategoryResponse
// @Failure      400   {object} map[string]string
// @Router       /catalog/categories/{id} [put]
func (a *App) UpdateCategory(c echo.Context) error {
	return controllers.UpdateCategory(c, a.DB)
}

// DeleteCategory godoc
// @Summary      Delete category
// @Description  Deletes a catalog category by ID
// @Tags         Catalog
// @Produce      json
// @Param        id    path  int  true  "Category ID"
// @Success      200   {object} map[string]string
// @Failure      404   {object} map[string]string
// @Router       /catalog/categories/{id} [delete]
func (a *App) DeleteCategory(c echo.Context) error {
	return controllers.DeleteCategory(c, a.DB)
}


// CreateProduct godoc
// @Summary      Create a product
// @Description  Creates a new catalog product
// @Tags         Catalog
// @Accept       json
// @Produce      json
// @Param        body  body  models.ProductRequest  true  "Product info"
// @Success      201   {object} models.ProductResponse
// @Failure      400   {object} map[string]string
// @Router       /catalog/products [post]
func (a *App) CreateProduct(c echo.Context) error {
	return controllers.CreateProduct(c, a.DB)
}

// ViewProducts godoc
// @Summary      List products
// @Description  Retrieves all catalog products
// @Tags         Catalog
// @Produce      json
// @Success      200  {array} models.ProductResponse
// @Router       /catalog/products [get]
func (a *App) ViewProducts(c echo.Context) error {
	return controllers.ViewProducts(c, a.DB)
}

// UpdateProduct godoc
// @Summary      Update product
// @Description  Updates a catalog product by ID
// @Tags         Catalog
// @Accept       json
// @Produce      json
// @Param        id    path  int                        true  "Product ID"
// @Param        body  body  models.ProductUpdateRequest true  "Updated product info"
// @Success      200   {object} models.ProductResponse
// @Failure      400   {object} map[string]string
// @Router       /catalog/products/{id} [put]
func (a *App) UpdateProduct(c echo.Context) error {
	return controllers.UpdateProduct(c, a.DB)
}

// DeleteProduct godoc
// @Summary      Delete product
// @Description  Deletes a catalog product by ID
// @Tags         Catalog
// @Produce      json
// @Param        id    path  int  true  "Product ID"
// @Success      200   {object} map[string]string
// @Failure      404   {object} map[string]string
// @Router       /catalog/products/{id} [delete]
func (a *App) DeleteProduct(c echo.Context) error {
	return controllers.DeleteProduct(c, a.DB)
}
