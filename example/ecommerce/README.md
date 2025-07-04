# E-Commerce Microservice Example

This is a sample e-commerce microservice designed to test AutoTracer's ability to identify and instrument business logic with OpenTelemetry.

## Architecture

The application consists of several interconnected services:

### Services

1. **Payment Service** (`internal/payment/`)
   - Processes payments through payment gateway
   - Validates payment methods
   - Handles refunds and rollbacks
   - Calculates processing fees

2. **Order Service** (`internal/order/`)
   - Creates and manages orders
   - Validates order requests
   - Coordinates with payment and inventory services
   - Handles order fulfillment

3. **Inventory Service** (`internal/inventory/`)
   - Manages product stock
   - Reserves inventory during checkout
   - Handles stock updates
   - Sends low stock alerts

4. **User Service** (`internal/user/`)
   - User registration and authentication
   - Profile management
   - Password hashing and token generation
   - Login tracking

5. **Notification Service** (`internal/notification/`)
   - Sends email and SMS notifications
   - Queues messages for delivery
   - Handles delivery retries
   - Supports multiple channels

## Business Logic Patterns

This example includes various business logic patterns that AutoTracer should identify:

- **Transaction Processing**: Payment processing with rollback capabilities
- **Inventory Management**: Stock checking, reservation, and deduction
- **User Authentication**: Registration, login, and token generation
- **Asynchronous Operations**: Background notifications and order fulfillment
- **Error Handling**: Validation, rollbacks, and retry mechanisms
- **Business Rules**: Payment method validation, inventory checks, user validation

## API Endpoints

- `POST /api/orders` - Create a new order
- `GET /api/orders/{id}` - Get order details
- `POST /api/users/register` - Register new user
- `POST /api/users/login` - User login
- `PUT /api/users/{id}/profile` - Update user profile

## Running the Application

```bash
cd example/ecommerce
go mod download
go run main.go
```

The server will start on port 8080.

## Testing

Run the integration tests:

```bash
go test ./tests/integration -v
```

## AutoTracer Expected Behavior

When AutoTracer analyzes this codebase, it should:

1. Identify critical business logic flows
2. Add OpenTelemetry spans to trace operations
3. Include relevant business context in span attributes
4. Properly propagate context through async operations
5. Handle error scenarios with appropriate span status
6. Avoid over-instrumenting infrastructure code