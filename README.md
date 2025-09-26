# Savannah Store Project

This project is an e-commerce platform built with **Go**, using a **microservices architecture**. It supports user authentication, product catalog management, order management, and notification services.  

---

## Prerequisites

The project uses the following technologies:

- **Go** for backend services
- **MySQL** as the primary relational database
- **Redis** for caching and fast access to cart items
- **RabbitMQ** for messaging between services (e.g., order notifications)
- **Swagger** for API documentation and testing
- **OAuth2/OpenID Connect** for authentication via Google

---

## Architecture

The project follows a **monorepository structure** with **four microservices**:

1. **Auth-Service**
   - Handles authentication and authorization
   - Implements OAuth2 with Google as the provider
   - Generates JWT tokens for API access

2. **Catalog-Service**
   - Manages products and categories
   - Supports hierarchical categories of arbitrary depth
   - Provides CRUD operations for products and categories
   - Computes average price for a given category

3. **Order-Service**
   - Manages shopping cart and orders
   - Stores cart items in Redis for fast access
   - Only admins can view or manage all user carts/orders; normal users can only manage their own
   - Sends order information via RabbitMQ for notifications

4. **Notification-Service**
   - Sends SMS and email notifications
   - Consumes messages from the order-service queue

---

## Deployment

The project is already deployed on a server. You can access the Swagger documentation for testing APIs:

- **Auth-Service:** [Swagger Docs](https://auth.vaslinkcomm.com/docs//index.html)
  - To generate an authorization token:
    - Navigate to: [Start Google Auth](https://auth.vaslinkcomm.com/docs//index.html#/Auth/post_auth_google_start)
    - Provide `phone` (optional) and `usertype` (`admin` or `customer`)
    - Follow the Google authentication link
    - Receive a JWT token for API authorization

- **Catalog-Service:** [Swagger Docs](https://catalog.vaslinkcomm.com/docs//index.html)
  - Admin users can add products and hierarchical categories
  - Normal users can view products without authentication
  - Hierarchy example:
    ```
    Electronics
      ├─ Phones
      │   ├─ Samsung
      │   ├─ iPhone
      │   └─ Techno
      ├─ Tablets
      └─ Laptops
    ```
  - Compute average price for a category

- **Order-Service:** [Swagger Docs](https://order.vaslinkcomm.com/docs//index.html)
  - Add, update, view, delete cart items
  - Place orders
  - Cart items stored in Redis for fast retrieval
  - Orders sent via RabbitMQ to notification service

---

## Workflow Diagram

Here is an overview of the service interactions:

```mermaid
flowchart LR
    A[User] -->|Login via Google| Auth[Auth-Service]
    Auth -->|JWT Token| User
    User -->|Add/View Products| Catalog[Catalog-Service]
    User -->|Add to Cart| Order[Order-Service]
    Order -->|Publish Order Message| RabbitMQ[Message Queue]
    RabbitMQ --> Notification[Notification-Service]
    Notification -->|SMS/Email| User


