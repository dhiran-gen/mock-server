package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// Student represents a student model
type Student struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// StudentStore holds the in-memory student data
type StudentStore struct {
	mu       sync.RWMutex
	students []Student
	nextID   int
}

// Global student store
var store = &StudentStore{
	students: []Student{},
	nextID:   1,
}

// CreateStudent - POST /students
func CreateStudent(c *gin.Context) {
	var student Student

	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	student.ID = store.nextID
	store.nextID++
	store.students = append(store.students, student)

	c.JSON(http.StatusCreated, student)
}

// GetAllStudents - GET /students
func GetAllStudents(c *gin.Context) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	c.JSON(http.StatusOK, store.students)
}

// GetStudent - GET /students/:id
func GetStudent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	for _, student := range store.students {
		if student.ID == id {
			c.JSON(http.StatusOK, student)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
}

// UpdateStudent - PUT /students/:id
func UpdateStudent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	var updatedStudent Student
	if err := c.ShouldBindJSON(&updatedStudent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	for i, student := range store.students {
		if student.ID == id {
			store.students[i].Name = updatedStudent.Name
			store.students[i].Email = updatedStudent.Email
			store.students[i].Age = updatedStudent.Age
			c.JSON(http.StatusOK, store.students[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
}

// DeleteStudent - DELETE /students/:id
func DeleteStudent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	for i, student := range store.students {
		if student.ID == id {
			store.students = append(store.students[:i], store.students[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Student deleted successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
}

func main() {
	// Create a new Gin router
	router := gin.Default()

	// CRUD Routes for Students
	router.POST("/students", CreateStudent)
	router.GET("/students", GetAllStudents)
	router.GET("/students/:id", GetStudent)
	router.PUT("/students/:id", UpdateStudent)
	router.DELETE("/students/:id", DeleteStudent)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Start the server on port 8080
	router.Run(":8080")
}
