# AutoTracer Project Guidelines for Claude Code

## Project Overview
AutoTracer is an automated observability instrumentation tool that uses AI to identify and instrument business logic in Go applications with OpenTelemetry.

## Key Concepts

### Business Logic vs Infrastructure
- **Business Logic**: Code that implements domain-specific rules, workflows, and processes
  - Examples: payment processing, order validation, pricing calculations, user authentication
  - Characteristics: Contains business rules, makes business decisions, handles domain entities
  
- **Infrastructure Code**: Technical implementation details
  - Examples: HTTP middleware, database connections, logging utilities, error handlers
  - Characteristics: Generic, reusable, not business-specific

## Instrumentation Patterns

### 1. Span Creation Pattern
```go
ctx, span := tracer.Start(ctx, "OperationName",
    trace.WithAttributes(
        // Include business-relevant attributes
        attribute.String("entity.id", id),
        attribute.String("entity.type", entityType),
    ),
)
defer span.End()
```

### 2. Error Recording Pattern
```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

### 3. Business Metrics Pattern
```go
// Add business events
span.AddEvent("payment.processed", trace.WithAttributes(
    attribute.Float64("amount", amount),
    attribute.String("currency", currency),
))
```

## Business Logic Identification Rules

### High Priority (Always Instrument)
1. **Transaction Processing**: Any code handling money, orders, or financial calculations
2. **User Actions**: Registration, login, profile updates, preferences
3. **Core Business Workflows**: Checkout, booking, subscription management
4. **External Integrations**: Payment gateways, third-party APIs, partner systems

### Medium Priority (Selective Instrumentation)
1. **Validation Logic**: Complex business rule validation
2. **Authorization**: Role-based access control, permission checks
3. **Data Transformations**: Business entity conversions, calculations

### Low Priority (Usually Skip)
1. **CRUD Operations**: Simple database operations without business logic
2. **Utilities**: Helpers, formatters, simple converters
3. **Infrastructure**: Middleware, routers, basic handlers

## Code Analysis Guidelines

When analyzing Go code:

1. **Look for Business Verbs**: Process, Calculate, Validate, Authorize, Transform
2. **Identify Business Entities**: Order, Payment, User, Product, Subscription
3. **Find Critical Paths**: Follow data flow from API entry to business outcome
4. **Detect Async Operations**: Goroutines handling business logic need special attention

## Instrumentation Best Practices

### DO:
- ✅ Create spans for complete business transactions
- ✅ Include business context in span attributes
- ✅ Propagate context through all async operations
- ✅ Record business events and outcomes
- ✅ Use semantic naming for operations

### DON'T:
- ❌ Instrument every function call
- ❌ Include sensitive data (passwords, credit cards) in spans
- ❌ Create spans for simple getters/setters
- ❌ Forget to end spans (use defer)
- ❌ Over-instrument infrastructure code

## Example Analysis Output

When you analyze code, provide output like this:

```markdown
## Business Logic Flows Identified

### 1. Order Processing Flow
- **Entry Point**: `POST /api/orders` → `CreateOrderHandler`
- **Business Logic**: Order validation → Inventory check → Payment processing → Order confirmation
- **Critical Functions**:
  - `ValidateOrder()` - Validates business rules
  - `CheckInventory()` - Ensures product availability  
  - `ProcessPayment()` - Handles payment transaction
  - `ConfirmOrder()` - Finalizes order and sends notifications

### 2. User Authentication Flow
- **Entry Point**: `POST /api/login` → `LoginHandler`
- **Business Logic**: Credential validation → Token generation → Session creation
- **Critical Functions**:
  - `ValidateCredentials()` - Authenticates user
  - `GenerateToken()` - Creates JWT token
  - `CreateSession()` - Establishes user session
```

## Testing Considerations

When adding instrumentation:
1. Ensure spans are properly closed in all code paths
2. Verify context propagation in concurrent operations
3. Check that span attributes don't expose sensitive data
4. Validate that instrumentation doesn't impact performance significantly

## Output Format

Always create:
1. Modified `.go` files with instrumentation
2. `instrumentation_report.md` with detailed analysis
3. Optional: `instrumentation_test.go` for testing the instrumentation