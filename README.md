# Gin Framework Starter

A starter template built with the **Gin** framework, integrated with:

- **User Authentication** (with email verification)
- **Google Authentication**
- **Stripe Payment Integration**
- **API-based Admin Panel** for easy management of models
- Supports **PostgreSQL** & **SQLite** databases

## Features

- **Email Verification:** Automatically send verification emails upon user registration.
- **Google Authentication:** Sign in with Google for easy access.
- **Stripe Integration:** Payment processing for subscriptions and more.
- **Admin Panel:** Manage and monitor your models via an intuitive API interface.
- **Database Support:** Use PostgreSQL or SQLite based on your preference.


#Control Panel
###Register the Model in the Control Panel
In the Routes function (usually located in controlpanel/process.go), register your model using the Register method of the Control struct.
```go
func Routes(r *gin.Engine) {
	controlPanel := NewControl(initials.DB)

	// Register Models
	controlPanel.Register("users", auth.User{})
	controlPanel.Register("products", Product{}) // Register your new model here

	SetupControlRoutes(r, controlPanel)
}
```
### Control Panel API 

This documentation provides guidance on using the Control Panel API endpoints with Postman. The API allows you to manage registered models (CRUD operations) and perform advanced queries.

---

## Base URL For ControlPanel
All endpoints start with:  
`http://localhost:8080/control`

---

## Authentication
- All endpoints require authentication.
- Ensure you have a valid session cookie or JWT token set in your requests.

---

## Endpoints Overview

| Method | Endpoint                     | Description                          |
|--------|------------------------------|--------------------------------------|
| GET    | `/`                          | List all registered models          |
| GET    | `/:model`                    | Get all instances of a model        |
| DELETE | `/:model/delete`             | Delete all instances of a model     |
| GET    | `/:model/:id`                | Get a specific instance             |
| PUT    | `/:model/:id`                | Update a specific instance          |
| DELETE | `/:model/:id`                | Delete a specific instance          |
| GET    | `/:model/q/:query`           | Query instances (GET)               |
| PUT    | `/:model/q/:query`           | Query & update instances            |
| DELETE | `/:model/q/:query`           | Query & delete instances            |

---

## 1. List Registered Models

**Endpoint**  
`GET /control/`

**Response**  
Returns a JSON object with model names as keys and their fields as values.

**Example:**
```json
{
  "users": ["ID", "Username", "Email"],
  "paymentMethod": ["ID", "Name", "Type"],
  "Payment": ["ID", "Amount", "Date"]
}
````
## 2. Get All Instances of a Model
Endpoint
````
GET /control/:model
````
Example:
````
GET /control/users
````
Response
Returns an array of all instances in the specified model.

##3. Delete All Instances of a Model
Endpoint
````
DELETE /control/:model/delete
````
Example:
````
DELETE /control/users/delete
````
Response

```json

{
  "message": "All instances of users deleted"
}
````
## 4. Get/Update/Delete Specific Instance
Get Instance
````
GET /control/users/1
````
Returns the user with ID=1.

Update Instance
````
PUT /control/users/1
````
Body:

````json

{
  "Username": "new_username"
}
````
Delete Instance
````
DELETE /control/users/1
````
## 5. Advanced Query Endpoint (/control/:model/q/:query)


Perform advanced queries on model instances using a flexible syntax. Supports `GET`, `PUT`, and `DELETE` methods.

---

## Syntax

| Operator | Description                          | Example                     | SQL Equivalent                  |
|----------|--------------------------------------|-----------------------------|---------------------------------|
| `=`      | Equality                             | `age=25`                    | `age = 25`                     |
| `~`      | `LIKE` operator (wildcard search)    | `username~john`             | `username LIKE '%john%'`       |
| `{ }`    | `IN` clause (multiple values)        | `id{1,2,3}`                 | `id IN (1, 2, 3)`              |
| `( )`    | `BETWEEN` (range)                    | `age(20,30)`                | `age BETWEEN 20 AND 30`        |
| `&`      | Logical `AND` (within a group)       | `age=25&country=US`         | `age = 25 AND country = 'US'`  |
| `\|`     | Logical `OR` (between groups)        | `status=pending\|status=approved` | `(status = 'pending') OR (status = 'approved')` |

---

## Examples

### 1. Get Users with a Specific Username (LIKE)
Request
````
GET /control/users/q/username~john
````
SQL Equivalent
````sql

SELECT * FROM users WHERE username LIKE '%john%'
Response
//Returns all users with a username containing john.
````

### 2. Get Verified Users
Request
````
GET /control/users/q/IsVerified=true
````
SQL Equivalent
````sql
SELECT * FROM users WHERE is_verified = true
// Returns all users where IsVerified is true.
````
## 3. Get Users with a Specific First Name and Last Name
Request
````
GET /control/users/q/FirstName=John_&_LastName=Doe
````
SQL Equivalent
````sql
SELECT * FROM users 
WHERE first_name = 'John' AND last_name = 'Doe'
//Returns users with FirstName = "John" and LastName = "Doe".
````
## 4. Get Users Created Between Two Dates
Request
````
GET /control/users/q/CreatedAt(2023-01-01,2023-12-31)
````
SQL Equivalent
````
SELECT * FROM users 
WHERE created_at BETWEEN '2023-01-01' AND '2023-12-31'
//Returns users created between January 1, 2023, and December 31, 2023.
````
## 5. Get Admin or Superuser Users
Request
````
GET /control/users/q/IsAdmin=true_|_IsSuperuser=true
````
SQL Equivalent
```sql
SELECT * FROM users 
WHERE is_admin = true OR is_superuser = true
//Returns users who are either admins or superusers.
````
#6. Update Multiple Users (Set as Verified)
Request
````
PUT /control/users/q/IsVerified=false
````
Body
````json
Copy
{
  "IsVerified": true
}
````
Action
Updates all users where IsVerified = false to IsVerified = true.

##7. Delete Unverified Users
Request
````
DELETE /control/users/q/IsVerified=false
````
Action
Deletes all users where IsVerified = false.
