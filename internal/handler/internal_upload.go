package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/whynot00/imsi-bot/internal/service"
)

// InternalUploadHandler handles large file uploads via API key auth.
type InternalUploadHandler struct {
	svc  *service.ImportService
	jobs *JobStore
}

func NewInternalUploadHandler(svc *service.ImportService, jobs *JobStore) *InternalUploadHandler {
	return &InternalUploadHandler{svc: svc, jobs: jobs}
}

// UploadParametr godoc
// POST /internal/upload/parametr  (multipart/form-data, field "file")
func (h *InternalUploadHandler) UploadParametr(c *gin.Context) {
	data, err := readFileField(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := h.jobs.Create()
	go h.processParametr(job.ID, data)

	c.JSON(http.StatusAccepted, gin.H{"job_id": job.ID})
}

// UploadRK godoc
// POST /internal/upload/rk  (multipart/form-data, field "file")
func (h *InternalUploadHandler) UploadRK(c *gin.Context) {
	data, err := readFileField(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := h.jobs.Create()
	go h.processRK(job.ID, data)

	c.JSON(http.StatusAccepted, gin.H{"job_id": job.ID})
}

// JobStatus returns the current state of a background job.
// GET /internal/upload/status/:id
func (h *InternalUploadHandler) JobStatus(c *gin.Context) {
	jobStatus(c, h.jobs)
}

func (h *InternalUploadHandler) processParametr(jobID string, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[internal] panic in processParametr job %s: %v", jobID, r)
			h.jobs.SetFailed(jobID, fmt.Sprintf("internal panic: %v", r))
		}
	}()

	h.jobs.SetProcessing(jobID)
	result, err := h.svc.ImportParametrFromCSV(context.Background(), data)
	if err != nil {
		log.Printf("[internal] job %s failed: %v", jobID, err)
		h.jobs.SetFailed(jobID, err.Error())
		return
	}
	log.Printf("[internal] job %s done: %+v", jobID, result)
	h.jobs.SetDone(jobID, result)
}

func (h *InternalUploadHandler) processRK(jobID string, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[internal] panic in processRK job %s: %v", jobID, r)
			h.jobs.SetFailed(jobID, fmt.Sprintf("internal panic: %v", r))
		}
	}()

	h.jobs.SetProcessing(jobID)
	result, err := h.svc.ImportRKFromCSV(context.Background(), data)
	if err != nil {
		log.Printf("[internal] job %s failed: %v", jobID, err)
		h.jobs.SetFailed(jobID, err.Error())
		return
	}
	log.Printf("[internal] job %s done: %+v", jobID, result)
	h.jobs.SetDone(jobID, result)
}
