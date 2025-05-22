// handlers/person.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"effective-mobile/config"
	"effective-mobile/db"
	"effective-mobile/models"
	"effective-mobile/services"
)

// @Summary Get all people
// @Description Get a list of all people with optional filters and pagination
// @Tags people
// @Accept json
// @Produce json
// @Param name query string false "Filter by name"
// @Param surname query string false "Filter by surname"
// @Param patronymic query string false "Filter by patronymic"
// @Param age_min query int false "Filter by minimum age"
// @Param age_max query int false "Filter by maximum age"
// @Param gender query string false "Filter by gender"
// @Param nationality query string false "Filter by nationality"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {array} models.Person
// @Failure 500 {object} map[string]string
// @Router /people [get]
func GetPeople(c *gin.Context) {
	config.Log.Info("Handling GetPeople request")

	var people []models.Person
	query := db.DB

	// Фильтры
	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
		config.Log.Debugf("Filtering by name: %s", name)
	}
	if surname := c.Query("surname"); surname != "" {
		query = query.Where("surname ILIKE ?", "%"+surname+"%")
		config.Log.Debugf("Filtering by surname: %s", surname)
	}
	if patronymic := c.Query("patronymic"); patronymic != "" {
		query = query.Where("patronymic ILIKE ?", "%"+patronymic+"%")
		config.Log.Debugf("Filtering by patronymic: %s", patronymic)
	}
	if ageMinStr := c.Query("age_min"); ageMinStr != "" {
		if ageMin, err := strconv.Atoi(ageMinStr); err == nil {
			query = query.Where("age >= ?", ageMin)
			config.Log.Debugf("Filtering by min age: %d", ageMin)
		}
	}
	if ageMaxStr := c.Query("age_max"); ageMaxStr != "" {
		if ageMax, err := strconv.Atoi(ageMaxStr); err == nil {
			query = query.Where("age <= ?", ageMax)
			config.Log.Debugf("Filtering by max age: %d", ageMax)
		}
	}
	if gender := c.Query("gender"); gender != "" {
		query = query.Where("gender ILIKE ?", "%"+gender+"%")
		config.Log.Debugf("Filtering by gender: %s", gender)
	}
	if nationality := c.Query("nationality"); nationality != "" {
		query = query.Where("nationality ILIKE ?", "%"+nationality+"%")
		config.Log.Debugf("Filtering by nationality: %s", nationality)
	}

	// Пагинация
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit
	config.Log.Debugf("Applying pagination: page=%d, limit=%d, offset=%d", page, limit, offset)

	if err := query.Limit(limit).Offset(offset).Find(&people).Error; err != nil {
		config.Log.Errorf("Error getting people: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve people"})
		return
	}

	c.JSON(http.StatusOK, people)
	config.Log.Info("Successfully retrieved people")
}

// @Summary Create a new person
// @Description Add a new person to the database, enriching their data with age, gender, and nationality
// @Tags people
// @Accept json
// @Produce json
// @Param person body models.PersonInput true "Person object to be created"
// @Success 201 {object} models.Person
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people [post]
func CreatePerson(c *gin.Context) {
	config.Log.Info("Handling CreatePerson request")

	var inputPerson models.PersonInput
	if err := c.ShouldBindJSON(&inputPerson); err != nil {
		config.Log.Warnf("Invalid input for CreatePerson: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	person := models.Person{
		Name:       inputPerson.Name,
		Surname:    inputPerson.Surname,
		Patronymic: inputPerson.Patronymic,
	}

	// Обогащаем данные
	if err := services.EnrichPerson(&person); err != nil {
		config.Log.Errorf("Error enriching person data: %v", err)
		// В зависимости от требований, можно вернуть ошибку или сохранить без обогащения
		// Для данного ТЗ, лучше продолжить и сохранить то, что есть, логгируя ошибку обогащения.
	}

	if err := db.DB.Create(&person).Error; err != nil {
		config.Log.Errorf("Error creating person in DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create person"})
		return
	}

	config.Log.Infof("Successfully created person with ID: %d", person.ID)
	c.JSON(http.StatusCreated, person)
}

// @Summary Get a person by ID
// @Description Get a single person by their ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Success 200 {object} models.Person
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people/{id} [get]
func GetPersonByID(c *gin.Context) {
	config.Log.Info("Handling GetPersonByID request")
	id := c.Param("id")
	var person models.Person
	if err := db.DB.First(&person, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			config.Log.Warnf("Person with ID %s not found", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}
		config.Log.Errorf("Error getting person by ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve person"})
		return
	}
	config.Log.Infof("Successfully retrieved person with ID: %s", id)
	c.JSON(http.StatusOK, person)
}

// @Summary Update an existing person
// @Description Update a person's details by ID. Only provided fields will be updated.
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Param person body models.PersonInput true "Updated person object"
// @Success 200 {object} models.Person
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people/{id} [put]
func UpdatePerson(c *gin.Context) {
	config.Log.Info("Handling UpdatePerson request")
	id := c.Param("id")
	var existingPerson models.Person
	if err := db.DB.First(&existingPerson, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			config.Log.Warnf("Person with ID %s not found for update", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}
		config.Log.Errorf("Error finding person for update with ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve person for update"})
		return
	}

	var updateInput models.PersonInput
	if err := c.ShouldBindJSON(&updateInput); err != nil {
		config.Log.Warnf("Invalid input for UpdatePerson with ID %s: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем только те поля, которые пришли в запросе
	if updateInput.Name != "" {
		existingPerson.Name = updateInput.Name
	}
	if updateInput.Surname != "" {
		existingPerson.Surname = updateInput.Surname
	}
	if updateInput.Patronymic != nil {
		existingPerson.Patronymic = updateInput.Patronymic
	}

	// Если имя изменилось, то нужно переобогатить данные
	// Для простоты, переобогащаем всегда при обновлении
	config.Log.Debugf("Re-enriching person with ID %s after update attempt", id)
	if err := services.EnrichPerson(&existingPerson); err != nil {
		config.Log.Errorf("Error re-enriching person data for ID %s: %v", id, err)
	}

	if err := db.DB.Save(&existingPerson).Error; err != nil {
		config.Log.Errorf("Error updating person with ID %s in DB: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update person"})
		return
	}

	config.Log.Infof("Successfully updated person with ID: %s", id)
	c.JSON(http.StatusOK, existingPerson)
}

// @Summary Delete a person by ID
// @Description Delete a person from the database by their ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people/{id} [delete]
func DeletePerson(c *gin.Context) {
	config.Log.Info("Handling DeletePerson request")
	id := c.Param("id")
	var person models.Person
	if err := db.DB.First(&person, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			config.Log.Warnf("Person with ID %s not found for deletion", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}
		config.Log.Errorf("Error finding person for deletion with ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve person for deletion"})
		return
	}

	if err := db.DB.Delete(&person).Error; err != nil {
		config.Log.Errorf("Error deleting person with ID %s from DB: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete person"})
		return
	}

	config.Log.Infof("Successfully deleted person with ID: %s", id)
	c.Status(http.StatusNoContent)
}

// Добавим структуру для входных данных, как в ТЗ
type PersonInput struct {
	Name       string  `json:"name" binding:"required"`
	Surname    string  `json:"surname" binding:"required"`
	Patronymic *string `json:"patronymic,omitempty"`
}