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


# Authentication Middleware and Helper Methods
This provides an overview of the predefined authentication middlewares and helper methods available in the auth package. These methods are used to enforce authentication, verification, and role-based access control in your application.

##Middlewares and Methods Overview
###Middlewares

	LoginRequired
		Ensures the user is authenticated by validating the JWT token.

 	RequireVerification
		Ensures the user is verified (if verification is enabled).
	
 	RequireAdmin
		Ensures the user is an admin.
	
 	RequireSuperuser
		Ensures the user is a superuser.

###Helper Methods
	GetUser
		Retrieves the authenticated user from the context.
	IsVerified
		Checks if the user is verified.
	IsAdmin
		Checks if the user is an admin.
	IsSuperuser
		Checks if the user is a superuser.

##Use Cases and Examples
### 1. Enforcing Authentication (LoginRequired)
Use this middleware to protect routes that require authentication.

Example:
````go
r.GET("/protected", auth.LoginRequired, func(c *gin.Context) {
    user := auth.GetUser(c)
    c.JSON(http.StatusOK, gin.H{"message": "Welcome, " + user.Username})
})
````
Behavior:
If the user is not authenticated, they will receive a 401 Unauthorized response.

If authenticated, the user is attached to the context for further use.

###2. Enforcing Verification (RequireVerification)
Use this middleware to ensure the user is verified (if verification is enabled in the environment).

Example:
````go
r.GET("/verified-only", auth.RequireVerification, func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"message": "You are verified!"})
})
````
Behavior:
If verification is enabled (VERIFICATION=true) and the user is not verified, they will receive a 403 Forbidden response.

If verified, the request proceeds.

3. Enforcing Admin Access (RequireAdmin)
Use this middleware to restrict access to admin users.

Example:
````go
r.GET("/admin-only", auth.RequireAdmin, func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"message": "Welcome, Admin!"})
})
````
Behavior:
If the user is not an admin, they will receive a 403 Forbidden response.

If the user is an admin, the request proceeds.

###4. Enforcing Superuser Access (RequireSuperuser)
Use this middleware to restrict access to superusers.

Example:
````go
r.GET("/superuser-only", auth.RequireSuperuser, func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"message": "Welcome, Superuser!"})
})
````
Behavior:
If the user is not a superuser, they will receive a 403 Forbidden response.

If the user is a superuser, the request proceeds.

### 5. Checking Admin Status (IsAdmin)
Use this method to check if the user is an admin.

Example:
````go
r.GET("/check-admin", auth.LoginRequired, func(c *gin.Context) {
    user := auth.GetUser(c)
    if auth.IsAdmin(c, user) {
        c.JSON(http.StatusOK, gin.H{"message": "User is an admin"})
    }
})
````
Behavior:
If the user is not an admin, a 403 Forbidden response is sent.

If the user is an admin, the request proceeds.

###8. Checking Superuser Status (IsSuperuser)
Use this method to check if the user is a superuser.

Example:
````go

r.GET("/check-superuser", auth.LoginRequired, func(c *gin.Context) {
    user := auth.GetUser(c)
    if auth.IsSuperuser(c, user) {
        c.JSON(http.StatusOK, gin.H{"message": "User is a superuser"})
    }
})
````
Behavior:
If the user is not a superuser, a 403 Forbidden response is sent.

If the user is a superuser, the request proceeds.

##Environment Variables
SECRET: The secret key used to sign and verify JWT tokens.

VERIFICATION: Set to "true" to enable verification checks.

##Example Workflow
User logs in:

A JWT token is generated and sent to the client (e.g., in a cookie or response body).

User accesses a protected route:

The LoginRequired middleware validates the token and attaches the user to the context.

User accesses an admin-only route:

The RequireAdmin middleware checks if the user is an admin. If not, access is denied.

User accesses a verified-only route:

The RequireVerification middleware checks if the user is verified. If not, access is denied.

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


# Advanced Query Endpoint (`/q/:query`)

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
| `_&_`    | Groups conditions with **AND**       | `group1_&_group2`           | `(group1) AND (group2)`         |
| `_|_`    | Groups conditions with **OR**        | `group1_|_group2`           | `(group1) OR (group2)`          |

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
##8. Grouping Conditions with _&_ and _|_
Request
````
GET /control/users/q/username~jo_|_id(1,5)_&_first_name=john&last_name~tidor
````
SQL Equivalent

````sql

SELECT * FROM users 
WHERE (
  (username LIKE '%jo%' OR id BETWEEN 1 AND 5)
) AND (
  (first_name = 'john' AND last_name LIKE '%tidor%')
````
