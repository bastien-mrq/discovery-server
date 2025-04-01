package discoveryserver

import (
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var microservices = make(map[string]map[string]Instance)

func main() {
	r := gin.Default()

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/list", handleListMicoservices)
	r.POST("/register", handleRegisterMicroservice)
	r.POST("/unregister", handleUnregisterInstance)

	r.Run("127.0.0.1:1111")
}

func handleRegisterMicroservice(ctx *gin.Context) {
	var input struct {
		Name string `json:"name"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if microservices[input.Name] == nil {
		microservices[input.Name] = make(map[string]Instance)
	}
	uuid := uuid.New().String()

	port, err := GetFreePort()
	if err != nil {
		log.Fatalf("Unable to find a free port: %v", err)
	}
	microservices[input.Name][uuid] = Instance{Port: port}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Microservice registered",
		"uuid":    uuid,
		"port":    port,
	})
}

func handleUnregisterInstance(ctx *gin.Context) {
	var input struct {
		Uuid string `json:"uuid"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	found := false
	for serviceName, instances := range microservices {
		if _, exists := instances[input.Uuid]; exists {
			delete(instances, input.Uuid)
			found = true

			if len(instances) == 0 {
				delete(microservices, serviceName)
			}

			ctx.JSON(http.StatusOK, gin.H{
				"message":     "Instance deregistered",
				"uuid":        input.Uuid,
				"serviceName": serviceName,
			})
			return
		}
	}

	if !found {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Instance with given UUID not found",
		})
	}
}

func handleListMicoservices(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"microservices": microservices,
	})
}

type Instance struct {
	Port int `json:"port"`
}

func GetFreePort() (int, error) {
	l, err := net.Listen("tcp", ":0") // Ask OS to assign a free port
	if err != nil {
		return 0, err
	}
	defer l.Close()

	addr := l.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
