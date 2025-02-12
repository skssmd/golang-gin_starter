package controlPanel

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"

	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Control struct to manage registered models
type Control struct {
	DB     *gorm.DB
	Models map[string]interface{}
}

// NewControl creates a new Control instance
func NewControl(db *gorm.DB) *Control {
	return &Control{
		DB:     db,
		Models: make(map[string]interface{}),
	}
}

// Register a model in the Control panel
func (a *Control) Register(name string, model interface{}) {
	a.Models[name] = model
}

// Get registered models and their fields
func (a *Control) GetModels(c *gin.Context) {
	modelDetails := make(map[string][]string)

	// Loop through the models and extract their field names using reflection
	for name, model := range a.Models {
		var fields []string
		modelType := reflect.TypeOf(model)

		// Iterate over the fields of the struct
		for i := 0; i < modelType.NumField(); i++ {
			field := modelType.Field(i)
			fields = append(fields, field.Name)
		}

		// Add model name and fields to the map
		modelDetails[name] = fields
	}

	// Return the model details in the requested format
	c.JSON(http.StatusOK, modelDetails)
}

// Get all instances of a model
func (a *Control) GetModelInstances(c *gin.Context) {
	modelName := c.Param("model")
	model, exists := a.Models[modelName]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	// Create a slice of the model type
	modelSlice := reflect.New(reflect.SliceOf(reflect.TypeOf(model))).Interface()

	if err := a.DB.Find(modelSlice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch instances"})
		return
	}

	c.JSON(http.StatusOK, modelSlice)
}

// Delete all instances of a model
func (a *Control) DeleteModelInstances(c *gin.Context) {
	modelName := c.Param("model")
	model, exists := a.Models[modelName]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	if err := a.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete instances"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("All instances of %s deleted", modelName)})
}

// Get/Edit an instance of a model
func (a *Control) GetEditModelInstance(c *gin.Context) {
	modelName := c.Param("model")
	id := c.Param("id")
	model, exists := a.Models[modelName]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	instance := reflect.New(reflect.TypeOf(model)).Interface()

	// Fetch the instance
	if err := a.DB.First(instance, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}

	if c.Request.Method == http.MethodGet {
		c.JSON(http.StatusOK, instance)
		return
	}

	// Update instance (except ID)
	if err := c.ShouldBindJSON(instance); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := a.DB.Save(instance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance updated"})
}

// Delete an instance of a model
func (a *Control) DeleteModelInstance(c *gin.Context) {
	modelName := c.Param("model")
	id := c.Param("id")
	model, exists := a.Models[modelName]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	instance := reflect.New(reflect.TypeOf(model)).Interface()

	if err := a.DB.Delete(instance, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete instance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Instance %s deleted", id)})
}

func (a *Control) QueryModel(c *gin.Context) {
	modelName := c.Param("model")
	queryString := c.Param("query") // Example: "username=exampleUser&age=30|id=5"
	method := c.Request.Method      // GET, PUT, or DELETE

	var values []interface{}

	// Step 1: Handle IN condition (e.g., field{1,2,3})
	if strings.Contains(queryString, "{") && strings.Contains(queryString, "}") {
		queryString = strings.ReplaceAll(queryString, "{", " IN (")
		queryString = strings.ReplaceAll(queryString, "}", ")")
	}

	// Step 2: Handle LIKE condition (e.g., field~value)
	if strings.Contains(queryString, "~") {
		// Replace '~' with LIKE and wrap value with '%' for SQL LIKE pattern
		queryString = regexp.MustCompile(`(\w+)~([^\s_|&]+)`).ReplaceAllString(queryString, `$1 LIKE '%$2%'`)
	}

	// Step 3: Handle BETWEEN condition (e.g., field(10,20))
	if strings.Contains(queryString, "(") && strings.Contains(queryString, ")") {
		// Process the BETWEEN syntax
		regex := regexp.MustCompile(`(\w+)\(([^)]+)\)`)
		matches := regex.FindAllStringSubmatch(queryString, -1)
		for _, match := range matches {
			field := match[1]
			rangeValues := match[2]
			rangeParts := strings.Split(rangeValues, ",")
			if len(rangeParts) == 2 {
				// Replace the range with the correct BETWEEN syntax
				queryString = strings.Replace(queryString, match[0], fmt.Sprintf("%s BETWEEN ? AND ?", field), 1)
				values = append(values, rangeParts[0], rangeParts[1])
			}
		}
	}

	// Step 4: Handle equality condition (e.g., field=value)
	if strings.Contains(queryString, "=") {
		// Replace '=' with = "value" to handle string values
		queryString = regexp.MustCompile(`(\w+)=([^\s&|_]+)`).ReplaceAllString(queryString, `$1 = '$2'`)
	}
	fmt.Println("the quoted", queryString)
	// Step 5: Perform the logical operator replacements at the end
	queryString = "(" + queryString + ")"
	queryString = strings.ReplaceAll(queryString, "_&_", ") AND (")
	queryString = strings.ReplaceAll(queryString, "_|_", ") OR (")
	queryString = strings.ReplaceAll(queryString, "&", " AND ")
	queryString = strings.ReplaceAll(queryString, "|", " OR ")
	whereClause := " WHERE " + queryString

	fmt.Println("Generated SQL WHERE Clause:", whereClause)
	fmt.Println("Values:", values)

	// Initialize query variable
	var query string
	var err error

	switch method {
	case http.MethodGet:
		// Build SELECT query for GET method
		query = fmt.Sprintf("SELECT * FROM %s%s", modelName, whereClause)
		fmt.Println("Generated GET SQL:", query)

		// Execute query
		var results []map[string]interface{}
		err = a.DB.Raw(query, values...).Scan(&results).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute query", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)

	case http.MethodPut:
		// Get JSON data for updates
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Construct UPDATE query
		var updateClauses []string
		var updateValues []interface{}

		for field, value := range data {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = ?", field))
			updateValues = append(updateValues, value)
		}

		if len(updateClauses) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields provided for update"})
			return
		}

		// Generate SQL query with proper WHERE clause
		query = fmt.Sprintf("UPDATE %s SET %s  %s", modelName, strings.Join(updateClauses, ", "), whereClause)
		fmt.Println("Generated PUT SQL:", query)

		// Execute UPDATE query safely with bound parameters
		err = a.DB.Exec(query, append(updateValues, values...)...).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instances", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Instances updated"})

	case http.MethodDelete:
		// Prevent deleting all records accidentally

		query = fmt.Sprintf("DELETE FROM %s%s", modelName, whereClause)
		fmt.Println("Generated DELETE SQL:", query)

		// Execute DELETE query
		err = a.DB.Exec(query, values...).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete instances", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Instances deleted"})

	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
	}
}
