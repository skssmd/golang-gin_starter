# Gin Framework Starter Template

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

# Control Panel API Documentation

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
Query Syntax
= : Equality (age=25)

~ : LIKE operator (username~john → username LIKE '%john%')

{ } : IN clause (id{1,2,3} → id IN (1,2,3))

( ) : BETWEEN (age(20,30) → age BETWEEN 20 AND 30)

& : AND operator (age=25&country=US)

| : OR operator (status=pending|status=approved)

Example Queries
Get users named "John" (LIKE):
GET /control/users/q/username~john

Get payments between 
100
a
n
d
100and500:
GET /control/Payment/q/Amount(100,500)

Combined conditions (AND/OR):
GET /control/users/q/age=30_|_country=US&role=admin
(Translates to (age=30) OR (country=US AND role=admin))

Postman Examples
Query with Multiple Conditions
Request:

Copy
GET /control/users/q/username~john_&_age(25,35)
SQL Equivalent:

sql
Copy
SELECT * FROM users 
WHERE (username LIKE '%john%') AND (age BETWEEN 25 AND 35)
Update Multiple Records
Request:

Copy
PUT /control/users/q/status=pending&role=user
Body:

json
Copy
{
  "status": "approved"
}
Action: Updates all users with status=pending AND role=user.

Notes
URL Encoding:
Encode special characters in query parameters (e.g., { → %7B, } → %7D).

Security:

Use caution with DELETE endpoints.

AllowGlobalUpdate: true is enabled for batch deletes.

Responses:

Successful operations return HTTP 200 with a JSON payload.

Errors include details in the error field.
