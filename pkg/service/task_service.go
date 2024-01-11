package service

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/katana/worker/orcafacil-go/internal/dto"
	"github.com/katana/worker/orcafacil-go/pkg/adapter/mongodb"
	"github.com/katana/worker/orcafacil-go/pkg/adapter/rabbitmq"
	"github.com/katana/worker/orcafacil-go/pkg/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

const DATE_FORMAT = "2006-01-02 15:04:05"

type taskService struct {
	rbmq_conn rabbitmq.RabbitInterface
	db_conn   mongodb.MongoDBInterface
}

func NewTaskService(rbmq_conn rabbitmq.RabbitInterface, db_conn mongodb.MongoDBInterface) *taskService {
	return &taskService{
		rbmq_conn: rbmq_conn,
		db_conn:   db_conn,
	}
}
func (tmcs *taskService) ReadFromQueue(queue_name string) (<-chan dto.ProdutoEnviadoParaFilaDeOrcamentoDTO, <-chan error) {
	output := make(chan dto.ProdutoEnviadoParaFilaDeOrcamentoDTO)
	errors := make(chan error)

	go func() {
		tmcs.rbmq_conn.Consumer(queue_name, func(msg *amqp.Delivery) {
			var produtoMsg dto.ProdutoEnviadoParaFilaDeOrcamentoDTO

			// Decodificar a mensagem JSON
			err := json.Unmarshal(msg.Body, &produtoMsg)
			if err != nil {
				log.Println("Erro ao decodificar a mensagem JSON:", err)
				errors <- err
				return
			}

			// Enviar a struct para o canal de saída
			output <- produtoMsg
		})
	}()

	return output, errors
}

func (tmcs *taskService) consumerCallback(msg *amqp.Delivery) {

	log.Println("New MSG received")

	mc := &model.MessageConversator{}
	err := json.Unmarshal(msg.Body, mc)
	if err != nil {
		log.Println("Failed to read a consumer msg")
		log.Println(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mc.CreatedAt = time.Now().Format(time.RFC3339)

	err = tmcs.ReceiveMessage(ctx, mc)
	if err != nil {
		log.Println("Failed to save msg")
		log.Println(err)

		dlq_msg := &rabbitmq.Message{
			Data:        msg.Body,
			ContentType: msg.ContentType,
		}

		err := tmcs.rbmq_conn.Publish("FINANCIAL_MESSAGE_COUNTER_DLQ", dlq_msg)
		if err != nil {
			log.Println("Failed to save msg in DLQ")
			log.Println(err)
			log.Println(string(dlq_msg.Data))
		}
	}

	if err := msg.Ack(false); err != nil {
		log.Println("Erro to ACK MSG Status")
	} else {
		log.Println("MSG Status update success")
	}
}

func (tmcs *task_message_counter_service) Run() {

	err := tmcs.rbmq_conn.Connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	tmcs.rbmq_conn.Start("FINANCIAL_MESSAGE_COUNTER", tmcs.consumerCallback)
}
