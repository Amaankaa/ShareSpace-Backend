# Health Check Endpoint Implementation

To complete the CI/CD setup, you need to add a health check endpoint to your application.

## Add to your router (in Delivery/routers/router.go):

```go
// Add this to your router setup
func SetupRouter(controller *controllers.Controller) *gin.Engine {
    router := gin.Default()
    
    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "healthy",
            "timestamp": time.Now().UTC(),
            "version": "1.0.0",
            "service": "sharespace-backend",
        })
    })
    
    // Your existing routes...
    api := router.Group("/api")
    {
        // ... existing routes
    }
    
    return router
}
```

## Or create a dedicated health controller:

```go
// Delivery/controllers/health_controller.go
package controllers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
    return &HealthController{}
}

func (hc *HealthController) HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "healthy",
        "timestamp": time.Now().UTC(),
        "version":   "1.0.0",
        "service":   "sharespace-backend",
    })
}

func (hc *HealthController) ReadinessCheck(c *gin.Context) {
    // Add database connectivity check here
    // Add other service dependency checks
    
    c.JSON(http.StatusOK, gin.H{
        "status":    "ready",
        "timestamp": time.Now().UTC(),
        "checks": gin.H{
            "database": "connected",
            "redis":    "connected",
        },
    })
}
```

This endpoint will be used by:
- Docker health checks
- Load balancer health checks  
- CI/CD pipeline verification
- Monitoring systems
