package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whynot00/imsi-bot/internal/service"
)

const maxUploadSize = 256 << 20 // 256 MB

// UploadHandler handles CSV file uploads for both parametr and rk kinds.
type UploadHandler struct {
	svc  *service.ImportService
	jobs *JobStore
}

func NewUploadHandler(svc *service.ImportService, jobs *JobStore) *UploadHandler {
	return &UploadHandler{svc: svc, jobs: jobs}
}

// UploadParametr godoc
// POST /upload/parametr  (multipart/form-data, field "file")
func (h *UploadHandler) UploadParametr(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

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
// POST /upload/rk  (multipart/form-data, field "file")
func (h *UploadHandler) UploadRK(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

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
// GET /upload/status/:id
func (h *UploadHandler) JobStatus(c *gin.Context) {
	jobStatus(c, h.jobs)
}

func (h *UploadHandler) processParametr(jobID string, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[upload] panic in processParametr job %s: %v", jobID, r)
			h.jobs.SetFailed(jobID, fmt.Sprintf("internal panic: %v", r))
		}
	}()

	h.jobs.SetProcessing(jobID)
	onProgress := func(p service.Progress) {
		h.jobs.UpdateProgress(jobID, p)
	}
	result, err := h.svc.ImportParametrFromCSV(context.Background(), data, onProgress)
	if err != nil {
		log.Printf("[upload] job %s failed: %v", jobID, err)
		h.jobs.SetFailed(jobID, err.Error())
		return
	}
	log.Printf("[upload] job %s done: %+v", jobID, result)
	h.jobs.SetDone(jobID, result)
}

func (h *UploadHandler) processRK(jobID string, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[upload] panic in processRK job %s: %v", jobID, r)
			h.jobs.SetFailed(jobID, fmt.Sprintf("internal panic: %v", r))
		}
	}()

	h.jobs.SetProcessing(jobID)
	onProgress := func(p service.Progress) {
		h.jobs.UpdateProgress(jobID, p)
	}
	result, err := h.svc.ImportRKFromCSV(context.Background(), data, onProgress)
	if err != nil {
		log.Printf("[upload] job %s failed: %v", jobID, err)
		h.jobs.SetFailed(jobID, err.Error())
		return
	}
	log.Printf("[upload] job %s done: %+v", jobID, result)
	h.jobs.SetDone(jobID, result)
}

// --- shared helpers ---

func readFileField(c *gin.Context) ([]byte, error) {
	f, _, err := c.Request.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("file is required")
	}
	defer f.Close()
	return io.ReadAll(f)
}

func jobStatus(c *gin.Context, jobs *JobStore) {
	id := c.Param("id")
	j := jobs.Get(id)
	if j == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, j)
}
