package handlers

import (
	"mcpdocs/schema"
	"mcpdocs/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService services.ProjectServiceInterface
}

func NewProjectHandler(projectService services.ProjectServiceInterface) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID := c.GetString("userID") // Assuming userID is set in context by auth middleware
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req schema.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) GetProjects(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projects, err := h.projectService.GetProjects(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.projectService.GetProject(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Check if the project belongs to the user
	userID := c.GetString("userID")
	if project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req schema.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check ownership before update
	existingProject, err := h.projectService.GetProject(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	userID := c.GetString("userID")
	if existingProject.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	project, err := h.projectService.UpdateProject(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check ownership before delete
	existingProject, err := h.projectService.GetProject(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	userID := c.GetString("userID")
	if existingProject.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if err := h.projectService.DeleteProject(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}
