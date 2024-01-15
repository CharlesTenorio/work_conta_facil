package main

import (
	"log"
	"os"

	"github.com/katana/worker/orcafacil-go/internal/config"
	"github.com/katana/worker/orcafacil-go/pkg/adapter/mongodb"
	"github.com/katana/worker/orcafacil-go/pkg/adapter/rabbitmq"
	"github.com/katana/worker/orcafacil-go/pkg/service"
)

var (
	VERSION = "0.1.0-dev"
	COMMIT  = "ABCDEFG-dev"
)

func main() {
	os.Setenv("SRV_RMQ_URI", "amqp://admin:supersenha@localhost:5672/")
	filas := []rabbitmq.Fila{
		{
			Name:       "QUEUE_PRDS_PARA_COTACAO",
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
		},
		{
			Name:       "QUEUE_PRDS_PARA_COTACAO_DLQ",
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
		},
	}

	conf := config.NewConfig()
	db_conn := mongodb.New(conf)
	rbmq_conn := rabbitmq.NewRabbitMQ(filas, conf)
	task_service := service.NewTaskService(rbmq_conn, db_conn)

	done := make(chan bool)
	go task_service.Run()
	log.Printf("Worker Running [Mode: %s], [Version: %s], [Commit: %s]", conf.Mode, VERSION, COMMIT)
	<-done
}
